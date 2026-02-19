package automation

import "context"

// CapabilityTargetRef identifies scope and subject for a state transition.
type CapabilityTargetRef struct {
	Scope    CapabilityScope `json:"scope"`
	DeviceID string          `json:"device_id,omitempty"`
}

// AutomationEngine coordinates capability state transitions and sync.
type AutomationEngine interface {
	SetCapabilityState(
		ctx context.Context,
		target CapabilityTargetRef,
		capabilityID string,
		newState string,
	) (SetStateResult, error)
	SyncOnce(ctx context.Context) error
}
