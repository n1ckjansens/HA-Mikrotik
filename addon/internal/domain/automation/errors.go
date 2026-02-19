package automation

import "errors"

var (
	// ErrCapabilityNotFound means a capability template is missing.
	ErrCapabilityNotFound = errors.New("capability not found")
	// ErrCapabilityConflict means template id already exists.
	ErrCapabilityConflict = errors.New("capability already exists")
	// ErrCapabilityInvalid means payload failed validation.
	ErrCapabilityInvalid = errors.New("capability invalid")
	// ErrCapabilityStateInvalid means target state is unsupported.
	ErrCapabilityStateInvalid = errors.New("capability state invalid")
	// ErrDeviceNotFound means target device is missing.
	ErrDeviceNotFound = errors.New("device not found")
	// ErrIntegrationNotConfigured means MikroTik/HA config is absent.
	ErrIntegrationNotConfigured = errors.New("integration not configured")
	// ErrNotFound is generic repository-level missing row marker.
	ErrNotFound = errors.New("not found")
)
