package configsync

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestFetchConfigFromOptionsFile(t *testing.T) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "options.json")
	if err := os.WriteFile(path, []byte(`{
		"router_host": "192.168.88.1",
		"router_username": "admin",
		"router_password": "secret",
		"router_ssl": true,
		"router_verify_tls": false,
		"poll_interval_sec": 3
	}`), 0o644); err != nil {
		t.Fatalf("write options file: %v", err)
	}

	client := NewClient(path)
	got, err := client.FetchConfig(context.Background())
	if err != nil {
		t.Fatalf("FetchConfig() error: %v", err)
	}
	if !got.Configured {
		t.Fatalf("FetchConfig() configured = false, want true")
	}
	if got.Config.Host != "192.168.88.1" {
		t.Fatalf("Host = %q, want %q", got.Config.Host, "192.168.88.1")
	}
	if got.Config.PollIntervalSec != 5 {
		t.Fatalf("PollIntervalSec = %d, want 5", got.Config.PollIntervalSec)
	}
	if !got.Config.SSL {
		t.Fatalf("SSL = false, want true")
	}
	if got.Config.VerifyTLS {
		t.Fatalf("VerifyTLS = true, want false")
	}
}

func TestFetchConfigReturnsNotConfiguredWhenCredentialsMissing(t *testing.T) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "options.json")
	if err := os.WriteFile(path, []byte(`{
		"router_host": "192.168.88.1",
		"router_username": "",
		"router_password": "secret"
	}`), 0o644); err != nil {
		t.Fatalf("write options file: %v", err)
	}

	client := NewClient(path)
	got, err := client.FetchConfig(context.Background())
	if err != nil {
		t.Fatalf("FetchConfig() error: %v", err)
	}
	if got.Configured {
		t.Fatalf("FetchConfig() configured = true, want false")
	}
}

func TestFetchConfigFallsBackToEnvWhenOptionsFileMissing(t *testing.T) {
	t.Helper()

	t.Setenv("ROUTER_HOST", "192.168.1.1")
	t.Setenv("ROUTER_USERNAME", "env-user")
	t.Setenv("ROUTER_PASSWORD", "env-pass")
	t.Setenv("ROUTER_SSL", "true")
	t.Setenv("ROUTER_VERIFY_TLS", "true")
	t.Setenv("ROUTER_POLL_INTERVAL_SEC", "9")

	client := NewClient(filepath.Join(t.TempDir(), "missing-options.json"))
	got, err := client.FetchConfig(context.Background())
	if err != nil {
		t.Fatalf("FetchConfig() error: %v", err)
	}
	if !got.Configured {
		t.Fatalf("FetchConfig() configured = false, want true")
	}
	if got.Config.Host != "192.168.1.1" {
		t.Fatalf("Host = %q, want %q", got.Config.Host, "192.168.1.1")
	}
	if got.Config.Username != "env-user" {
		t.Fatalf("Username = %q, want %q", got.Config.Username, "env-user")
	}
	if got.Config.Password != "env-pass" {
		t.Fatalf("Password = %q, want %q", got.Config.Password, "env-pass")
	}
	if !got.Config.SSL {
		t.Fatalf("SSL = false, want true")
	}
	if !got.Config.VerifyTLS {
		t.Fatalf("VerifyTLS = false, want true")
	}
	if got.Config.PollIntervalSec != 9 {
		t.Fatalf("PollIntervalSec = %d, want 9", got.Config.PollIntervalSec)
	}
}

func TestFetchConfigReturnsErrorForInvalidOptionsJSON(t *testing.T) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "options.json")
	if err := os.WriteFile(path, []byte("{"), 0o644); err != nil {
		t.Fatalf("write options file: %v", err)
	}

	client := NewClient(path)
	if _, err := client.FetchConfig(context.Background()); err == nil {
		t.Fatal("FetchConfig() error = nil, want non-nil")
	}
}
