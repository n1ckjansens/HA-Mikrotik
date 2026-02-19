package aggregator

import (
	"testing"
	"time"

	"github.com/micro-ha/mikrotik-presence/addon/internal/routeros"
	"github.com/micro-ha/mikrotik-presence/addon/internal/subnet"
)

type fakeOUI struct{}

func (fakeOUI) Lookup(_ string) string { return "VendorY" }

func TestAggregatePriorityAndOnline(t *testing.T) {
	now := time.Date(2026, 2, 18, 12, 0, 0, 0, time.UTC)
	agg := New(subnet.New(), fakeOUI{})

	snap := &routeros.Snapshot{
		FetchedAt: now,
		DHCP: []routeros.DHCPLease{
			{MAC: "AA:BB:CC:DD:EE:FF", Address: "192.168.88.10", HostName: "phone", Status: "bound", LastSeen: "5s"},
		},
		WiFi: []routeros.WiFiRegistration{
			{MAC: "AA:BB:CC:DD:EE:FF", Interface: "wifi1", LastActivity: "30s"},
		},
		ARP: []routeros.ARPEntry{
			{MAC: "11:22:33:44:55:66", Address: "192.168.88.20", Interface: "bridge"},
		},
		Addresses: []routeros.IPAddress{
			{Address: "192.168.88.1/24"},
		},
	}

	items := agg.Aggregate(snap)
	first, ok := items["AA:BB:CC:DD:EE:FF"]
	if !ok {
		t.Fatalf("expected first device")
	}
	if !first.Online {
		t.Fatalf("expected first device online")
	}
	if first.LastSeenAt == nil {
		t.Fatalf("expected last seen")
	}
	if first.LastSeenAt.Before(now.Add(-time.Minute)) || first.LastSeenAt.After(now) {
		t.Fatalf("expected wifi/dhcp last_seen near now, got %v", first.LastSeenAt)
	}
	if first.LastSubnet != "192.168.88.1/24" {
		t.Fatalf("expected subnet match, got %s", first.LastSubnet)
	}
	if first.ConnectionStatus != "ONLINE" {
		t.Fatalf("expected connection status ONLINE, got %s", first.ConnectionStatus)
	}

	second, ok := items["11:22:33:44:55:66"]
	if !ok {
		t.Fatalf("expected arp device")
	}
	if second.Online {
		t.Fatalf("expected arp-only incomplete device offline")
	}
	if second.ConnectionStatus != "IDLE_RECENT" {
		t.Fatalf("expected arp-only incomplete device idle_recent, got %s", second.ConnectionStatus)
	}
	if second.Vendor != "VendorY" {
		t.Fatalf("expected vendor from oui, got %s", second.Vendor)
	}
}

func TestAggregateDHCPIdleRecent(t *testing.T) {
	now := time.Date(2026, 2, 18, 12, 0, 0, 0, time.UTC)
	agg := New(subnet.New(), fakeOUI{})

	snap := &routeros.Snapshot{
		FetchedAt: now,
		DHCP: []routeros.DHCPLease{
			{
				MAC:      "AA:BB:CC:DD:EE:01",
				Address:  "192.168.88.30",
				Status:   "bound",
				LastSeen: "2h",
			},
		},
	}

	items := agg.Aggregate(snap)
	item, ok := items["AA:BB:CC:DD:EE:01"]
	if !ok {
		t.Fatalf("expected device")
	}
	if item.Online {
		t.Fatalf("expected stale dhcp device offline")
	}
	if item.ConnectionStatus != "IDLE_RECENT" {
		t.Fatalf("expected IDLE_RECENT, got %s", item.ConnectionStatus)
	}
}

func TestParseRouterOSDuration(t *testing.T) {
	cases := map[string]time.Duration{
		"5s":       5 * time.Second,
		"2m10s":    2*time.Minute + 10*time.Second,
		"1h2m3s":   time.Hour + 2*time.Minute + 3*time.Second,
		"1d":       24 * time.Hour,
		"00:01:05": time.Minute + 5*time.Second,
	}
	for in, expected := range cases {
		got, err := parseRouterOSDuration(in)
		if err != nil {
			t.Fatalf("unexpected parse error for %s: %v", in, err)
		}
		if got != expected {
			t.Fatalf("duration mismatch for %s: expected %v got %v", in, expected, got)
		}
	}
}
