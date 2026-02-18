package routeros

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/micro-ha/mikrotik-presence/addon/internal/model"
)

const (
	defaultTimeout   = 10 * time.Second
	maxRetryAttempts = 3
)

type Client struct {
	httpClient *http.Client
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
	if snapshot.DHCP, err = fetchDHCP(ctx, &client, base, cfg); err != nil {
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

func fetchDHCP(ctx context.Context, client *http.Client, base string, cfg model.RouterConfig) ([]DHCPLease, error) {
	rows, err := fetchRows(ctx, client, base+"/ip/dhcp-server/lease", cfg)
	if err != nil {
		return nil, err
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
			Status:   str(row["status"]),
			LastSeen: str(row["last-seen"]),
		})
	}
	return items, nil
}

func fetchWiFi(ctx context.Context, client *http.Client, base string, cfg model.RouterConfig) ([]WiFiRegistration, error) {
	rows, err := fetchRows(ctx, client, base+"/interface/wifi/registration-table", cfg)
	if err != nil {
		return nil, err
	}
	items := make([]WiFiRegistration, 0, len(rows))
	for _, row := range rows {
		mac := canonicalMAC(str(row["mac-address"]))
		if mac == "" {
			continue
		}
		items = append(items, WiFiRegistration{
			MAC:          mac,
			Interface:    str(row["interface"]),
			Uptime:       str(row["uptime"]),
			LastActivity: str(row["last-activity"]),
		})
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
		items = append(items, BridgeHost{MAC: mac, Interface: str(row["on-interface"])})
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
		items = append(items, ARPEntry{MAC: mac, Address: str(row["address"]), Interface: str(row["interface"])})
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
		lastErr = err
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Duration(attempt) * 400 * time.Millisecond):
		}
	}
	return nil, fmt.Errorf("routeros request failed for %s: %w", endpoint, lastErr)
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
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
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
