package subnet

import (
	"testing"

	"github.com/micro-ha/mikrotik-presence/addon/internal/routeros"
)

func TestLongestPrefixMatch(t *testing.T) {
	matcher := New().WithAddresses([]routeros.IPAddress{
		{Address: "10.0.0.1/8"},
		{Address: "10.1.0.1/16"},
		{Address: "10.1.2.1/24"},
	})

	if got := matcher.Match("10.1.2.99"); got != "10.1.2.1/24" {
		t.Fatalf("expected /24 match, got %s", got)
	}
	if got := matcher.Match("10.1.50.2"); got != "10.1.0.1/16" {
		t.Fatalf("expected /16 match, got %s", got)
	}
	if got := matcher.Match("10.99.1.1"); got != "10.0.0.1/8" {
		t.Fatalf("expected /8 match, got %s", got)
	}
	if got := matcher.Match("192.168.1.2"); got != "" {
		t.Fatalf("expected empty match, got %s", got)
	}
}
