package routeros

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/micro-ha/mikrotik-presence/addon/internal/model"
)

func TestAddressListMembershipAddAndRemove(t *testing.T) {
	t.Helper()

	var (
		mu          sync.Mutex
		rows        = []map[string]any{}
		putCount    int
		deleteCount int
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/ip/firewall/address-list" && r.Method == http.MethodPut {
			var payload map[string]string
			_ = json.NewDecoder(r.Body).Decode(&payload)
			mu.Lock()
			defer mu.Unlock()
			putCount += 1
			for _, row := range rows {
				if row["list"] == payload["list"] && row["address"] == payload["address"] {
					w.WriteHeader(http.StatusBadRequest)
					_, _ = w.Write([]byte(`{"detail":"already have such entry","error":400}`))
					return
				}
			}
			rows = append(rows, map[string]any{
				".id":     "*1",
				"list":    payload["list"],
				"address": payload["address"],
			})
			w.WriteHeader(http.StatusCreated)
			return
		}

		if r.URL.Path == "/rest/ip/firewall/address-list/print" && r.Method == http.MethodPost {
			var payload map[string]any
			_ = json.NewDecoder(r.Body).Decode(&payload)
			requestedList := ""
			if rawQuery, ok := payload[".query"]; ok {
				switch query := rawQuery.(type) {
				case []any:
					for _, item := range query {
						entry, _ := item.(string)
						if strings.HasPrefix(entry, "list=") {
							requestedList = strings.TrimPrefix(entry, "list=")
						}
					}
				case []string:
					for _, entry := range query {
						if strings.HasPrefix(entry, "list=") {
							requestedList = strings.TrimPrefix(entry, "list=")
						}
					}
				}
			}

			mu.Lock()
			defer mu.Unlock()
			if strings.TrimSpace(requestedList) == "" {
				_ = json.NewEncoder(w).Encode(rows)
				return
			}

			filtered := make([]map[string]any, 0, len(rows))
			for _, row := range rows {
				if row["list"] == requestedList {
					filtered = append(filtered, row)
				}
			}
			_ = json.NewEncoder(w).Encode(filtered)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/rest/ip/firewall/address-list/") && r.Method == http.MethodDelete {
			if strings.Contains(r.URL.Path, "%2A") {
				http.Error(w, `{"detail":"missing or invalid resource identifier","error":400}`, http.StatusBadRequest)
				return
			}
			mu.Lock()
			defer mu.Unlock()
			deleteCount += 1
			rows = []map[string]any{}
			w.WriteHeader(http.StatusNoContent)
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	cfg := model.RouterConfig{
		Host:            strings.TrimPrefix(server.URL, "http://"),
		Username:        "u",
		Password:        "p",
		SSL:             false,
		VerifyTLS:       true,
		PollIntervalSec: 5,
	}

	client := NewClient()

	if err := client.AddAddressListEntry(context.Background(), cfg, "VPN_CLIENTS", "192.168.88.10"); err != nil {
		t.Fatalf("add entry: %v", err)
	}
	if err := client.AddAddressListEntry(context.Background(), cfg, "VPN_CLIENTS", "192.168.88.10"); err != nil {
		t.Fatalf("add duplicate entry: %v", err)
	}

	mu.Lock()
	if putCount != 2 {
		t.Fatalf("expected 2 puts (second duplicate), got %d", putCount)
	}
	mu.Unlock()

	if err := client.RemoveAddressListEntry(context.Background(), cfg, "VPN_CLIENTS", "192.168.88.10"); err != nil {
		t.Fatalf("remove entry: %v", err)
	}

	mu.Lock()
	if deleteCount != 1 {
		t.Fatalf("expected 1 delete, got %d", deleteCount)
	}
	mu.Unlock()
}

func TestAddressListAddDoesNotFallbackToPostAdd(t *testing.T) {
	t.Helper()

	var (
		mu        sync.Mutex
		putCount  int
		postCount int
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/ip/firewall/address-list" && r.Method == http.MethodPut {
			mu.Lock()
			putCount += 1
			mu.Unlock()
			http.Error(w, `{"message":"Method Not Allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path == "/rest/ip/firewall/address-list/add" && r.Method == http.MethodPost {
			mu.Lock()
			postCount += 1
			mu.Unlock()
			w.WriteHeader(http.StatusCreated)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	cfg := model.RouterConfig{
		Host:            strings.TrimPrefix(server.URL, "http://"),
		Username:        "u",
		Password:        "p",
		SSL:             false,
		VerifyTLS:       true,
		PollIntervalSec: 5,
	}

	client := NewClient()
	err := client.AddAddressListEntry(context.Background(), cfg, "VPN_CLIENTS", "192.168.88.10")
	if err == nil {
		t.Fatalf("expected error when PUT is not supported")
	}

	mu.Lock()
	defer mu.Unlock()
	if putCount != 1 {
		t.Fatalf("expected 1 put, got %d", putCount)
	}
	if postCount != 0 {
		t.Fatalf("expected no post fallback, got %d", postCount)
	}
}

func TestNormalizeRouterObjectID_RemovesInvisibleFormatRunes(t *testing.T) {
	t.Helper()

	input := "*4300\u2060"
	got := normalizeRouterObjectID(input)
	if got != "*4300" {
		t.Fatalf("normalizeRouterObjectID(%q) = %q, want %q", input, got, "*4300")
	}
}
