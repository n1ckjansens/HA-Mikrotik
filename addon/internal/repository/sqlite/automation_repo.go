package sqlite

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	legacyautomationdomain "github.com/micro-ha/mikrotik-presence/addon/internal/automation/domain"
	legacyautomationrepo "github.com/micro-ha/mikrotik-presence/addon/internal/automation/repository"
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
	items, err := r.db.automationRepo.ListTemplates(ctx, search, category)
	if err != nil {
		return nil, fmt.Errorf("list templates: %w", err)
	}
	out := make([]automationdomain.CapabilityTemplate, 0, len(items))
	for _, item := range items {
		converted, err := toCapabilityTemplate(item)
		if err != nil {
			return nil, err
		}
		out = append(out, converted)
	}
	return out, nil
}

// GetTemplate returns capability template by ID.
func (r *AutomationRepository) GetTemplate(ctx context.Context, id string) (automationdomain.CapabilityTemplate, error) {
	item, err := r.db.automationRepo.GetTemplate(ctx, id)
	if errors.Is(err, legacyautomationrepo.ErrNotFound) {
		return automationdomain.CapabilityTemplate{}, automationdomain.ErrNotFound
	}
	if err != nil {
		return automationdomain.CapabilityTemplate{}, fmt.Errorf("get template: %w", err)
	}
	return toCapabilityTemplate(item)
}

// CreateTemplate inserts capability template.
func (r *AutomationRepository) CreateTemplate(ctx context.Context, template automationdomain.CapabilityTemplate) error {
	legacyTemplate, err := toLegacyCapabilityTemplate(template)
	if err != nil {
		return err
	}
	return r.db.automationRepo.CreateTemplate(ctx, legacyTemplate)
}

// UpdateTemplate updates capability template.
func (r *AutomationRepository) UpdateTemplate(ctx context.Context, template automationdomain.CapabilityTemplate) error {
	legacyTemplate, err := toLegacyCapabilityTemplate(template)
	if err != nil {
		return err
	}
	err = r.db.automationRepo.UpdateTemplate(ctx, legacyTemplate)
	if errors.Is(err, legacyautomationrepo.ErrNotFound) {
		return automationdomain.ErrNotFound
	}
	return err
}

// DeleteTemplate deletes template row by ID.
func (r *AutomationRepository) DeleteTemplate(ctx context.Context, id string) error {
	err := r.db.automationRepo.DeleteTemplate(ctx, id)
	if errors.Is(err, legacyautomationrepo.ErrNotFound) {
		return automationdomain.ErrNotFound
	}
	return err
}

// UpsertDeviceCapabilityState stores device capability state.
func (r *AutomationRepository) UpsertDeviceCapabilityState(ctx context.Context, state automationdomain.DeviceCapability) error {
	return r.db.automationRepo.UpsertDeviceCapabilityState(ctx, legacyautomationdomain.DeviceCapabilityState{
		DeviceID:     state.DeviceID,
		CapabilityID: state.CapabilityID,
		Enabled:      state.Enabled,
		State:        state.State,
		UpdatedAt:    state.UpdatedAt,
	})
}

// GetDeviceCapabilityState returns device capability state if it exists.
func (r *AutomationRepository) GetDeviceCapabilityState(
	ctx context.Context,
	deviceID string,
	capabilityID string,
) (automationdomain.DeviceCapability, bool, error) {
	item, ok, err := r.db.automationRepo.GetDeviceCapabilityState(ctx, deviceID, capabilityID)
	if err != nil {
		return automationdomain.DeviceCapability{}, false, err
	}
	return automationdomain.DeviceCapability{
		DeviceID:     item.DeviceID,
		CapabilityID: item.CapabilityID,
		Enabled:      item.Enabled,
		State:        item.State,
		UpdatedAt:    item.UpdatedAt,
	}, ok, nil
}

// ListDeviceCapabilityStates returns states by capability ID for one device.
func (r *AutomationRepository) ListDeviceCapabilityStates(
	ctx context.Context,
	deviceID string,
) (map[string]automationdomain.DeviceCapability, error) {
	items, err := r.db.automationRepo.ListDeviceCapabilityStates(ctx, deviceID)
	if err != nil {
		return nil, err
	}
	out := make(map[string]automationdomain.DeviceCapability, len(items))
	for key, item := range items {
		out[key] = automationdomain.DeviceCapability{
			DeviceID:     item.DeviceID,
			CapabilityID: item.CapabilityID,
			Enabled:      item.Enabled,
			State:        item.State,
			UpdatedAt:    item.UpdatedAt,
		}
	}
	return out, nil
}

// ListCapabilityDeviceStates returns states by device ID for one capability.
func (r *AutomationRepository) ListCapabilityDeviceStates(
	ctx context.Context,
	capabilityID string,
) (map[string]automationdomain.DeviceCapability, error) {
	items, err := r.db.automationRepo.ListCapabilityDeviceStates(ctx, capabilityID)
	if err != nil {
		return nil, err
	}
	out := make(map[string]automationdomain.DeviceCapability, len(items))
	for key, item := range items {
		out[key] = automationdomain.DeviceCapability{
			DeviceID:     item.DeviceID,
			CapabilityID: item.CapabilityID,
			Enabled:      item.Enabled,
			State:        item.State,
			UpdatedAt:    item.UpdatedAt,
		}
	}
	return out, nil
}

func toCapabilityTemplate(item legacyautomationdomain.CapabilityTemplate) (automationdomain.CapabilityTemplate, error) {
	body, err := json.Marshal(item)
	if err != nil {
		return automationdomain.CapabilityTemplate{}, fmt.Errorf("encode legacy template: %w", err)
	}
	var out automationdomain.CapabilityTemplate
	if err := json.Unmarshal(body, &out); err != nil {
		return automationdomain.CapabilityTemplate{}, fmt.Errorf("decode capability template: %w", err)
	}
	return out, nil
}

func toLegacyCapabilityTemplate(item automationdomain.CapabilityTemplate) (legacyautomationdomain.CapabilityTemplate, error) {
	body, err := json.Marshal(item)
	if err != nil {
		return legacyautomationdomain.CapabilityTemplate{}, fmt.Errorf("encode capability template: %w", err)
	}
	var out legacyautomationdomain.CapabilityTemplate
	if err := json.Unmarshal(body, &out); err != nil {
		return legacyautomationdomain.CapabilityTemplate{}, fmt.Errorf("decode legacy template: %w", err)
	}
	return out, nil
}
