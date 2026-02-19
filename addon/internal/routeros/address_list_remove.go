package routeros

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode"

	"github.com/micro-ha/mikrotik-presence/addon/internal/model"
)

func (c *Client) RemoveAddressListEntry(
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
	base := strings.TrimSuffix(cfg.BaseURL(), "/")
	ids, err := c.findAddressListEntryIDs(ctx, httpClient, cfg, base, list, address)
	if err != nil {
		return err
	}

	for _, id := range ids {
		endpoint := base + "/ip/firewall/address-list/" + id
		c.logRouterRequest("address-list.remove", http.MethodDelete, endpoint)
		if err := executeRouterRequestWithRetry(ctx, httpClient, http.MethodDelete, endpoint, cfg, nil); err != nil {
			var statusErr *HTTPStatusError
			if errors.As(err, &statusErr) && statusErr.StatusCode == http.StatusNotFound {
				continue
			}
			return err
		}
	}
	return nil
}

func (c *Client) findAddressListEntryIDs(
	ctx context.Context,
	httpClient *http.Client,
	cfg model.RouterConfig,
	base string,
	list string,
	address string,
) ([]string, error) {
	endpoint := base + "/ip/firewall/address-list/print"
	payload := map[string]any{
		".proplist": []string{".id", "list", "address"},
		".query":    []string{"list=" + list},
	}

	c.logRouterRequest("address-list.lookup", http.MethodPost, endpoint)
	rows, err := fetchRowsByPost(ctx, httpClient, endpoint, cfg, payload)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0)
	for _, row := range rows {
		if !addressListEntryMatches(row, list, address) {
			continue
		}
		id := normalizeRouterObjectID(str(row[".id"]))
		if id != "" {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func fetchRowsByPost(
	ctx context.Context,
	client *http.Client,
	endpoint string,
	cfg model.RouterConfig,
	payload any,
) ([]map[string]any, error) {
	var lastErr error
	for attempt := 1; attempt <= maxRetryAttempts; attempt++ {
		rows, err := doFetchRowsByPost(ctx, client, endpoint, cfg, payload)
		if err == nil {
			return rows, nil
		}
		if isMissingEndpointError(err) {
			return nil, fmt.Errorf("routeros POST %s failed: %w", endpoint, err)
		}
		lastErr = fmt.Errorf("routeros POST %s failed: %w", endpoint, err)
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("routeros POST %s canceled: %w", endpoint, ctx.Err())
		case <-time.After(time.Duration(attempt) * 400 * time.Millisecond):
		}
	}
	return nil, fmt.Errorf(
		"routeros POST %s failed after %d attempts: %w",
		endpoint,
		maxRetryAttempts,
		lastErr,
	)
}

func doFetchRowsByPost(
	ctx context.Context,
	client *http.Client,
	endpoint string,
	cfg model.RouterConfig,
	payload any,
) ([]map[string]any, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(encoded))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(cfg.Username, cfg.Password)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
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

func normalizeRouterObjectID(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	cleaned := strings.Map(func(r rune) rune {
		if unicode.Is(unicode.Cf, r) || unicode.IsControl(r) {
			return -1
		}
		return r
	}, value)
	cleaned = strings.TrimSpace(cleaned)
	if cleaned == "" {
		return ""
	}
	return cleaned
}

func addressListEntryMatches(row map[string]any, list string, address string) bool {
	rowList := strings.TrimSpace(str(row["list"]))
	if rowList != list {
		return false
	}
	rowAddress := strings.TrimSpace(str(row["address"]))
	return equalAddressListTarget(rowAddress, address)
}

func equalAddressListTarget(actual string, expected string) bool {
	actual = strings.TrimSpace(actual)
	expected = strings.TrimSpace(expected)
	if actual == expected {
		return true
	}
	if strings.HasSuffix(actual, "/32") {
		return strings.TrimSuffix(actual, "/32") == expected
	}
	if strings.HasSuffix(expected, "/32") {
		return strings.TrimSuffix(expected, "/32") == actual
	}
	return false
}
