package routeros

import (
	"bytes"
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

func (c *Client) AddAddressListEntry(
	ctx context.Context,
	cfg model.RouterConfig,
	list string,
	address string,
) error {
	list = strings.TrimSpace(list)
	address = strings.TrimSpace(address)
	if list == "" {
		return fmt.Errorf("address-list name is required")
	}
	if address == "" {
		return fmt.Errorf("address value is required")
	}

	httpClient := c.httpClientForConfig(cfg)
	endpoint := strings.TrimSuffix(cfg.BaseURL(), "/") + "/ip/firewall/address-list"
	payload := map[string]string{
		"list":    list,
		"address": address,
	}

	c.logRouterRequest("address-list.add", http.MethodPut, endpoint)
	err := executeRouterRequestWithRetry(ctx, httpClient, http.MethodPut, endpoint, cfg, payload)
	if err != nil && isDuplicateAddressListEntryError(err) {
		return nil
	}
	return err
}

func (c *Client) httpClientForConfig(cfg model.RouterConfig) *http.Client {
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
	if client.Timeout <= 0 {
		client.Timeout = defaultTimeout
	}
	return &client
}

func executeRouterRequestWithRetry(
	ctx context.Context,
	client *http.Client,
	method string,
	endpoint string,
	cfg model.RouterConfig,
	payload any,
) error {
	var lastErr error
	for attempt := 1; attempt <= maxRetryAttempts; attempt++ {
		err := doRouterRequest(ctx, client, method, endpoint, cfg, payload)
		if err == nil {
			return nil
		}
		if isNonRetriableHTTPStatus(err) {
			return fmt.Errorf("routeros %s %s failed: %w", method, endpoint, err)
		}
		lastErr = err
		select {
		case <-ctx.Done():
			return fmt.Errorf("routeros %s %s canceled: %w", method, endpoint, ctx.Err())
		case <-time.After(time.Duration(attempt) * 400 * time.Millisecond):
		}
	}
	return fmt.Errorf(
		"routeros %s %s failed after %d attempts: %w",
		method,
		endpoint,
		maxRetryAttempts,
		lastErr,
	)
}

func doRouterRequest(
	ctx context.Context,
	client *http.Client,
	method string,
	endpoint string,
	cfg model.RouterConfig,
	payload any,
) error {
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(encoded)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return err
	}
	req.SetBasicAuth(cfg.Username, cfg.Password)
	req.Header.Set("Accept", "application/json")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return &HTTPStatusError{StatusCode: resp.StatusCode, Body: string(body)}
	}
	return nil
}

func isNonRetriableHTTPStatus(err error) bool {
	var statusErr *HTTPStatusError
	if !errors.As(err, &statusErr) {
		return false
	}
	return statusErr.StatusCode >= 400 && statusErr.StatusCode < 500
}

func isDuplicateAddressListEntryError(err error) bool {
	var statusErr *HTTPStatusError
	if !errors.As(err, &statusErr) {
		return false
	}
	if statusErr.StatusCode != http.StatusBadRequest && statusErr.StatusCode != http.StatusConflict {
		return false
	}
	message := strings.ToLower(statusErr.Body)
	return strings.Contains(message, "already have") || strings.Contains(message, "already exists")
}
