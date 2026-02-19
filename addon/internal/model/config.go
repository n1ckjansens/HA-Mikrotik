package model

import (
	"net/url"
	"strings"
	"time"
)

// RouterConfig represents a normalized integration configuration payload.
type RouterConfig struct {
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

func (c RouterConfig) PollInterval() time.Duration {
	interval := time.Duration(c.PollIntervalSec) * time.Second
	if interval < 5*time.Second {
		return 5 * time.Second
	}
	return interval
}

func (c RouterConfig) BaseURL() string {
	defaultScheme := "https"
	if !c.SSL {
		defaultScheme = "http"
	}

	raw := strings.TrimSpace(c.Host)
	if raw == "" {
		return defaultScheme + ":///rest"
	}
	if !strings.Contains(raw, "://") {
		raw = defaultScheme + "://" + raw
	}

	parsed, err := url.Parse(raw)
	if err != nil || strings.TrimSpace(parsed.Host) == "" {
		host := strings.TrimSpace(c.Host)
		host = strings.TrimPrefix(strings.TrimPrefix(host, "http://"), "https://")
		host = strings.Trim(host, "/")
		return defaultScheme + "://" + host + "/rest"
	}

	scheme := strings.TrimSpace(parsed.Scheme)
	if scheme == "" {
		scheme = defaultScheme
	}
	path := strings.TrimSuffix(strings.TrimSpace(parsed.Path), "/")
	switch {
	case path == "", path == "/":
		path = "/rest"
	case strings.HasSuffix(path, "/rest"):
		// Keep an explicit REST path (for example behind reverse proxy).
	default:
		path = path + "/rest"
	}

	return scheme + "://" + parsed.Host + path
}
