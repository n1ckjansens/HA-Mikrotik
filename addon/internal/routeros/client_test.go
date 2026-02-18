package routeros

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/micro-ha/mikrotik-presence/addon/internal/model"
)

type mockRoundTripper struct {
	payload map[string]any
}

func (m mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	body, _ := json.Marshal(m.payload[req.URL.Path])
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(body)),
		Request:    req,
	}, nil
}

func TestFetchSnapshotWithMockRouterOS(t *testing.T) {
	transport := mockRoundTripper{payload: map[string]any{
		"/rest/ip/dhcp-server/lease":              []map[string]any{{"mac-address": "AA:BB:CC:DD:EE:FF", "address": "192.168.88.10", "status": "bound"}},
		"/rest/interface/wifi/registration-table": []map[string]any{{"mac-address": "AA:BB:CC:DD:EE:FF", "interface": "wifi1"}},
		"/rest/interface/bridge/host":             []map[string]any{{"mac-address": "11:22:33:44:55:66", "on-interface": "bridge"}},
		"/rest/ip/arp":                            []map[string]any{{"mac-address": "11:22:33:44:55:66", "address": "192.168.88.20"}},
		"/rest/ip/address":                        []map[string]any{{"address": "192.168.88.1/24"}},
	}}

	httpClient := &http.Client{Transport: transport}
	client := NewClientWithHTTPClient(httpClient)
	cfg := model.RouterConfig{Host: "router.local", Username: "u", Password: "p", SSL: false, VerifyTLS: true, PollIntervalSec: 5}

	snapshot, err := client.FetchSnapshot(context.Background(), cfg)
	if err != nil {
		t.Fatalf("fetch failed: %v", err)
	}
	if len(snapshot.DHCP) != 1 || len(snapshot.WiFi) != 1 || len(snapshot.Bridge) != 1 || len(snapshot.ARP) != 1 {
		t.Fatalf("unexpected snapshot lengths: %+v", snapshot)
	}
}
