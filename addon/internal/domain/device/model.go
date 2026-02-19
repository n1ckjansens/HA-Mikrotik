package device

import (
	"time"

	"github.com/micro-ha/mikrotik-presence/addon/internal/model"
)

// ConnectionStatus is device connectivity state persisted by backend.
type ConnectionStatus = model.ConnectionStatus

const (
	// ConnectionStatusOnline indicates active network activity.
	ConnectionStatusOnline = model.ConnectionStatusOnline
	// ConnectionStatusIdleRecent indicates stale but recent activity.
	ConnectionStatusIdleRecent = model.ConnectionStatusIdleRecent
	// ConnectionStatusOffline indicates inactive device.
	ConnectionStatusOffline = model.ConnectionStatusOffline
	// ConnectionStatusUnknown indicates no known signal.
	ConnectionStatusUnknown = model.ConnectionStatusUnknown
)

// PresenceThresholds defines transitions between device statuses.
type PresenceThresholds = model.PresenceThresholds

// Observation is a merged network snapshot for one MAC.
type Observation = model.Observation

// State stores persisted derived signals for a device.
type State = model.DeviceState

// Registered stores user metadata for a known device.
type Registered = model.DeviceRegistered

// NewCache stores metadata for not-yet-registered devices.
type NewCache = model.DeviceNewCache

// Device represents UI/API read model for a single device.
type Device = model.DeviceView

// RegisterInput is API payload for creating/updating registered metadata.
type RegisterInput struct {
	Name    *string `json:"name"`
	Icon    *string `json:"icon"`
	Comment *string `json:"comment"`
}

// ListFilter applies device list query constraints.
type ListFilter struct {
	Status string
	Online *bool
	Query  string
}

// PollSnapshotResult summarizes writes from one poll cycle.
type PollSnapshotResult struct {
	ObservedCount int
	UpdatedAt     time.Time
}
