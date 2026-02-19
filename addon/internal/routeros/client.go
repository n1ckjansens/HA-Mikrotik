package routeros

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/micro-ha/mikrotik-presence/addon/internal/model"
)

const (
	defaultTimeout   = 10 * time.Second
	maxRetryAttempts = 3
)

type HTTPStatusError struct {
	StatusCode int
	Body       string
}

func (e *HTTPStatusError) Error() string {
	return fmt.Sprintf("status %d: %s", e.StatusCode, e.Body)
}

type Client struct {
	httpClient *http.Client
	logger     *slog.Logger
}

func NewClient() *Client {
	return NewClientWithHTTPClient(&http.Client{Timeout: defaultTimeout})
}

func NewClientWithHTTPClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultTimeout}
	}
	if httpClient.Timeout == 0 {
		httpClient.Timeout = defaultTimeout
	}
	return &Client{httpClient: httpClient}
}

func (c *Client) WithLogger(logger *slog.Logger) *Client {
	c.logger = logger
	return c
}

func (c *Client) logRouterRequest(operation string, method string, endpoint string) {
	if c == nil || c.logger == nil {
		return
	}
	c.logger.Info("routeros request", "operation", operation, "method", method, "endpoint", endpoint)
}

func (c *Client) FetchSnapshot(ctx context.Context, cfg model.RouterConfig) (*Snapshot, error) {
	client := *c.httpClient
	if cfg.SSL {
		var transport *http.Transport
		if existing, ok := client.Transport.(*http.Transport); ok {
			transport = existing.Clone()
		} else if defaultTransport, ok := http.DefaultTransport.(*http.Transport); ok {
			transport = defaultTransport.Clone()
		} else {
			transport = &http.Transport{}
		}
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: !cfg.VerifyTLS} //nolint:gosec
		client.Transport = transport
	}

	base := strings.TrimSuffix(cfg.BaseURL(), "/")
	snapshot := &Snapshot{FetchedAt: time.Now().UTC()}

	var err error
	if snapshot.DHCP, base, err = fetchDHCP(ctx, &client, base, cfg); err != nil {
		return nil, err
	}
	if snapshot.WiFi, err = fetchWiFi(ctx, &client, base, cfg); err != nil {
		return nil, err
	}
	if snapshot.Bridge, err = fetchBridge(ctx, &client, base, cfg); err != nil {
		return nil, err
	}
	if snapshot.ARP, err = fetchARP(ctx, &client, base, cfg); err != nil {
		return nil, err
	}
	if snapshot.Addresses, err = fetchIPAddresses(ctx, &client, base, cfg); err != nil {
		return nil, err
	}
	return snapshot, nil
}

func fetchDHCP(
	ctx context.Context,
	client *http.Client,
	base string,
	cfg model.RouterConfig,
) ([]DHCPLease, string, error) {
	rows, nextBase, err := fetchRowsWithHTTPSFallback(ctx, client, base, "/ip/dhcp-server/lease", cfg)
	if err != nil {
		return nil, base, err
	}
	items := make([]DHCPLease, 0, len(rows))
	for _, row := range rows {
		mac := canonicalMAC(str(row["mac-address"]))
		if mac == "" {
			continue
		}
		items = append(items, DHCPLease{
			MAC:      mac,
			Address:  str(row["address"]),
			HostName: str(row["host-name"]),
			Server:   str(row["server"]),
			Status:   str(row["status"]),
			LastSeen: str(row["last-seen"]),
			Dynamic:  boolValue(row["dynamic"]),
			Blocked:  boolValue(row["blocked"]),
			Disabled: boolValue(row["disabled"]),
		})
	}
	return items, nextBase, nil
}

func fetchWiFi(ctx context.Context, client *http.Client, base string, cfg model.RouterConfig) ([]WiFiRegistration, error) {
	paths := []struct {
		endpoint string
		driver   string
	}{
		{endpoint: "/interface/wifi/registration-table", driver: "wifi"},
		{endpoint: "/interface/wifiwave2/registration-table", driver: "wifiwave2"},
		{endpoint: "/interface/wireless/registration-table", driver: "wireless"},
	}

	items := make([]WiFiRegistration, 0)
	for _, path := range paths {
		rows, err := fetchRows(ctx, client, base+path.endpoint, cfg)
		if err != nil {
			if isMissingEndpointError(err) {
				continue
			}
			return nil, err
		}
		for _, row := range rows {
			mac := canonicalMAC(str(row["mac-address"]))
			if mac == "" {
				continue
			}
			items = append(items, WiFiRegistration{
				MAC:          mac,
				Interface:    str(row["interface"]),
				SSID:         str(row["ssid"]),
				Uptime:       str(row["uptime"]),
				LastActivity: str(row["last-activity"]),
				Signal:       firstNonEmpty(str(row["signal"]), str(row["tx-signal"])),
				AuthType:     firstNonEmpty(str(row["auth-type"]), str(row["authentication-types"])),
				Band:         str(row["band"]),
				Driver:       path.driver,
			})
		}
	}
	return items, nil
}

func fetchBridge(ctx context.Context, client *http.Client, base string, cfg model.RouterConfig) ([]BridgeHost, error) {
	rows, err := fetchRows(ctx, client, base+"/interface/bridge/host", cfg)
	if err != nil {
		return nil, err
	}
	items := make([]BridgeHost, 0, len(rows))
	for _, row := range rows {
		mac := canonicalMAC(str(row["mac-address"]))
		if mac == "" {
			continue
		}
		items = append(items, BridgeHost{
			MAC:       mac,
			Bridge:    str(row["bridge"]),
			Interface: firstNonEmpty(str(row["interface"]), str(row["on-interface"])),
			VID:       str(row["vid"]),
		})
	}
	return items, nil
}

func fetchARP(ctx context.Context, client *http.Client, base string, cfg model.RouterConfig) ([]ARPEntry, error) {
	rows, err := fetchRows(ctx, client, base+"/ip/arp", cfg)
	if err != nil {
		return nil, err
	}
	items := make([]ARPEntry, 0, len(rows))
	for _, row := range rows {
		mac := canonicalMAC(str(row["mac-address"]))
		if mac == "" {
			continue
		}
		items = append(items, ARPEntry{
			MAC:       mac,
			Address:   str(row["address"]),
			Interface: str(row["interface"]),
			Flags:     str(row["flags"]),
		})
	}
	return items, nil
}

func fetchIPAddresses(ctx context.Context, client *http.Client, base string, cfg model.RouterConfig) ([]IPAddress, error) {
	rows, err := fetchRows(ctx, client, base+"/ip/address", cfg)
	if err != nil {
		return nil, err
	}
	items := make([]IPAddress, 0, len(rows))
	for _, row := range rows {
		address := str(row["address"])
		if address == "" {
			continue
		}
		items = append(items, IPAddress{Address: address, Interface: str(row["interface"])})
	}
	return items, nil
}

func fetchRows(ctx context.Context, client *http.Client, endpoint string, cfg model.RouterConfig) ([]map[string]any, error) {
	var lastErr error
	for attempt := 1; attempt <= maxRetryAttempts; attempt++ {
		rows, err := doFetchRows(ctx, client, endpoint, cfg)
		if err == nil {
			return rows, nil
		}
		if isMissingEndpointError(err) {
			return nil, fmt.Errorf("routeros GET %s failed: %w", endpoint, err)
		}
		lastErr = fmt.Errorf("routeros GET %s failed: %w", endpoint, err)
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("routeros GET %s canceled: %w", endpoint, ctx.Err())
		case <-time.After(time.Duration(attempt) * 400 * time.Millisecond):
		}
	}
	return nil, fmt.Errorf(
		"routeros GET %s failed after %d attempts: %w",
		endpoint,
		maxRetryAttempts,
		lastErr,
	)
}

func fetchRowsWithHTTPSFallback(
	ctx context.Context,
	client *http.Client,
	base string,
	path string,
	cfg model.RouterConfig,
) ([]map[string]any, string, error) {
	endpoint := base + path
	rows, err := fetchRows(ctx, client, endpoint, cfg)
	if err == nil {
		return rows, base, nil
	}
	if !shouldFallbackToHTTP(cfg, endpoint, err) {
		return nil, base, err
	}

	fallbackBase := downgradeBaseToHTTP(base)
	fallbackRows, fallbackErr := fetchRows(ctx, client, fallbackBase+path, cfg)
	if fallbackErr != nil {
		return nil, base, fmt.Errorf("%w; http fallback failed: %v", err, fallbackErr)
	}
	return fallbackRows, fallbackBase, nil
}

func doFetchRows(ctx context.Context, client *http.Client, endpoint string, cfg model.RouterConfig) ([]map[string]any, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(cfg.Username, cfg.Password)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		return nil, &HTTPStatusError{StatusCode: resp.StatusCode, Body: string(body)}
	}

	var rows []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return nil, err
	}
	if rows == nil {
		return nil, errors.New("empty response")
	}
	return rows, nil
}

func str(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case fmt.Stringer:
		return t.String()
	default:
		if v == nil {
			return ""
		}
		return fmt.Sprintf("%v", v)
	}
}

func canonicalMAC(v string) string {
	v = strings.TrimSpace(strings.ToUpper(v))
	if v == "" {
		return ""
	}
	return strings.ReplaceAll(v, "-", ":")
}

func isMissingEndpointError(err error) bool {
	var statusErr *HTTPStatusError
	if !errors.As(err, &statusErr) {
		return false
	}
	return statusErr.StatusCode == http.StatusNotFound || statusErr.StatusCode == http.StatusBadRequest
}

func shouldFallbackToHTTP(cfg model.RouterConfig, endpoint string, err error) bool {
	if !cfg.SSL || !strings.HasPrefix(endpoint, "https://") {
		return false
	}
	if errors.Is(err, context.Canceled) {
		return false
	}
	if isMissingEndpointError(err) {
		return false
	}
	return isLikelyHTTPSServiceMismatch(err)
}

func isLikelyHTTPSServiceMismatch(err error) bool {
	var urlErr *url.Error
	if errors.As(err, &urlErr) && urlErr.Err != nil {
		err = urlErr.Err
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	var opErr *net.OpError
	if errors.As(err, &opErr) {
		lower := strings.ToLower(opErr.Err.Error())
		if strings.Contains(lower, "refused") {
			return true
		}
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "eof") ||
		strings.Contains(lower, "tls:") ||
		strings.Contains(lower, "https client") ||
		strings.Contains(lower, "first record does not look like a tls handshake")
}

func downgradeBaseToHTTP(base string) string {
	if strings.HasPrefix(base, "https://") {
		return "http://" + strings.TrimPrefix(base, "https://")
	}
	return base
}

func boolValue(v any) bool {
	switch typed := v.(type) {
	case bool:
		return typed
	case string:
		parsed, err := strconv.ParseBool(strings.TrimSpace(typed))
		return err == nil && parsed
	default:
		return false
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
