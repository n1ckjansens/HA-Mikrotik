package device

import "errors"

var (
	// ErrDeviceNotFound indicates missing device by MAC.
	ErrDeviceNotFound = errors.New("device not found")
	// ErrIntegrationNotConfigured indicates router credentials are missing.
	ErrIntegrationNotConfigured = errors.New("integration not configured")
)
