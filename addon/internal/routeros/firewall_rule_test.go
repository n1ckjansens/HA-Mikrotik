package routeros

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/micro-ha/mikrotik-presence/addon/internal/model"
)

func TestSetFirewallRuleDisabled(t *testing.T) {
	var (
		requestPath string
		requestBody map[string]any
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestPath = r.URL.Path
		if r.Method != http.MethodPatch {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient()
	cfg := model.RouterConfig{
		Host:      strings.TrimPrefix(server.URL, "http://"),
		Username:  "u",
		Password:  "p",
		SSL:       false,
		VerifyTLS: true,
	}
	if err := client.SetFirewallRuleDisabled(context.Background(), cfg, "filter", "rule-1", true); err != nil {
		t.Fatalf("SetFirewallRuleDisabled returned error: %v", err)
	}

	if requestPath != "/rest/ip/firewall/filter/rule-1" {
		t.Fatalf("unexpected request path %q", requestPath)
	}
	disabled, ok := requestBody["disabled"].(bool)
	if !ok || !disabled {
		t.Fatalf("unexpected payload: %+v", requestBody)
	}
}

func TestSetFirewallRulesDisabledByComment(t *testing.T) {
	patchCalls := map[string]int{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/rest/ip/firewall/filter/print":
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{".id": "rule-1", "comment": "VPN_PROFILE"},
				{".id": "rule-2", "comment": "VPN_PROFILE"},
				{".id": "rule-3", "comment": "OTHER"},
			})
		case r.Method == http.MethodPatch && strings.HasPrefix(r.URL.Path, "/rest/ip/firewall/filter/"):
			patchCalls[r.URL.Path]++
			w.WriteHeader(http.StatusOK)
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewClient()
	cfg := model.RouterConfig{
		Host:      strings.TrimPrefix(server.URL, "http://"),
		Username:  "u",
		Password:  "p",
		SSL:       false,
		VerifyTLS: true,
	}
	if err := client.SetFirewallRulesDisabledByComment(context.Background(), cfg, "filter", "VPN_PROFILE", false); err != nil {
		t.Fatalf("SetFirewallRulesDisabledByComment returned error: %v", err)
	}

	if patchCalls["/rest/ip/firewall/filter/rule-1"] != 1 || patchCalls["/rest/ip/firewall/filter/rule-2"] != 1 {
		t.Fatalf("unexpected patch calls: %+v", patchCalls)
	}
	if _, exists := patchCalls["/rest/ip/firewall/filter/rule-3"]; exists {
		t.Fatalf("rule with non-matching comment should not be patched")
	}
}

func TestSetFirewallRulesDisabledByCommentReturnsErrorWhenNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/rest/ip/firewall/filter/print" {
			_ = json.NewEncoder(w).Encode([]map[string]any{})
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient()
	cfg := model.RouterConfig{
		Host:      strings.TrimPrefix(server.URL, "http://"),
		Username:  "u",
		Password:  "p",
		SSL:       false,
		VerifyTLS: true,
	}
	if err := client.SetFirewallRulesDisabledByComment(context.Background(), cfg, "filter", "VPN_PROFILE", true); err == nil {
		t.Fatalf("expected error when no rules match comment")
	}
}
