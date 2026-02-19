package automation

import "context"

// Repository defines persistence operations for automation domain.
type Repository interface {
	ListTemplates(ctx context.Context, search, category string) ([]CapabilityTemplate, error)
	GetTemplate(ctx context.Context, id string) (CapabilityTemplate, error)
	CreateTemplate(ctx context.Context, template CapabilityTemplate) error
	UpdateTemplate(ctx context.Context, template CapabilityTemplate) error
	DeleteTemplate(ctx context.Context, id string) error

	UpsertDeviceCapabilityState(ctx context.Context, state DeviceCapability) error
	GetDeviceCapabilityState(ctx context.Context, deviceID, capabilityID string) (DeviceCapability, bool, error)
	ListDeviceCapabilityStates(ctx context.Context, deviceID string) (map[string]DeviceCapability, error)
	ListCapabilityDeviceStates(ctx context.Context, capabilityID string) (map[string]DeviceCapability, error)

	GetGlobalCapability(ctx context.Context, capabilityID string) (*GlobalCapability, error)
	SaveGlobalCapability(ctx context.Context, capability *GlobalCapability) error
	ListGlobalCapabilities(ctx context.Context) ([]GlobalCapability, error)
}
