package service

import "errors"

var (
	ErrCapabilityNotFound       = errors.New("capability not found")
	ErrCapabilityConflict       = errors.New("capability already exists")
	ErrCapabilityInvalid        = errors.New("capability invalid")
	ErrCapabilityStateInvalid   = errors.New("capability state invalid")
	ErrActionTypeUnknown        = errors.New("action type unknown")
	ErrDeviceNotFound           = errors.New("device not found")
	ErrIntegrationNotConfigured = errors.New("integration not configured")
)
