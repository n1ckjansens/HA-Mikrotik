package device

import "context"

// Repository defines persistent storage operations for device domain.
type Repository interface {
	LoadAllStates(ctx context.Context) (map[string]State, error)
	UpsertStates(ctx context.Context, states []State) error
	DeleteStates(ctx context.Context, macs []string) error

	ListRegistered(ctx context.Context) (map[string]Registered, error)
	UpsertRegistered(ctx context.Context, mac string, name, icon, comment *string) error
	PatchRegistered(ctx context.Context, mac string, name, icon, comment *string) error

	ListNewCache(ctx context.Context) (map[string]NewCache, error)
	UpsertNewCache(ctx context.Context, rows []NewCache) error
	DeleteNewCache(ctx context.Context, macs []string) error
}
