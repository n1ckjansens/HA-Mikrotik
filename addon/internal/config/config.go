package config

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/micro-ha/mikrotik-presence/addon/internal/model"
)

const (
	defaultHTTPAddr               = ":8099"
	defaultDBPath                 = "/data/mikrotik_presence.db"
	defaultFrontendDist           = "/app/frontend/dist"
	defaultAddonOptionsPath       = "/data/options.json"
	defaultAutomationSyncInterval = 20 * time.Second
	defaultConfigRefreshInterval  = 20 * time.Second
)

// Config stores runtime settings loaded from environment variables.
type Config struct {
	HTTPAddr               string
	DBPath                 string
	FrontendDist           string
	AddonOptionsPath       string
	ConfigRefreshInterval  time.Duration
	LogLevel               slog.Level
	AutomationSyncInterval time.Duration
	PresenceThresholds     model.PresenceThresholds
}

// Load builds Config from environment variables using stable defaults.
func Load() Config {
	return Config{
		HTTPAddr:               getenv("HTTP_ADDR", defaultHTTPAddr),
		DBPath:                 getenv("DB_PATH", defaultDBPath),
		FrontendDist:           getenv("FRONTEND_DIST", defaultFrontendDist),
		AddonOptionsPath:       getenv("ADDON_OPTIONS_PATH", defaultAddonOptionsPath),
		ConfigRefreshInterval:  parseDuration("CONFIG_REFRESH_INTERVAL", defaultConfigRefreshInterval),
		LogLevel:               parseLogLevel(getenv("LOG_LEVEL", "info")),
		AutomationSyncInterval: parseDuration("AUTOMATION_SYNC_INTERVAL", defaultAutomationSyncInterval),
		PresenceThresholds: model.PresenceThresholds{
			WiFiIdleThreshold:    parseDuration("WIFI_IDLE_THRESHOLD", 5*time.Minute),
			DHCPRecentThreshold:  parseDuration("DHCP_RECENT_THRESHOLD", 30*time.Minute),
			OfflineHardThreshold: parseDuration("OFFLINE_HARD_THRESHOLD", 24*time.Hour),
		}.Normalize(),
	}
}

// DBDir returns the target directory for DBPath.
func (c Config) DBDir() string {
	return filepath.Dir(c.DBPath)
}

func getenv(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return fallback
}

func parseDuration(key string, fallback time.Duration) time.Duration {
	raw, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	value, err := time.ParseDuration(strings.TrimSpace(raw))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func parseLogLevel(raw string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
