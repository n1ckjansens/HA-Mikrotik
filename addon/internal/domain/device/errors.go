package device

import "errors"

var (
	// ErrDeviceNotFound indicates missing device by MAC.
	ErrDeviceNotFound = errors.New("device not found")
	// ErrAddonNotConfigured indicates router credentials are missing in add-on options.
	ErrAddonNotConfigured = errors.New("addon not configured")
	// ErrIntegrationNotConfigured is kept as a backwards-compatible alias.
	ErrIntegrationNotConfigured = ErrAddonNotConfigured
)
