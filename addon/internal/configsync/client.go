package configsync

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"hash"
	"hash/fnv"
	"os"
	"strconv"
	"strings"

	"github.com/micro-ha/mikrotik-presence/addon/internal/model"
)

type FetchResult struct {
	Configured bool
	Config     model.RouterConfig
}

type Client struct {
	optionsPath string
}

func NewClient(optionsPath string) *Client {
	optionsPath = strings.TrimSpace(optionsPath)
	if optionsPath == "" {
		optionsPath = "/data/options.json"
	}
	return &Client{
		optionsPath: optionsPath,
	}
}

type optionsPayload struct {
	RouterHost      string   `json:"router_host"`
	RouterUsername  string   `json:"router_username"`
	RouterPassword  string   `json:"router_password"`
	RouterSSL       *bool    `json:"router_ssl"`
	RouterVerifyTLS *bool    `json:"router_verify_tls"`
	PollIntervalSec int      `json:"poll_interval_sec"`
	Roles           []string `json:"roles"`
	LegacyHost      string   `json:"host"`
	LegacyUsername  string   `json:"username"`
	LegacyPassword  string   `json:"password"`
	LegacySSL       *bool    `json:"ssl"`
	LegacyVerifyTLS *bool    `json:"verify_tls"`
}

func (c *Client) FetchConfig(ctx context.Context) (FetchResult, error) {
	options, err := c.loadOptionsFromFile(ctx)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return FetchResult{}, err
		}
		options = loadOptionsFromEnv()
	}
	cfg := model.RouterConfig{
		Host:            firstNonEmpty(options.RouterHost, options.LegacyHost),
		Username:        firstNonEmpty(options.RouterUsername, options.LegacyUsername),
		Password:        firstNonEmpty(options.RouterPassword, options.LegacyPassword),
		SSL:             pickBool(options.RouterSSL, options.LegacySSL, parseBoolEnv("ROUTER_SSL", false)),
		VerifyTLS:       pickBool(options.RouterVerifyTLS, options.LegacyVerifyTLS, parseBoolEnv("ROUTER_VERIFY_TLS", false)),
		PollIntervalSec: firstPositive(options.PollIntervalSec, parseIntEnv("ROUTER_POLL_INTERVAL_SEC", 5)),
		Roles:           options.Roles,
	}

	if cfg.PollIntervalSec < 5 {
		cfg.PollIntervalSec = 5
	}
	if len(cfg.Roles) == 0 {
		cfg.Roles = nil
	}
	if strings.TrimSpace(cfg.Host) == "" || strings.TrimSpace(cfg.Username) == "" || strings.TrimSpace(cfg.Password) == "" {
		return FetchResult{Configured: false}, nil
	}
	cfg.Version = configVersion(cfg)
	return FetchResult{Configured: true, Config: cfg}, nil
}

func (c *Client) loadOptionsFromFile(ctx context.Context) (optionsPayload, error) {
	select {
	case <-ctx.Done():
		return optionsPayload{}, ctx.Err()
	default:
	}

	raw, err := os.ReadFile(c.optionsPath)
	if err != nil {
		return optionsPayload{}, err
	}
	if len(raw) == 0 {
		return optionsPayload{}, nil
	}

	var payload optionsPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return optionsPayload{}, err
	}
	return payload, nil
}

func loadOptionsFromEnv() optionsPayload {
	return optionsPayload{
		RouterHost:      strings.TrimSpace(os.Getenv("ROUTER_HOST")),
		RouterUsername:  strings.TrimSpace(os.Getenv("ROUTER_USERNAME")),
		RouterPassword:  strings.TrimSpace(os.Getenv("ROUTER_PASSWORD")),
		PollIntervalSec: parseIntEnv("ROUTER_POLL_INTERVAL_SEC", 5),
	}
}

func pickBool(primary *bool, fallback *bool, defaultValue bool) bool {
	if primary != nil {
		return *primary
	}
	if fallback != nil {
		return *fallback
	}
	return defaultValue
}

func parseBoolEnv(key string, fallback bool) bool {
	raw, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

func parseIntEnv(key string, fallback int) int {
	raw, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	value := strings.TrimSpace(raw)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func firstPositive(values ...int) int {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}

func configVersion(cfg model.RouterConfig) int64 {
	hasher := fnv.New64a()
	writeHashString(hasher, cfg.Host)
	writeHashString(hasher, cfg.Username)
	writeHashString(hasher, cfg.Password)
	writeHashString(hasher, boolToString(cfg.SSL))
	writeHashString(hasher, boolToString(cfg.VerifyTLS))
	writeHashString(hasher, strings.TrimSpace(strings.Join(cfg.Roles, ",")))
	var interval [8]byte
	binary.LittleEndian.PutUint64(interval[:], uint64(cfg.PollIntervalSec))
	_, _ = hasher.Write(interval[:])
	return int64(hasher.Sum64())
}

func writeHashString(hasher hash.Hash64, value string) {
	_, _ = hasher.Write([]byte(strings.TrimSpace(value)))
	_, _ = hasher.Write([]byte{0})
}

func boolToString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}
