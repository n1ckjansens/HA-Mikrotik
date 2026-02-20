package routeros

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	goros "github.com/go-routeros/routeros/v3"
	mockapi "github.com/micro-ha/mikrotik-presence/addon/internal/routeros/mock"
)

func TestFirewallEnableDisable(t *testing.T) {
	t.Helper()

	var (
		mu    sync.Mutex
		rules = map[string]map[string]FirewallRule{
			"filter": {
				"*1": {ID: "*1", Table: "filter", Comment: "vpn", Disabled: false, Chain: "forward", Action: "accept"},
			},
		}
	)

	api := &mockapi.Client{}
	api.RunFunc = func(ctx context.Context, cmd string, args ...string) (*goros.Reply, error) {
		_ = ctx
		params := decodeArgs(args)

		mu.Lock()
		defer mu.Unlock()

		switch {
		case strings.HasSuffix(cmd, "/print"):
			table := tableFromCommand(cmd)
			rows := make([]map[string]string, 0)
			for _, rule := range rules[table] {
				rows = append(rows, map[string]string{
					".id":      rule.ID,
					"comment":  rule.Comment,
					"disabled": boolToWord(rule.Disabled),
					"chain":    rule.Chain,
					"action":   rule.Action,
				})
			}
			return mockapi.Reply(rows...), nil
		case strings.HasSuffix(cmd, "/set"):
			table := tableFromCommand(cmd)
			id := params[".id"]
			rule, ok := rules[table][id]
			if !ok {
				return nil, fmt.Errorf("no such item")
			}
			rule.Disabled = boolFromWord(params["disabled"])
			rules[table][id] = rule
			return mockapi.Reply(), nil
		default:
			return nil, fmt.Errorf("unexpected command %s", cmd)
		}
	}

	client := &Client{
		config: Config{Timeout: time.Second},
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		closed: make(chan struct{}),
		api:    api,
	}

	if err := client.DisableRule(context.Background(), "*1"); err != nil {
		t.Fatalf("DisableRule failed: %v", err)
	}
	if err := client.EnableRule(context.Background(), "*1"); err != nil {
		t.Fatalf("EnableRule failed: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if rules["filter"]["*1"].Disabled {
		t.Fatalf("expected rule to be enabled after toggle")
	}
}

func tableFromCommand(cmd string) string {
	parts := strings.Split(strings.Trim(cmd, "/"), "/")
	if len(parts) < 3 {
		return ""
	}
	return parts[2]
}
