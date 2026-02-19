package automation

import "context"

// Service exposes automation CRUD and assignment use-cases.
type Service interface {
	ActionTypes() []ActionMetadata
	StateSourceTypes() []StateSourceMetadata

	ListCapabilities(ctx context.Context, search string, category string) ([]CapabilityTemplate, error)
	GetCapability(ctx context.Context, capabilityID string) (CapabilityTemplate, error)
	CreateCapability(ctx context.Context, template CapabilityTemplate) error
	UpdateCapability(ctx context.Context, capabilityID string, template CapabilityTemplate) error
	DeleteCapability(ctx context.Context, capabilityID string) error

	GetDeviceCapabilities(ctx context.Context, deviceID string) ([]CapabilityUIModel, error)
	ListCapabilityAssignments(ctx context.Context, capabilityID string) ([]CapabilityDeviceAssignment, error)
	PatchDeviceCapability(
		ctx context.Context,
		deviceID string,
		capabilityID string,
		state *string,
		enabled *bool,
	) (SetStateResult, error)
}
