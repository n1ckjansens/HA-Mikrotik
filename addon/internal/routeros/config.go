package routeros

import (
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/micro-ha/mikrotik-presence/addon/internal/model"
)

// Config defines one RouterOS API connection profile.
type Config struct {
	Address   string
	Username  string
	Password  string
	UseTLS    bool
	VerifyTLS bool
	Timeout   time.Duration
}

func configFromModel(cfg model.RouterConfig) Config {
	return Config{
		Address:   strings.TrimSpace(cfg.Host),
		Username:  strings.TrimSpace(cfg.Username),
		Password:  cfg.Password,
		UseTLS:    cfg.SSL,
		VerifyTLS: cfg.VerifyTLS,
		Timeout:   10 * time.Second,
	}
}

func normalizeConfig(cfg Config) (Config, error) {
	cfg.Address = strings.TrimSpace(cfg.Address)
	cfg.Username = strings.TrimSpace(cfg.Username)
	cfg.Password = strings.TrimSpace(cfg.Password)
	if cfg.Timeout <= 0 {
		cfg.Timeout = 10 * time.Second
	}
	if cfg.Address == "" {
		return Config{}, &ValidationError{Field: "address", Reason: "is required"}
	}
	if cfg.Username == "" {
		return Config{}, &ValidationError{Field: "username", Reason: "is required"}
	}
	if cfg.Password == "" {
		return Config{}, &ValidationError{Field: "password", Reason: "is required"}
	}

	address, err := normalizeAddress(cfg.Address, cfg.UseTLS)
	if err != nil {
		return Config{}, err
	}
	cfg.Address = address
	return cfg, nil
}

func normalizeAddress(raw string, useTLS bool) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", &ValidationError{Field: "address", Reason: "is required"}
	}

	if strings.Contains(value, "/") && !strings.Contains(value, "://") {
		value = strings.Split(value, "/")[0]
	}
	if !strings.Contains(value, "://") {
		value = "routeros://" + value
	}

	parsed, err := url.Parse(value)
	if err != nil {
		return "", &ValidationError{Field: "address", Reason: fmt.Sprintf("invalid value: %v", err)}
	}

	host := strings.TrimSpace(parsed.Host)
	if host == "" {
		host = strings.TrimSpace(parsed.Path)
	}
	if host == "" {
		return "", &ValidationError{Field: "address", Reason: "host is empty"}
	}

	address, err := withDefaultPort(host, useTLS)
	if err != nil {
		return "", err
	}
	return address, nil
}

func withDefaultPort(host string, useTLS bool) (string, error) {
	port := "8728"
	if useTLS {
		port = "8729"
	}

	if _, _, err := net.SplitHostPort(host); err == nil {
		return host, nil
	}

	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		host = strings.TrimPrefix(strings.TrimSuffix(host, "]"), "[")
	}
	if strings.TrimSpace(host) == "" {
		return "", &ValidationError{Field: "address", Reason: "host is empty"}
	}

	return net.JoinHostPort(host, port), nil
}

func configKey(cfg Config) string {
	return strings.Join(
		[]string{
			cfg.Address,
			cfg.Username,
			cfg.Password,
			boolToWord(cfg.UseTLS),
			boolToWord(cfg.VerifyTLS),
		},
		"\x00",
	)
}
