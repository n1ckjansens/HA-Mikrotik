package device

import (
	"context"
	"reflect"
	"sort"
	"testing"
	"time"

	devicedomain "github.com/micro-ha/mikrotik-presence/addon/internal/domain/device"
	"github.com/micro-ha/mikrotik-presence/addon/internal/model"
)

type memoryRepo struct {
	states      map[string]devicedomain.State
	registered  map[string]devicedomain.Registered
	newCache    map[string]devicedomain.NewCache
	deletedRows []string
}

func newMemoryRepo() *memoryRepo {
	return &memoryRepo{
		states:     map[string]devicedomain.State{},
		registered: map[string]devicedomain.Registered{},
		newCache:   map[string]devicedomain.NewCache{},
	}
}

func (r *memoryRepo) LoadAllStates(ctx context.Context) (map[string]devicedomain.State, error) {
	_ = ctx
	out := make(map[string]devicedomain.State, len(r.states))
	for key, value := range r.states {
		out[key] = value
	}
	return out, nil
}

func (r *memoryRepo) UpsertStates(ctx context.Context, states []devicedomain.State) error {
	_ = ctx
	for _, state := range states {
		r.states[state.MAC] = state
	}
	return nil
}

func (r *memoryRepo) DeleteStates(ctx context.Context, macs []string) error {
	_ = ctx
	for _, mac := range macs {
		delete(r.states, mac)
		r.deletedRows = append(r.deletedRows, mac)
	}
	return nil
}

func (r *memoryRepo) ListRegistered(ctx context.Context) (map[string]devicedomain.Registered, error) {
	_ = ctx
	out := make(map[string]devicedomain.Registered, len(r.registered))
	for key, value := range r.registered {
		out[key] = value
	}
	return out, nil
}

func (r *memoryRepo) UpsertRegistered(ctx context.Context, mac string, name, icon, comment *string) error {
	_ = ctx
	_ = name
	_ = icon
	_ = comment
	r.registered[mac] = devicedomain.Registered{MAC: mac, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	return nil
}

func (r *memoryRepo) PatchRegistered(ctx context.Context, mac string, name, icon, comment *string) error {
	_ = ctx
	_ = name
	_ = icon
	_ = comment
	if _, ok := r.registered[mac]; !ok {
		return nil
	}
	row := r.registered[mac]
	row.UpdatedAt = time.Now().UTC()
	r.registered[mac] = row
	return nil
}

func (r *memoryRepo) ListNewCache(ctx context.Context) (map[string]devicedomain.NewCache, error) {
	_ = ctx
	out := make(map[string]devicedomain.NewCache, len(r.newCache))
	for key, value := range r.newCache {
		out[key] = value
	}
	return out, nil
}

func (r *memoryRepo) UpsertNewCache(ctx context.Context, rows []devicedomain.NewCache) error {
	_ = ctx
	for _, row := range rows {
		r.newCache[row.MAC] = row
	}
	return nil
}

func (r *memoryRepo) DeleteNewCache(ctx context.Context, macs []string) error {
	_ = ctx
	for _, mac := range macs {
		delete(r.newCache, mac)
	}
	return nil
}

func TestPersistSnapshotKeepsUnregisteredObservedDevice(t *testing.T) {
	t.Helper()

	repo := newMemoryRepo()
	svc := &Service{repo: repo, thresholds: model.DefaultPresenceThresholds()}

	mac := "AA:BB:CC:DD:EE:42"
	observed := map[string]model.Observation{
		mac: {
			MAC:              mac,
			ObservedAt:       time.Now().UTC(),
			ConnectionStatus: model.ConnectionStatusIdleRecent,
			StatusReason:     "trace_present",
			Sources:          []string{model.SourceARP},
			ARPIP:            "192.168.88.42",
			ARPInterface:     "bridge",
			ARPIsComplete:    false,
			Generated:        "Device-EE42",
			Vendor:           "Unknown",
		},
	}

	if err := svc.persistSnapshot(context.Background(), observed); err != nil {
		t.Fatalf("persistSnapshot failed: %v", err)
	}

	if _, ok := repo.states[mac]; !ok {
		t.Fatalf("expected unregistered observed device to remain in state")
	}
	if _, ok := repo.newCache[mac]; !ok {
		t.Fatalf("expected unregistered observed device to remain in new cache")
	}
	if len(repo.deletedRows) != 0 {
		t.Fatalf("did not expect deletions, got %v", repo.deletedRows)
	}
}

func TestPersistSnapshotRemovesUnregisteredWhenNotObserved(t *testing.T) {
	t.Helper()

	repo := newMemoryRepo()
	mac := "AA:BB:CC:DD:EE:43"
	repo.states[mac] = devicedomain.State{
		MAC:              mac,
		ConnectionStatus: string(model.ConnectionStatusIdleRecent),
		StatusReason:     "trace_present",
		UpdatedAt:        time.Now().UTC(),
	}
	repo.newCache[mac] = devicedomain.NewCache{
		MAC:         mac,
		FirstSeenAt: time.Now().UTC().Add(-5 * time.Minute),
	}

	svc := &Service{repo: repo, thresholds: model.DefaultPresenceThresholds()}
	if err := svc.persistSnapshot(context.Background(), map[string]model.Observation{}); err != nil {
		t.Fatalf("persistSnapshot failed: %v", err)
	}

	if _, ok := repo.states[mac]; ok {
		t.Fatalf("expected stale unregistered device to be deleted from state")
	}
	if _, ok := repo.newCache[mac]; ok {
		t.Fatalf("expected stale unregistered device to be deleted from new cache")
	}

	deleted := append([]string(nil), repo.deletedRows...)
	sort.Strings(deleted)
	if !reflect.DeepEqual(deleted, []string{mac}) {
		t.Fatalf("unexpected deleted rows: %v", repo.deletedRows)
	}
}
