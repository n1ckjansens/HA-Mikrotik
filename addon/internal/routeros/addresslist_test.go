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

func TestAddressListAddRemoveAndIdempotency(t *testing.T) {
	t.Helper()

	var (
		mu      sync.Mutex
		nextID  = 1
		entries = map[string]map[string]string{}
	)

	api := &mockapi.Client{}
	api.RunFunc = func(ctx context.Context, cmd string, args ...string) (*goros.Reply, error) {
		_ = ctx
		params := decodeArgs(args)

		mu.Lock()
		defer mu.Unlock()

		switch cmd {
		case "/ip/firewall/address-list/print":
			list := params["?list"]
			rows := make([]map[string]string, 0)
			for id, row := range entries {
				if list != "" && row["list"] != list {
					continue
				}
				rows = append(rows, map[string]string{".id": id, "list": row["list"], "address": row["address"]})
			}
			return mockapi.Reply(rows...), nil
		case "/ip/firewall/address-list/add":
			for _, row := range entries {
				if row["list"] == params["list"] && equalAddressTarget(row["address"], params["address"]) {
					return nil, fmt.Errorf("already have such entry")
				}
			}
			id := fmt.Sprintf("*%d", nextID)
			nextID++
			entries[id] = map[string]string{"list": params["list"], "address": params["address"]}
			return mockapi.Reply(), nil
		case "/ip/firewall/address-list/remove":
			id := params[".id"]
			if _, ok := entries[id]; !ok {
				return nil, fmt.Errorf("no such item")
			}
			delete(entries, id)
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

	ctx := context.Background()
	if err := client.AddAddressToList(ctx, "vpn_users", "192.168.88.10"); err != nil {
		t.Fatalf("AddAddressToList first call failed: %v", err)
	}
	if err := client.AddAddressToList(ctx, "vpn_users", "192.168.88.10"); err != nil {
		t.Fatalf("AddAddressToList second call failed: %v", err)
	}
	if err := client.RemoveAddressFromList(ctx, "vpn_users", "192.168.88.10"); err != nil {
		t.Fatalf("RemoveAddressFromList first call failed: %v", err)
	}
	if err := client.RemoveAddressFromList(ctx, "vpn_users", "192.168.88.10"); err != nil {
		t.Fatalf("RemoveAddressFromList second call failed: %v", err)
	}

	calls := api.CallsSnapshot()
	addCalls := 0
	removeCalls := 0
	for _, call := range calls {
		if call.Cmd == "/ip/firewall/address-list/add" {
			addCalls++
		}
		if call.Cmd == "/ip/firewall/address-list/remove" {
			removeCalls++
		}
	}
	if addCalls != 1 {
		t.Fatalf("expected exactly one add command, got %d", addCalls)
	}
	if removeCalls != 1 {
		t.Fatalf("expected exactly one remove command, got %d", removeCalls)
	}
}

func TestAddAddressToListConcurrentIdempotent(t *testing.T) {
	t.Helper()

	var (
		mu      sync.Mutex
		nextID  = 1
		entries = map[string]map[string]string{}
	)

	api := &mockapi.Client{}
	api.RunFunc = func(ctx context.Context, cmd string, args ...string) (*goros.Reply, error) {
		_ = ctx
		params := decodeArgs(args)

		mu.Lock()
		defer mu.Unlock()

		switch cmd {
		case "/ip/firewall/address-list/print":
			rows := make([]map[string]string, 0)
			for id, row := range entries {
				if row["list"] != params["?list"] {
					continue
				}
				rows = append(rows, map[string]string{".id": id, "list": row["list"], "address": row["address"]})
			}
			return mockapi.Reply(rows...), nil
		case "/ip/firewall/address-list/add":
			for _, row := range entries {
				if row["list"] == params["list"] && equalAddressTarget(row["address"], params["address"]) {
					return nil, fmt.Errorf("already have such entry")
				}
			}
			id := fmt.Sprintf("*%d", nextID)
			nextID++
			entries[id] = map[string]string{"list": params["list"], "address": params["address"]}
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

	const workers = 16
	wg := sync.WaitGroup{}
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			if err := client.AddAddressToList(context.Background(), "vpn_users", "192.168.88.15"); err != nil {
				t.Errorf("AddAddressToList failed: %v", err)
			}
		}()
	}
	wg.Wait()

	calls := api.CallsSnapshot()
	addCalls := 0
	for _, call := range calls {
		if call.Cmd == "/ip/firewall/address-list/add" {
			addCalls++
		}
	}
	if addCalls != 1 {
		t.Fatalf("expected one add command under concurrency, got %d", addCalls)
	}
}

func decodeArgs(args []string) map[string]string {
	decoded := make(map[string]string, len(args))
	for _, arg := range args {
		switch {
		case strings.HasPrefix(arg, "="):
			parts := strings.SplitN(strings.TrimPrefix(arg, "="), "=", 2)
			if len(parts) == 2 {
				decoded[parts[0]] = parts[1]
			}
		case strings.HasPrefix(arg, "?"):
			parts := strings.SplitN(strings.TrimPrefix(arg, "?"), "=", 2)
			if len(parts) == 2 {
				decoded["?"+parts[0]] = parts[1]
			}
		}
	}
	return decoded
}
