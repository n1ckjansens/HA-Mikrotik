package service

import (
	"context"
	"io"
	"log/slog"
	"path/filepath"
	"testing"
	"time"

	"github.com/micro-ha/mikrotik-presence/addon/internal/model"
	"github.com/micro-ha/mikrotik-presence/addon/internal/storage"
)

func TestPersistSnapshot_RemovesUnregisteredWhenGone(t *testing.T) {
	ctx := context.Background()
	repo := newTestRepo(t, ctx)
	now := time.Now().UTC().Add(-1 * time.Minute)
	mac := "AA:BB:CC:DD:EE:01"

	if err := repo.UpsertStates(ctx, []model.DeviceState{{
		MAC:             mac,
		Online:          true,
		LastSourcesJSON: `["arp"]`,
		UpdatedAt:       now,
	}}); err != nil {
		t.Fatalf("seed state: %v", err)
	}
	if err := repo.UpsertNewCache(ctx, []model.DeviceNewCache{{
		MAC:           mac,
		FirstSeenAt:   now,
		Vendor:        "Unknown",
		GeneratedName: "Device-EE01",
	}}); err != nil {
		t.Fatalf("seed cache: %v", err)
	}

	svc := &Service{repo: repo}
	if err := svc.persistSnapshot(ctx, map[string]model.Observation{}); err != nil {
		t.Fatalf("persist snapshot: %v", err)
	}

	states, err := repo.LoadAllStates(ctx)
	if err != nil {
		t.Fatalf("load states: %v", err)
	}
	if len(states) != 0 {
		t.Fatalf("expected 0 states, got %d", len(states))
	}

	cache, err := repo.ListNewCache(ctx)
	if err != nil {
		t.Fatalf("load cache: %v", err)
	}
	if len(cache) != 0 {
		t.Fatalf("expected 0 cache rows, got %d", len(cache))
	}
}

func TestPersistSnapshot_RemovesUnregisteredObservedOffline(t *testing.T) {
	ctx := context.Background()
	repo := newTestRepo(t, ctx)
	mac := "AA:BB:CC:DD:EE:02"
	now := time.Now().UTC()

	svc := &Service{repo: repo}
	observed := map[string]model.Observation{
		mac: {
			MAC:       mac,
			Online:    false,
			Vendor:    "Unknown",
			Generated: "Device-EE02",
			LastSeenAt: func() *time.Time {
				v := now
				return &v
			}(),
		},
	}
	if err := svc.persistSnapshot(ctx, observed); err != nil {
		t.Fatalf("persist snapshot: %v", err)
	}

	states, err := repo.LoadAllStates(ctx)
	if err != nil {
		t.Fatalf("load states: %v", err)
	}
	if len(states) != 0 {
		t.Fatalf("expected 0 states, got %d", len(states))
	}

	cache, err := repo.ListNewCache(ctx)
	if err != nil {
		t.Fatalf("load cache: %v", err)
	}
	if len(cache) != 0 {
		t.Fatalf("expected 0 cache rows, got %d", len(cache))
	}
}

func TestPersistSnapshot_KeepsRegisteredOffline(t *testing.T) {
	ctx := context.Background()
	repo := newTestRepo(t, ctx)
	now := time.Now().UTC().Add(-1 * time.Minute)
	mac := "AA:BB:CC:DD:EE:03"

	if err := repo.UpsertStates(ctx, []model.DeviceState{{
		MAC:             mac,
		Online:          true,
		LastSourcesJSON: `["wifi"]`,
		UpdatedAt:       now,
	}}); err != nil {
		t.Fatalf("seed state: %v", err)
	}
	if err := repo.UpsertNewCache(ctx, []model.DeviceNewCache{{
		MAC:           mac,
		FirstSeenAt:   now,
		Vendor:        "Vendor",
		GeneratedName: "Vendor-EE03",
	}}); err != nil {
		t.Fatalf("seed cache: %v", err)
	}
	if err := repo.UpsertRegistered(ctx, mac, nil, nil, nil); err != nil {
		t.Fatalf("seed registered: %v", err)
	}

	svc := &Service{repo: repo}
	if err := svc.persistSnapshot(ctx, map[string]model.Observation{}); err != nil {
		t.Fatalf("persist snapshot: %v", err)
	}

	states, err := repo.LoadAllStates(ctx)
	if err != nil {
		t.Fatalf("load states: %v", err)
	}
	state, ok := states[mac]
	if !ok {
		t.Fatalf("expected state for %s", mac)
	}
	if state.Online {
		t.Fatalf("expected registered device to become offline")
	}

	cache, err := repo.ListNewCache(ctx)
	if err != nil {
		t.Fatalf("load cache: %v", err)
	}
	if _, ok := cache[mac]; !ok {
		t.Fatalf("expected cache row for registered device")
	}
}

func TestListDevices_SortsAllWithNewFirstThenRegisteredByRecency(t *testing.T) {
	ctx := context.Background()
	repo := newTestRepo(t, ctx)

	now := time.Now().UTC()
	macNewRecent := "AA:BB:CC:DD:EE:11"
	macNewOld := "AA:BB:CC:DD:EE:12"
	macRegOld := "AA:BB:CC:DD:EE:13"
	macRegRecent := "AA:BB:CC:DD:EE:14"

	if err := repo.UpsertStates(ctx, []model.DeviceState{
		{
			MAC:             macNewRecent,
			Online:          true,
			LastSourcesJSON: `["arp"]`,
			UpdatedAt:       now,
		},
		{
			MAC:             macNewOld,
			Online:          true,
			LastSourcesJSON: `["arp"]`,
			UpdatedAt:       now,
		},
		{
			MAC:             macRegOld,
			Online:          false,
			LastSourcesJSON: `[]`,
			UpdatedAt:       now,
		},
		{
			MAC:             macRegRecent,
			Online:          true,
			LastSourcesJSON: `["wifi"]`,
			UpdatedAt:       now,
		},
	}); err != nil {
		t.Fatalf("seed states: %v", err)
	}

	if err := repo.UpsertNewCache(ctx, []model.DeviceNewCache{
		{
			MAC:           macNewRecent,
			FirstSeenAt:   now.Add(-1 * time.Minute),
			Vendor:        "Unknown",
			GeneratedName: "Device-EE11",
		},
		{
			MAC:           macNewOld,
			FirstSeenAt:   now.Add(-15 * time.Minute),
			Vendor:        "Unknown",
			GeneratedName: "Device-EE12",
		},
		{
			MAC:           macRegOld,
			FirstSeenAt:   now.Add(-30 * time.Minute),
			Vendor:        "Vendor",
			GeneratedName: "Vendor-EE13",
		},
		{
			MAC:           macRegRecent,
			FirstSeenAt:   now.Add(-5 * time.Minute),
			Vendor:        "Vendor",
			GeneratedName: "Vendor-EE14",
		},
	}); err != nil {
		t.Fatalf("seed cache: %v", err)
	}

	if err := repo.UpsertRegistered(ctx, macRegOld, nil, nil, nil); err != nil {
		t.Fatalf("seed old registered: %v", err)
	}
	time.Sleep(2 * time.Millisecond)
	if err := repo.UpsertRegistered(ctx, macRegRecent, nil, nil, nil); err != nil {
		t.Fatalf("seed recent registered: %v", err)
	}

	svc := &Service{repo: repo}
	items, err := svc.ListDevices(ctx, ListFilter{})
	if err != nil {
		t.Fatalf("list devices: %v", err)
	}
	if len(items) != 4 {
		t.Fatalf("expected 4 devices, got %d", len(items))
	}

	expectedOrder := []string{macNewRecent, macNewOld, macRegRecent, macRegOld}
	for i, mac := range expectedOrder {
		if items[i].MAC != mac {
			t.Fatalf("unexpected order at %d: got %s, want %s", i, items[i].MAC, mac)
		}
	}
}

func newTestRepo(t *testing.T, ctx context.Context) *storage.Repository {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	dbPath := filepath.Join(t.TempDir(), "test.db")
	repo, err := storage.New(ctx, dbPath, logger)
	if err != nil {
		t.Fatalf("create repository: %v", err)
	}
	t.Cleanup(func() {
		_ = repo.Close()
	})
	return repo
}
