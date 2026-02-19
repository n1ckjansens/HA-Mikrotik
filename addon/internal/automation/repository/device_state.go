package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/micro-ha/mikrotik-presence/addon/internal/automation/domain"
)

func (r *Repository) UpsertDeviceCapabilityState(
	ctx context.Context,
	state domain.DeviceCapabilityState,
) error {
	if state.UpdatedAt.IsZero() {
		state.UpdatedAt = time.Now().UTC()
	}
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO device_capabilities_state(device_id, capability_id, enabled, state, updated_at)
		 VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT(device_id, capability_id) DO UPDATE SET
			enabled = excluded.enabled,
			state = excluded.state,
			updated_at = excluded.updated_at`,
		state.DeviceID,
		state.CapabilityID,
		state.Enabled,
		state.State,
		state.UpdatedAt.UTC().Format(time.RFC3339Nano),
	)
	return err
}

func (r *Repository) GetDeviceCapabilityState(
	ctx context.Context,
	deviceID string,
	capabilityID string,
) (domain.DeviceCapabilityState, bool, error) {
	var (
		state     domain.DeviceCapabilityState
		enabled   bool
		updatedAt string
	)
	err := r.db.QueryRowContext(
		ctx,
		`SELECT device_id, capability_id, enabled, state, updated_at
		 FROM device_capabilities_state
		 WHERE device_id = ? AND capability_id = ?`,
		deviceID,
		capabilityID,
	).Scan(&state.DeviceID, &state.CapabilityID, &enabled, &state.State, &updatedAt)
	if err == sql.ErrNoRows {
		return domain.DeviceCapabilityState{}, false, nil
	}
	if err != nil {
		return domain.DeviceCapabilityState{}, false, err
	}
	state.Enabled = enabled
	if parsed, err := time.Parse(time.RFC3339Nano, updatedAt); err == nil {
		state.UpdatedAt = parsed.UTC()
	}
	return state, true, nil
}

func (r *Repository) ListDeviceCapabilityStates(
	ctx context.Context,
	deviceID string,
) (map[string]domain.DeviceCapabilityState, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT device_id, capability_id, enabled, state, updated_at
		 FROM device_capabilities_state
		 WHERE device_id = ?`,
		deviceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]domain.DeviceCapabilityState)
	for rows.Next() {
		item, err := scanDeviceCapabilityState(rows)
		if err != nil {
			return nil, err
		}
		result[item.CapabilityID] = item
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *Repository) ListCapabilityDeviceStates(
	ctx context.Context,
	capabilityID string,
) (map[string]domain.DeviceCapabilityState, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT device_id, capability_id, enabled, state, updated_at
		 FROM device_capabilities_state
		 WHERE capability_id = ?`,
		capabilityID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]domain.DeviceCapabilityState)
	for rows.Next() {
		item, err := scanDeviceCapabilityState(rows)
		if err != nil {
			return nil, err
		}
		result[item.DeviceID] = item
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func scanDeviceCapabilityState(scanner interface {
	Scan(dest ...any) error
}) (domain.DeviceCapabilityState, error) {
	var (
		item      domain.DeviceCapabilityState
		enabled   bool
		updatedAt string
	)
	if err := scanner.Scan(&item.DeviceID, &item.CapabilityID, &enabled, &item.State, &updatedAt); err != nil {
		return domain.DeviceCapabilityState{}, fmt.Errorf("scan device capability: %w", err)
	}
	item.Enabled = enabled
	if parsed, err := time.Parse(time.RFC3339Nano, updatedAt); err == nil {
		item.UpdatedAt = parsed.UTC()
	}
	return item, nil
}
