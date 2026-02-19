package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/micro-ha/mikrotik-presence/addon/internal/automation/actions"
	"github.com/micro-ha/mikrotik-presence/addon/internal/automation/domain"
	"github.com/micro-ha/mikrotik-presence/addon/internal/automation/repository"
	"github.com/micro-ha/mikrotik-presence/addon/internal/model"
	presence "github.com/micro-ha/mikrotik-presence/addon/internal/service"
	"github.com/micro-ha/mikrotik-presence/addon/internal/storage"
)

type DeviceService interface {
	GetDevice(ctx context.Context, mac string) (model.DeviceView, error)
	ListDevices(ctx context.Context, filter presence.ListFilter) ([]model.DeviceView, error)
}

type RouterConfigProvider interface {
	Get() (model.RouterConfig, bool)
}

type Service struct {
	repo    *repository.Repository
	devices DeviceService
	config  RouterConfigProvider
	router  actions.AddressListClient
	logger  *slog.Logger
}

func New(
	repo *repository.Repository,
	devices DeviceService,
	config RouterConfigProvider,
	router actions.AddressListClient,
	logger *slog.Logger,
) *Service {
	return &Service{
		repo:    repo,
		devices: devices,
		config:  config,
		router:  router,
		logger:  logger,
	}
}

func (s *Service) ActionTypes() []domain.ActionType {
	return actions.ActionTypes()
}

func (s *Service) ListCapabilities(
	ctx context.Context,
	search string,
	category string,
) ([]domain.CapabilityTemplate, error) {
	return s.repo.ListTemplates(ctx, search, category)
}

func (s *Service) GetCapability(ctx context.Context, capabilityID string) (domain.CapabilityTemplate, error) {
	item, err := s.repo.GetTemplate(ctx, strings.TrimSpace(capabilityID))
	if errors.Is(err, repository.ErrNotFound) {
		return domain.CapabilityTemplate{}, ErrCapabilityNotFound
	}
	if err != nil {
		return domain.CapabilityTemplate{}, err
	}
	return item, nil
}

func (s *Service) CreateCapability(ctx context.Context, template domain.CapabilityTemplate) error {
	template = normalizeTemplate(template)
	if err := validateTemplate(template); err != nil {
		return fmt.Errorf("%w: %s", ErrCapabilityInvalid, err)
	}
	if err := s.repo.CreateTemplate(ctx, template); err != nil {
		if isUniqueConstraintErr(err) {
			return ErrCapabilityConflict
		}
		return err
	}
	return nil
}

func (s *Service) UpdateCapability(
	ctx context.Context,
	capabilityID string,
	template domain.CapabilityTemplate,
) error {
	capabilityID = strings.TrimSpace(capabilityID)
	if capabilityID == "" {
		return fmt.Errorf("%w: capability id is required", ErrCapabilityInvalid)
	}
	if strings.TrimSpace(template.ID) != capabilityID {
		return fmt.Errorf("%w: capability id in path and payload must match", ErrCapabilityInvalid)
	}
	template = normalizeTemplate(template)
	if err := validateTemplate(template); err != nil {
		return fmt.Errorf("%w: %s", ErrCapabilityInvalid, err)
	}
	if err := s.repo.UpdateTemplate(ctx, template); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrCapabilityNotFound
		}
		return err
	}
	return nil
}

func (s *Service) DeleteCapability(ctx context.Context, capabilityID string) error {
	if err := s.repo.DeleteTemplate(ctx, strings.TrimSpace(capabilityID)); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrCapabilityNotFound
		}
		return err
	}
	return nil
}

func (s *Service) GetDeviceCapabilities(
	ctx context.Context,
	deviceID string,
) ([]domain.CapabilityUIModel, error) {
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

	items := make([]domain.CapabilityUIModel, 0, len(templates))
	for _, template := range templates {
		state := template.DefaultState
		enabled := true
		if saved, ok := states[template.ID]; ok {
			if strings.TrimSpace(saved.State) != "" {
				state = saved.State
			}
			enabled = saved.Enabled
		}
		items = append(items, domain.CapabilityUIModel{
			ID:          template.ID,
			Label:       template.Label,
			Description: template.Description,
			Control: domain.CapabilityControlDTO{
				Type:    template.Control.Type,
				Options: append([]domain.CapabilityControlOption(nil), template.Control.Options...),
			},
			State:   state,
			Enabled: enabled,
		})
	}
	sortCapabilityUIModels(items)
	return items, nil
}

func (s *Service) ListCapabilityAssignments(
	ctx context.Context,
	capabilityID string,
) ([]domain.CapabilityDeviceAssignment, error) {
	capabilityID = strings.TrimSpace(capabilityID)
	template, err := s.repo.GetTemplate(ctx, capabilityID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrCapabilityNotFound
	}
	if err != nil {
		return nil, err
	}

	devices, err := s.devices.ListDevices(ctx, presence.ListFilter{})
	if err != nil {
		return nil, err
	}
	states, err := s.repo.ListCapabilityDeviceStates(ctx, capabilityID)
	if err != nil {
		return nil, err
	}

	items := make([]domain.CapabilityDeviceAssignment, 0, len(devices))
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
		assignment := domain.CapabilityDeviceAssignment{
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

func (s *Service) PatchDeviceCapability(
	ctx context.Context,
	deviceID string,
	capabilityID string,
	state *string,
	enabled *bool,
) (domain.SetStateResult, error) {
	result := domain.SetStateResult{OK: true}
	if state == nil && enabled == nil {
		return domain.SetStateResult{}, fmt.Errorf("%w: either state or enabled must be provided", ErrCapabilityInvalid)
	}

	if state != nil {
		stateResult, err := s.SetCapabilityState(ctx, deviceID, capabilityID, *state)
		if err != nil {
			return domain.SetStateResult{}, err
		}
		result.Warnings = append(result.Warnings, stateResult.Warnings...)
	}

	if enabled != nil {
		if err := s.SetDeviceCapabilityEnabled(ctx, deviceID, capabilityID, *enabled); err != nil {
			return domain.SetStateResult{}, err
		}
	}

	return result, nil
}

func (s *Service) requireDevice(ctx context.Context, deviceID string) (model.DeviceView, error) {
	item, err := s.devices.GetDevice(ctx, deviceID)
	if errors.Is(err, storage.ErrNotFound) {
		return model.DeviceView{}, ErrDeviceNotFound
	}
	if err != nil {
		return model.DeviceView{}, err
	}
	return item, nil
}
