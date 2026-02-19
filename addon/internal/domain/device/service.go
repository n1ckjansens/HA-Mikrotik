package device

import "context"

// Service exposes device use-cases used by HTTP and automation layers.
type Service interface {
	PollOnce(ctx context.Context) error
	ListDevices(ctx context.Context, filter ListFilter) ([]Device, error)
	GetDevice(ctx context.Context, mac string) (Device, error)
	RegisterDevice(ctx context.Context, mac string, in RegisterInput) error
	PatchDevice(ctx context.Context, mac string, in RegisterInput) error
}
