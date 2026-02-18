package configsync

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/micro-ha/mikrotik-presence/addon/internal/model"
)

type FetchResult struct {
	Configured bool
	Config     model.RouterConfig
}

type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

func NewClient(baseURL, token string) *Client {
	baseURL = strings.TrimSuffix(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = "http://supervisor/core"
	}
	return &Client{
		baseURL: baseURL,
		token:   strings.TrimSpace(token),
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

type configResponse struct {
	Configured      bool      `json:"configured"`
	Version         int64     `json:"version"`
	UpdatedAt       time.Time `json:"updated_at"`
	Host            string    `json:"host"`
	Username        string    `json:"username"`
	Password        string    `json:"password"`
	SSL             bool      `json:"ssl"`
	VerifyTLS       bool      `json:"verify_tls"`
	PollIntervalSec int       `json:"poll_interval_sec"`
	Roles           []string  `json:"roles"`
}

func (c *Client) FetchConfig(ctx context.Context) (FetchResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/mikrotik_presence/config", nil)
	if err != nil {
		return FetchResult{}, err
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return FetchResult{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return FetchResult{Configured: false}, nil
	}
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		return FetchResult{}, fmt.Errorf("config fetch status %d: %s", resp.StatusCode, string(body))
	}

	var payload configResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return FetchResult{}, err
	}
	if !payload.Configured {
		return FetchResult{Configured: false}, nil
	}

	cfg := model.RouterConfig{
		Version:         payload.Version,
		UpdatedAt:       payload.UpdatedAt.UTC(),
		Host:            payload.Host,
		Username:        payload.Username,
		Password:        payload.Password,
		SSL:             payload.SSL,
		VerifyTLS:       payload.VerifyTLS,
		PollIntervalSec: payload.PollIntervalSec,
		Roles:           payload.Roles,
	}
	if cfg.PollIntervalSec < 5 {
		cfg.PollIntervalSec = 5
	}
	return FetchResult{Configured: true, Config: cfg}, nil
}
