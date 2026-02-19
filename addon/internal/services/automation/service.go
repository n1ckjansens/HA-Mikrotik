package automation

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	automationdomain "github.com/micro-ha/mikrotik-presence/addon/internal/domain/automation"
	devicedomain "github.com/micro-ha/mikrotik-presence/addon/internal/domain/device"
	"github.com/micro-ha/mikrotik-presence/addon/internal/pkg/utils"
	"github.com/micro-ha/mikrotik-presence/addon/internal/services/automation/engine"
	"github.com/micro-ha/mikrotik-presence/addon/internal/services/automation/registry"
)

// Service implements automation.Service.
type Service struct {
	repo     automationdomain.Repository
	devices  devicedomain.Service
	engine   *engine.Engine
	registry *registry.Registry
	logger   *slog.Logger
}

// New creates automation service and binds automation engine.
func New(
	repo automationdomain.Repository,
	devices devicedomain.Service,
	engine *engine.Engine,
	registry *registry.Registry,
	logger *slog.Logger,
) *Service {
	return &Service{
		repo:     repo,
		devices:  devices,
		engine:   engine,
		registry: registry,
		logger:   logger,
	}
}

// ActionTypes returns registered action metadata.
func (s *Service) ActionTypes() []automationdomain.ActionMetadata {
	return s.registry.ActionTypes()
}

// StateSourceTypes returns registered state-source metadata.
func (s *Service) StateSourceTypes() []automationdomain.StateSourceMetadata {
	return s.registry.StateSourceTypes()
}

// ListCapabilities returns capability templates filtered by search/category.
func (s *Service) ListCapabilities(
	ctx context.Context,
	search string,
	category string,
) ([]automationdomain.CapabilityTemplate, error) {
	return s.repo.ListTemplates(ctx, search, category)
}

// GetCapability returns one capability template by ID.
func (s *Service) GetCapability(
	ctx context.Context,
	capabilityID string,
) (automationdomain.CapabilityTemplate, error) {
	item, err := s.repo.GetTemplate(ctx, strings.TrimSpace(capabilityID))
	if errors.Is(err, automationdomain.ErrNotFound) {
		return automationdomain.CapabilityTemplate{}, automationdomain.ErrCapabilityNotFound
	}
	if err != nil {
		return automationdomain.CapabilityTemplate{}, err
	}
	return item, nil
}

// CreateCapability validates and inserts capability template.
func (s *Service) CreateCapability(ctx context.Context, template automationdomain.CapabilityTemplate) error {
	template = normalizeTemplate(template)
	if err := validateTemplate(template, s.registry); err != nil {
		return fmt.Errorf("%w: %s", automationdomain.ErrCapabilityInvalid, err)
	}
	if err := s.repo.CreateTemplate(ctx, template); err != nil {
		if utils.IsUniqueConstraintError(err) {
			return automationdomain.ErrCapabilityConflict
		}
		return err
	}
	return nil
}

// UpdateCapability validates and updates capability template.
func (s *Service) UpdateCapability(
	ctx context.Context,
	capabilityID string,
	template automationdomain.CapabilityTemplate,
) error {
	capabilityID = strings.TrimSpace(capabilityID)
	if capabilityID == "" {
		return fmt.Errorf("%w: capability id is required", automationdomain.ErrCapabilityInvalid)
	}
	if strings.TrimSpace(template.ID) != capabilityID {
		return fmt.Errorf("%w: capability id in path and payload must match", automationdomain.ErrCapabilityInvalid)
	}
	if template.Sync == nil {
		// Keep existing sync config when clients (older UI) send payloads without `sync`.
		existing, err := s.repo.GetTemplate(ctx, capabilityID)
		if errors.Is(err, automationdomain.ErrNotFound) {
			return automationdomain.ErrCapabilityNotFound
		}
		if err != nil {
			return err
		}
		template.Sync = existing.Sync
	}

	template = normalizeTemplate(template)
	if err := validateTemplate(template, s.registry); err != nil {
		return fmt.Errorf("%w: %s", automationdomain.ErrCapabilityInvalid, err)
	}
	if err := s.repo.UpdateTemplate(ctx, template); err != nil {
		if errors.Is(err, automationdomain.ErrNotFound) {
			return automationdomain.ErrCapabilityNotFound
		}
		return err
	}
	return nil
}

// DeleteCapability deletes capability template by ID.
func (s *Service) DeleteCapability(ctx context.Context, capabilityID string) error {
	if err := s.repo.DeleteTemplate(ctx, strings.TrimSpace(capabilityID)); err != nil {
		if errors.Is(err, automationdomain.ErrNotFound) {
			return automationdomain.ErrCapabilityNotFound
		}
		return err
	}
	return nil
}

// GetDeviceCapabilities returns per-device capabilities for controls UI.
func (s *Service) GetDeviceCapabilities(
	ctx context.Context,
	deviceID string,
) ([]automationdomain.CapabilityUIModel, error) {
	deviceID = normalizeDeviceID(deviceID)
	if _, err := s.requireDevice(ctx, deviceID); err != nil {
		return nil, err
	}

	templates, err := s.repo.ListTemplates(ctx, "", "")
	if err != nil {
		return nil, err
	}
	states, err := s.repo.ListDeviceCapabilityStates(ctx, deviceID)
	if err != nil {
		return nil, err
	}

	items := make([]automationdomain.CapabilityUIModel, 0, len(templates))
	for _, template := range templates {
		if automationdomain.NormalizeCapabilityScope(template.Scope) != automationdomain.ScopeDevice {
			continue
		}
		state := template.DefaultState
		enabled := true
		if saved, ok := states[template.ID]; ok {
			if strings.TrimSpace(saved.State) != "" {
				state = saved.State
			}
			enabled = saved.Enabled
		}
		items = append(items, automationdomain.CapabilityUIModel{
			ID:          template.ID,
			Label:       template.Label,
			Description: template.Description,
			Control: automationdomain.CapabilityControlDTO{
				Type:    template.Control.Type,
				Options: append([]automationdomain.CapabilityControlOption(nil), template.Control.Options...),
			},
			State:   state,
			Enabled: enabled,
		})
	}
	sortCapabilityUIModels(items)
	return items, nil
}

// ListCapabilityAssignments returns capability state for all devices.
func (s *Service) ListCapabilityAssignments(
	ctx context.Context,
	capabilityID string,
) ([]automationdomain.CapabilityDeviceAssignment, error) {
	capabilityID = strings.TrimSpace(capabilityID)
	template, err := s.repo.GetTemplate(ctx, capabilityID)
	if errors.Is(err, automationdomain.ErrNotFound) {
		return nil, automationdomain.ErrCapabilityNotFound
	}
	if err != nil {
		return nil, err
	}
	if automationdomain.NormalizeCapabilityScope(template.Scope) != automationdomain.ScopeDevice {
		return nil, fmt.Errorf("%w: capability %q is not device-scoped", automationdomain.ErrCapabilityScopeMismatch, template.ID)
	}

	devices, err := s.devices.ListDevices(ctx, devicedomain.ListFilter{})
	if err != nil {
		return nil, err
	}
	states, err := s.repo.ListCapabilityDeviceStates(ctx, capabilityID)
	if err != nil {
		return nil, err
	}

	items := make([]automationdomain.CapabilityDeviceAssignment, 0, len(devices))
	for _, device := range devices {
		state := template.DefaultState
		enabled := true
		normalizedMAC := normalizeDeviceID(device.MAC)
		if saved, ok := states[normalizedMAC]; ok {
			if strings.TrimSpace(saved.State) != "" {
				state = saved.State
			}
			enabled = saved.Enabled
		}
		assignment := automationdomain.CapabilityDeviceAssignment{
			DeviceID:   normalizedMAC,
			DeviceName: device.Name,
			Online:     device.Online,
			Enabled:    enabled,
			State:      state,
		}
		if device.LastIP != nil {
			assignment.DeviceIP = *device.LastIP
		}
		items = append(items, assignment)
	}
	sortCapabilityAssignments(items)
	return items, nil
}

// PatchDeviceCapability updates state and/or enabled flag for one device capability.
func (s *Service) PatchDeviceCapability(
	ctx context.Context,
	deviceID string,
	capabilityID string,
	state *string,
	enabled *bool,
) (automationdomain.SetStateResult, error) {
	result := automationdomain.SetStateResult{OK: true}
	if state == nil && enabled == nil {
		return automationdomain.SetStateResult{}, fmt.Errorf("%w: either state or enabled must be provided", automationdomain.ErrCapabilityInvalid)
	}

	if state != nil {
		stateResult, err := s.engine.SetCapabilityState(ctx, automationdomain.CapabilityTargetRef{
			Scope:    automationdomain.ScopeDevice,
			DeviceID: deviceID,
		}, capabilityID, *state)
		if err != nil {
			return automationdomain.SetStateResult{}, err
		}
		result.Warnings = append(result.Warnings, stateResult.Warnings...)
	}

	if enabled != nil {
		if err := s.SetDeviceCapabilityEnabled(ctx, deviceID, capabilityID, *enabled); err != nil {
			return automationdomain.SetStateResult{}, err
		}
	}
	return result, nil
}

// SetDeviceCapabilityEnabled toggles capability without executing actions.
func (s *Service) SetDeviceCapabilityEnabled(
	ctx context.Context,
	deviceID string,
	capabilityID string,
	enabled bool,
) error {
	deviceID = normalizeDeviceID(deviceID)
	capabilityID = strings.TrimSpace(capabilityID)

	if _, err := s.requireDevice(ctx, deviceID); err != nil {
		return err
	}
	template, err := s.repo.GetTemplate(ctx, capabilityID)
	if errors.Is(err, automationdomain.ErrNotFound) {
		return automationdomain.ErrCapabilityNotFound
	}
	if err != nil {
		return err
	}
	if automationdomain.NormalizeCapabilityScope(template.Scope) != automationdomain.ScopeDevice {
		return fmt.Errorf("%w: capability %q is not device-scoped", automationdomain.ErrCapabilityScopeMismatch, template.ID)
	}

	current, exists, err := s.repo.GetDeviceCapabilityState(ctx, deviceID, capabilityID)
	if err != nil {
		return err
	}
	if !exists {
		current = automationdomain.DeviceCapability{
			DeviceID:     deviceID,
			CapabilityID: capabilityID,
			State:        template.DefaultState,
		}
	}
	current.Enabled = enabled
	current.UpdatedAt = time.Now().UTC()
	if strings.TrimSpace(current.State) == "" {
		current.State = template.DefaultState
	}
	return s.repo.UpsertDeviceCapabilityState(ctx, current)
}

// GetGlobalCapabilities returns global capabilities for controls UI.
func (s *Service) GetGlobalCapabilities(
	ctx context.Context,
) ([]automationdomain.CapabilityUIModel, error) {
	templates, err := s.repo.ListTemplates(ctx, "", "")
	if err != nil {
		return nil, err
	}
	states, err := s.repo.ListGlobalCapabilities(ctx)
	if err != nil {
		return nil, err
	}
	stateMap := make(map[string]automationdomain.GlobalCapability, len(states))
	for _, item := range states {
		stateMap[item.CapabilityID] = item
	}

	items := make([]automationdomain.CapabilityUIModel, 0, len(templates))
	for _, template := range templates {
		if automationdomain.NormalizeCapabilityScope(template.Scope) != automationdomain.ScopeGlobal {
			continue
		}
		state := template.DefaultState
		enabled := true
		if saved, ok := stateMap[template.ID]; ok {
			if strings.TrimSpace(saved.State) != "" {
				state = saved.State
			}
			enabled = saved.Enabled
		}
		items = append(items, automationdomain.CapabilityUIModel{
			ID:          template.ID,
			Label:       template.Label,
			Description: template.Description,
			Control: automationdomain.CapabilityControlDTO{
				Type:    template.Control.Type,
				Options: append([]automationdomain.CapabilityControlOption(nil), template.Control.Options...),
			},
			State:   state,
			Enabled: enabled,
		})
	}
	sortCapabilityUIModels(items)
	return items, nil
}

// PatchGlobalCapability updates state and/or enabled flag for global capability.
func (s *Service) PatchGlobalCapability(
	ctx context.Context,
	capabilityID string,
	state *string,
	enabled *bool,
) (automationdomain.SetStateResult, error) {
	result := automationdomain.SetStateResult{OK: true}
	if state == nil && enabled == nil {
		return automationdomain.SetStateResult{}, fmt.Errorf("%w: either state or enabled must be provided", automationdomain.ErrCapabilityInvalid)
	}

	if state != nil {
		stateResult, err := s.engine.SetCapabilityState(ctx, automationdomain.CapabilityTargetRef{
			Scope: automationdomain.ScopeGlobal,
		}, capabilityID, *state)
		if err != nil {
			return automationdomain.SetStateResult{}, err
		}
		result.Warnings = append(result.Warnings, stateResult.Warnings...)
	}

	if enabled != nil {
		if err := s.SetGlobalCapabilityEnabled(ctx, capabilityID, *enabled); err != nil {
			return automationdomain.SetStateResult{}, err
		}
	}
	return result, nil
}

// SetGlobalCapabilityEnabled toggles global capability without executing actions.
func (s *Service) SetGlobalCapabilityEnabled(
	ctx context.Context,
	capabilityID string,
	enabled bool,
) error {
	capabilityID = strings.TrimSpace(capabilityID)
	template, err := s.repo.GetTemplate(ctx, capabilityID)
	if errors.Is(err, automationdomain.ErrNotFound) {
		return automationdomain.ErrCapabilityNotFound
	}
	if err != nil {
		return err
	}
	if automationdomain.NormalizeCapabilityScope(template.Scope) != automationdomain.ScopeGlobal {
		return fmt.Errorf("%w: capability %q is not global-scoped", automationdomain.ErrCapabilityScopeMismatch, template.ID)
	}

	current, err := s.repo.GetGlobalCapability(ctx, capabilityID)
	if err != nil {
		return err
	}
	if current == nil {
		current = &automationdomain.GlobalCapability{
			CapabilityID: capabilityID,
			State:        template.DefaultState,
			Enabled:      enabled,
		}
	} else {
		current.Enabled = enabled
		if strings.TrimSpace(current.State) == "" {
			current.State = template.DefaultState
		}
	}
	return s.repo.SaveGlobalCapability(ctx, current)
}

func (s *Service) requireDevice(ctx context.Context, deviceID string) (devicedomain.Device, error) {
	item, err := s.devices.GetDevice(ctx, deviceID)
	if errors.Is(err, devicedomain.ErrDeviceNotFound) {
		return devicedomain.Device{}, automationdomain.ErrDeviceNotFound
	}
	if err != nil {
		return devicedomain.Device{}, err
	}
	return item, nil
}
