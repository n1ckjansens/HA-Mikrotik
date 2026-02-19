package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	automationdomain "github.com/micro-ha/mikrotik-presence/addon/internal/domain/automation"
)

// AutomationRepository is sqlite implementation of automation.Repository.
type AutomationRepository struct {
	db *DB
}

// NewAutomationRepository creates sqlite-backed automation repository.
func NewAutomationRepository(db *DB) *AutomationRepository {
	return &AutomationRepository{db: db}
}

// ListTemplates returns capability templates.
func (r *AutomationRepository) ListTemplates(
	ctx context.Context,
	search string,
	category string,
) ([]automationdomain.CapabilityTemplate, error) {
	rows, err := r.db.SQLDB().QueryContext(ctx, `SELECT id, data FROM capability_templates ORDER BY id`) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("list templates: %w", err)
	}
	defer rows.Close()

	search = strings.ToLower(strings.TrimSpace(search))
	category = strings.ToLower(strings.TrimSpace(category))
	items := make([]automationdomain.CapabilityTemplate, 0)
	for rows.Next() {
		var (
			id      string
			encoded string
		)
		if err := rows.Scan(&id, &encoded); err != nil {
			return nil, err
		}
		item, err := decodeTemplate(id, encoded)
		if err != nil {
			if r.db.logger != nil {
				r.db.logger.Warn("failed to decode capability template", "id", id, "err", err)
			}
			continue
		}
		if !matchesTemplateFilter(item, search, category) {
			continue
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// GetTemplate returns capability template by ID.
func (r *AutomationRepository) GetTemplate(ctx context.Context, id string) (automationdomain.CapabilityTemplate, error) {
	var encoded string
	err := r.db.SQLDB().QueryRowContext(
		ctx,
		`SELECT data FROM capability_templates WHERE id = ?`,
		id,
	).Scan(&encoded)
	if err == sql.ErrNoRows {
		return automationdomain.CapabilityTemplate{}, automationdomain.ErrNotFound
	}
	if err != nil {
		return automationdomain.CapabilityTemplate{}, fmt.Errorf("get template: %w", err)
	}
	return decodeTemplate(id, encoded)
}

// CreateTemplate inserts capability template.
func (r *AutomationRepository) CreateTemplate(ctx context.Context, template automationdomain.CapabilityTemplate) error {
	template.Scope = automationdomain.NormalizeCapabilityScope(template.Scope)
	encoded, err := json.Marshal(template)
	if err != nil {
		return fmt.Errorf("encode template: %w", err)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err = r.db.SQLDB().ExecContext(
		ctx,
		`INSERT INTO capability_templates(id, data, created_at, updated_at) VALUES (?, ?, ?, ?)`,
		template.ID,
		string(encoded),
		now,
		now,
	)
	return err
}

// UpdateTemplate updates capability template.
func (r *AutomationRepository) UpdateTemplate(ctx context.Context, template automationdomain.CapabilityTemplate) error {
	template.Scope = automationdomain.NormalizeCapabilityScope(template.Scope)
	encoded, err := json.Marshal(template)
	if err != nil {
		return fmt.Errorf("encode template: %w", err)
	}
	res, err := r.db.SQLDB().ExecContext(
		ctx,
		`UPDATE capability_templates SET data = ?, updated_at = ? WHERE id = ?`,
		string(encoded),
		time.Now().UTC().Format(time.RFC3339Nano),
		template.ID,
	)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return automationdomain.ErrNotFound
	}
	return nil
}

// DeleteTemplate deletes template row by ID.
func (r *AutomationRepository) DeleteTemplate(ctx context.Context, id string) error {
	res, err := r.db.SQLDB().ExecContext(ctx, `DELETE FROM capability_templates WHERE id = ?`, id)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return automationdomain.ErrNotFound
	}
	return nil
}

// UpsertDeviceCapabilityState stores device capability state.
func (r *AutomationRepository) UpsertDeviceCapabilityState(ctx context.Context, state automationdomain.DeviceCapability) error {
	if state.UpdatedAt.IsZero() {
		state.UpdatedAt = time.Now().UTC()
	}
	_, err := r.db.SQLDB().ExecContext(
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

// GetDeviceCapabilityState returns device capability state if it exists.
func (r *AutomationRepository) GetDeviceCapabilityState(
	ctx context.Context,
	deviceID string,
	capabilityID string,
) (automationdomain.DeviceCapability, bool, error) {
	var (
		state     automationdomain.DeviceCapability
		enabled   bool
		updatedAt string
	)
	err := r.db.SQLDB().QueryRowContext(
		ctx,
		`SELECT device_id, capability_id, enabled, state, updated_at
		 FROM device_capabilities_state
		 WHERE device_id = ? AND capability_id = ?`,
		deviceID,
		capabilityID,
	).Scan(&state.DeviceID, &state.CapabilityID, &enabled, &state.State, &updatedAt)
	if err == sql.ErrNoRows {
		return automationdomain.DeviceCapability{}, false, nil
	}
	if err != nil {
		return automationdomain.DeviceCapability{}, false, err
	}
	state.Enabled = enabled
	if parsed, err := time.Parse(time.RFC3339Nano, updatedAt); err == nil {
		state.UpdatedAt = parsed.UTC()
	}
	return state, true, nil
}

// ListDeviceCapabilityStates returns states by capability ID for one device.
func (r *AutomationRepository) ListDeviceCapabilityStates(
	ctx context.Context,
	deviceID string,
) (map[string]automationdomain.DeviceCapability, error) {
	rows, err := r.db.SQLDB().QueryContext(
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

	out := make(map[string]automationdomain.DeviceCapability)
	for rows.Next() {
		item, err := scanDeviceCapability(rows)
		if err != nil {
			return nil, err
		}
		out[item.CapabilityID] = item
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// ListCapabilityDeviceStates returns states by device ID for one capability.
func (r *AutomationRepository) ListCapabilityDeviceStates(
	ctx context.Context,
	capabilityID string,
) (map[string]automationdomain.DeviceCapability, error) {
	rows, err := r.db.SQLDB().QueryContext(
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

	out := make(map[string]automationdomain.DeviceCapability)
	for rows.Next() {
		item, err := scanDeviceCapability(rows)
		if err != nil {
			return nil, err
		}
		out[item.DeviceID] = item
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// GetGlobalCapability returns stored global capability state.
func (r *AutomationRepository) GetGlobalCapability(
	ctx context.Context,
	capabilityID string,
) (*automationdomain.GlobalCapability, error) {
	var item automationdomain.GlobalCapability
	err := r.db.SQLDB().QueryRowContext(
		ctx,
		`SELECT capability_id, enabled, state
		 FROM global_capabilities_state
		 WHERE capability_id = ?`,
		capabilityID,
	).Scan(&item.CapabilityID, &item.Enabled, &item.State)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// SaveGlobalCapability stores global capability state.
func (r *AutomationRepository) SaveGlobalCapability(
	ctx context.Context,
	capability *automationdomain.GlobalCapability,
) error {
	if capability == nil {
		return fmt.Errorf("global capability is nil")
	}
	_, err := r.db.SQLDB().ExecContext(
		ctx,
		`INSERT INTO global_capabilities_state(capability_id, enabled, state, updated_at)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(capability_id) DO UPDATE SET
			enabled = excluded.enabled,
			state = excluded.state,
			updated_at = excluded.updated_at`,
		capability.CapabilityID,
		capability.Enabled,
		capability.State,
		time.Now().UTC().Format(time.RFC3339Nano),
	)
	return err
}

// ListGlobalCapabilities returns all stored global capability states.
func (r *AutomationRepository) ListGlobalCapabilities(
	ctx context.Context,
) ([]automationdomain.GlobalCapability, error) {
	rows, err := r.db.SQLDB().QueryContext(
		ctx,
		`SELECT capability_id, enabled, state
		 FROM global_capabilities_state`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]automationdomain.GlobalCapability, 0)
	for rows.Next() {
		var item automationdomain.GlobalCapability
		if err := rows.Scan(&item.CapabilityID, &item.Enabled, &item.State); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func decodeTemplate(id string, encoded string) (automationdomain.CapabilityTemplate, error) {
	var template automationdomain.CapabilityTemplate
	if err := json.Unmarshal([]byte(encoded), &template); err != nil {
		return automationdomain.CapabilityTemplate{}, err
	}
	if strings.TrimSpace(template.ID) == "" {
		template.ID = id
	}
	template.Scope = automationdomain.NormalizeCapabilityScope(template.Scope)
	return template, nil
}

func matchesTemplateFilter(item automationdomain.CapabilityTemplate, search string, category string) bool {
	if category != "" && strings.ToLower(strings.TrimSpace(item.Category)) != category {
		return false
	}
	if search == "" {
		return true
	}
	haystack := strings.ToLower(strings.Join([]string{item.ID, item.Label, item.Description}, " "))
	return strings.Contains(haystack, search)
}

func scanDeviceCapability(scanner interface {
	Scan(dest ...any) error
}) (automationdomain.DeviceCapability, error) {
	var (
		item      automationdomain.DeviceCapability
		enabled   bool
		updatedAt string
	)
	if err := scanner.Scan(&item.DeviceID, &item.CapabilityID, &enabled, &item.State, &updatedAt); err != nil {
		return automationdomain.DeviceCapability{}, fmt.Errorf("scan device capability: %w", err)
	}
	item.Enabled = enabled
	if parsed, err := time.Parse(time.RFC3339Nano, updatedAt); err == nil {
		item.UpdatedAt = parsed.UTC()
	}
	return item, nil
}
