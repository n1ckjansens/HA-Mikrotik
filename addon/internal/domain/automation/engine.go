package automation

import "context"

// AutomationEngine coordinates capability state transitions and sync.
type AutomationEngine interface {
	SetCapabilityState(ctx context.Context, deviceID, capabilityID, newState string) (SetStateResult, error)
	SyncOnce(ctx context.Context) error
}
