package storage

import (
	"testing"
	"time"

	"github.com/micro-ha/mikrotik-presence/addon/internal/model"
)

func TestMergeDeviceViews_ZeroFirstSeenIsNil(t *testing.T) {
	mac := "AA:BB:CC:DD:EE:10"
	updatedAt := time.Now().UTC()
	createdAt := time.Now().UTC().Add(-1 * time.Hour)

	items := MergeDeviceViews(
		map[string]model.DeviceState{
			mac: {
				MAC:             mac,
				Online:          false,
				LastSourcesJSON: "[]",
				UpdatedAt:       updatedAt,
			},
		},
		map[string]model.DeviceRegistered{
			mac: {
				MAC:       mac,
				CreatedAt: createdAt,
				UpdatedAt: createdAt,
			},
		},
		map[string]model.DeviceNewCache{},
	)

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].FirstSeenAt != nil {
		t.Fatalf("expected nil first_seen_at when cache first_seen is zero")
	}
}
