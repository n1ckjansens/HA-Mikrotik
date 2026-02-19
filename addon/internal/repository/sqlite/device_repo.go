package sqlite

import (
	"context"

	devicedomain "github.com/micro-ha/mikrotik-presence/addon/internal/domain/device"
)

// DeviceRepository is sqlite implementation of device.Repository.
type DeviceRepository struct {
	db *DB
}

// NewDeviceRepository creates sqlite-backed device repository.
func NewDeviceRepository(db *DB) *DeviceRepository {
	return &DeviceRepository{db: db}
}

// LoadAllStates returns persisted device states by MAC.
func (r *DeviceRepository) LoadAllStates(ctx context.Context) (map[string]devicedomain.State, error) {
	return r.db.storage.LoadAllStates(ctx)
}

// UpsertStates stores observed state rows.
func (r *DeviceRepository) UpsertStates(ctx context.Context, states []devicedomain.State) error {
	return r.db.storage.UpsertStates(ctx, states)
}

// DeleteStates removes state rows for given MACs.
func (r *DeviceRepository) DeleteStates(ctx context.Context, macs []string) error {
	return r.db.storage.DeleteStates(ctx, macs)
}

// ListRegistered returns registered devices by MAC.
func (r *DeviceRepository) ListRegistered(ctx context.Context) (map[string]devicedomain.Registered, error) {
	return r.db.storage.ListRegistered(ctx)
}

// UpsertRegistered creates or updates registered device metadata.
func (r *DeviceRepository) UpsertRegistered(
	ctx context.Context,
	mac string,
	name *string,
	icon *string,
	comment *string,
) error {
	return r.db.storage.UpsertRegistered(ctx, mac, name, icon, comment)
}

// PatchRegistered updates selected registered metadata fields.
func (r *DeviceRepository) PatchRegistered(
	ctx context.Context,
	mac string,
	name *string,
	icon *string,
	comment *string,
) error {
	return r.db.storage.PatchRegistered(ctx, mac, name, icon, comment)
}

// ListNewCache returns new-device cache rows by MAC.
func (r *DeviceRepository) ListNewCache(ctx context.Context) (map[string]devicedomain.NewCache, error) {
	return r.db.storage.ListNewCache(ctx)
}

// UpsertNewCache writes new-device cache rows.
func (r *DeviceRepository) UpsertNewCache(ctx context.Context, rows []devicedomain.NewCache) error {
	return r.db.storage.UpsertNewCache(ctx, rows)
}

// DeleteNewCache removes new-device cache rows for MACs.
func (r *DeviceRepository) DeleteNewCache(ctx context.Context, macs []string) error {
	return r.db.storage.DeleteNewCache(ctx, macs)
}
