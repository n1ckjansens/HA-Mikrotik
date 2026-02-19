package engine

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	automationdomain "github.com/micro-ha/mikrotik-presence/addon/internal/domain/automation"
	devicedomain "github.com/micro-ha/mikrotik-presence/addon/internal/domain/device"
	"github.com/micro-ha/mikrotik-presence/addon/internal/model"
	"github.com/micro-ha/mikrotik-presence/addon/internal/services/automation/registry"
	"github.com/micro-ha/mikrotik-presence/addon/internal/storage"
)

const actionExecutionTimeout = 12 * time.Second

// DeviceService describes device read operations required by automation engine.
type DeviceService interface {
	GetDevice(ctx context.Context, mac string) (devicedomain.Device, error)
	ListDevices(ctx context.Context, filter devicedomain.ListFilter) ([]devicedomain.Device, error)
}

// RouterConfigProvider exposes current integration config.
type RouterConfigProvider interface {
	Get() (model.RouterConfig, bool)
}

// RouterClient groups action/state-source dependencies.
type RouterClient interface {
	automationdomain.AddressListClient
	automationdomain.AddressListStateClient
}

// Engine implements automation.AutomationEngine.
type Engine struct {
	repo         automationdomain.Repository
	devices      DeviceService
	registry     *registry.Registry
	config       RouterConfigProvider
	routerClient RouterClient
	logger       *slog.Logger
}

// New creates automation engine.
func New(
	repo automationdomain.Repository,
	devices DeviceService,
	registry *registry.Registry,
	config RouterConfigProvider,
	routerClient RouterClient,
	logger *slog.Logger,
) *Engine {
	return &Engine{
		repo:         repo,
		devices:      devices,
		registry:     registry,
		config:       config,
		routerClient: routerClient,
		logger:       logger,
	}
}

// SetCapabilityState executes actions and persists new state.
func (e *Engine) SetCapabilityState(
	ctx context.Context,
	deviceID string,
	capabilityID string,
	newState string,
) (automationdomain.SetStateResult, error) {
	deviceID = normalizeDeviceID(deviceID)
	capabilityID = strings.TrimSpace(capabilityID)
	newState = strings.TrimSpace(newState)
	if newState == "" {
		return automationdomain.SetStateResult{}, fmt.Errorf("%w: state is required", automationdomain.ErrCapabilityStateInvalid)
	}

	device, err := e.requireDevice(ctx, deviceID)
	if err != nil {
		return automationdomain.SetStateResult{}, err
	}
	template, err := e.repo.GetTemplate(ctx, capabilityID)
	if errors.Is(err, automationdomain.ErrNotFound) {
		return automationdomain.SetStateResult{}, automationdomain.ErrCapabilityNotFound
	}
	if err != nil {
		return automationdomain.SetStateResult{}, err
	}
	stateConfig, ok := template.States[newState]
	if !ok {
		return automationdomain.SetStateResult{}, fmt.Errorf("%w: unknown state %q", automationdomain.ErrCapabilityStateInvalid, newState)
	}

	current, exists, err := e.repo.GetDeviceCapabilityState(ctx, deviceID, capabilityID)
	if err != nil {
		return automationdomain.SetStateResult{}, err
	}
	if !exists {
		current = automationdomain.DeviceCapability{
			DeviceID:     deviceID,
			CapabilityID: capabilityID,
			Enabled:      true,
			State:        template.DefaultState,
		}
	}

	if current.Enabled && current.State == newState {
		return automationdomain.SetStateResult{OK: true}, nil
	}

	result := automationdomain.SetStateResult{OK: true}
	routerConfig, configured := e.config.Get()
	for index, actionInstance := range stateConfig.ActionsOnEnter {
		action, ok := e.registry.Action(actionInstance.TypeID)
		if !ok {
			result.Warnings = append(result.Warnings, warningForAction(
				actionInstance,
				fmt.Sprintf("action type %q is not registered", actionInstance.TypeID),
			))
			continue
		}
		if err := action.Validate(actionInstance.Params); err != nil {
			result.Warnings = append(result.Warnings, warningForAction(actionInstance, err.Error()))
			continue
		}
		if !configured {
			result.Warnings = append(result.Warnings, warningForAction(
				actionInstance,
				"router integration is not configured",
			))
			continue
		}

		actionLogger := e.logger
		if actionLogger != nil {
			actionLogger = actionLogger.With(
				"device_mac", device.MAC,
				"capability_id", capabilityID,
				"state", newState,
				"action_type", actionInstance.TypeID,
				"action_index", index,
			)
		}

		startedAt := time.Now()
		actionCtx, cancel := context.WithTimeout(ctx, actionExecutionTimeout)
		err := action.Execute(actionCtx, automationdomain.ActionExecutionContext{
			Device:       device,
			RouterClient: e.routerClient,
			RouterConfig: routerConfig,
			Logger:       actionLogger,
		}, actionInstance.Params)
		cancel()

		duration := time.Since(startedAt)
		if err != nil {
			if actionLogger != nil {
				actionLogger.Warn("automation action failed", "duration_ms", duration.Milliseconds(), "err", err)
			}
			result.Warnings = append(result.Warnings, warningForAction(actionInstance, err.Error()))
			continue
		}
		if actionLogger != nil {
			actionLogger.Info("automation action succeeded", "duration_ms", duration.Milliseconds())
		}
	}

	current.State = newState
	current.Enabled = true
	current.UpdatedAt = time.Now().UTC()
	if err := e.repo.UpsertDeviceCapabilityState(ctx, current); err != nil {
		return automationdomain.SetStateResult{}, err
	}
	return result, nil
}

// SyncOnce reads external state-sources and aligns device capability states.
func (e *Engine) SyncOnce(ctx context.Context) error {
	routerConfig, configured := e.config.Get()
	if !configured {
		return automationdomain.ErrIntegrationNotConfigured
	}

	templates, err := e.repo.ListTemplates(ctx, "", "")
	if err != nil {
		return err
	}
	devices, err := e.devices.ListDevices(ctx, devicedomain.ListFilter{})
	if err != nil {
		return err
	}

	var syncErrors []error
	for _, template := range templates {
		if template.Sync == nil || !template.Sync.Enabled {
			continue
		}
		source, ok := e.registry.StateSource(template.Sync.Source.TypeID)
		if !ok {
			syncErrors = append(syncErrors, fmt.Errorf("capability %s: statesource %q not found", template.ID, template.Sync.Source.TypeID))
			continue
		}
		if err := source.Validate(template.Sync.Source.Params); err != nil {
			syncErrors = append(syncErrors, fmt.Errorf("capability %s: invalid sync source params: %w", template.ID, err))
			continue
		}
		states, err := e.repo.ListCapabilityDeviceStates(ctx, template.ID)
		if err != nil {
			syncErrors = append(syncErrors, fmt.Errorf("capability %s: list states: %w", template.ID, err))
			continue
		}

		for _, device := range devices {
			current := currentStateOrDefault(states, device.MAC, template)
			if !current.Enabled {
				continue
			}

			// `internal_truth` keeps local state as source of truth.
			if strings.EqualFold(strings.TrimSpace(template.Sync.Mode), "internal_truth") {
				continue
			}

			rawValue, err := source.Read(ctx, automationdomain.StateSourceContext{
				Device:       device,
				RouterClient: e.routerClient,
				RouterConfig: routerConfig,
				Logger:       e.logger,
			}, template.Sync.Source.Params)
			if err != nil {
				syncErrors = append(syncErrors, fmt.Errorf("capability %s device %s: read sync source: %w", template.ID, device.MAC, err))
				continue
			}

			boolValue, ok := rawValue.(bool)
			if !ok {
				syncErrors = append(syncErrors, fmt.Errorf("capability %s device %s: expected boolean source output", template.ID, device.MAC))
				continue
			}

			targetState := template.Sync.Mapping.WhenFalse
			if boolValue {
				targetState = template.Sync.Mapping.WhenTrue
			}
			targetState = strings.TrimSpace(targetState)
			if targetState == "" || targetState == current.State {
				continue
			}

			if !template.Sync.TriggerActionsOnSync {
				current.State = targetState
				current.UpdatedAt = time.Now().UTC()
				if err := e.repo.UpsertDeviceCapabilityState(ctx, current); err != nil {
					syncErrors = append(syncErrors, fmt.Errorf("capability %s device %s: upsert sync state: %w", template.ID, device.MAC, err))
				}
				continue
			}

			if _, err := e.SetCapabilityState(ctx, device.MAC, template.ID, targetState); err != nil {
				syncErrors = append(syncErrors, fmt.Errorf("capability %s device %s: apply sync state: %w", template.ID, device.MAC, err))
			}
		}
	}
	return errors.Join(syncErrors...)
}

func currentStateOrDefault(
	states map[string]automationdomain.DeviceCapability,
	deviceID string,
	template automationdomain.CapabilityTemplate,
) automationdomain.DeviceCapability {
	deviceID = normalizeDeviceID(deviceID)
	if existing, ok := states[deviceID]; ok {
		if strings.TrimSpace(existing.State) == "" {
			existing.State = template.DefaultState
		}
		return existing
	}
	return automationdomain.DeviceCapability{
		DeviceID:     deviceID,
		CapabilityID: template.ID,
		Enabled:      true,
		State:        template.DefaultState,
	}
}

func (e *Engine) requireDevice(ctx context.Context, deviceID string) (devicedomain.Device, error) {
	item, err := e.devices.GetDevice(ctx, deviceID)
	if errors.Is(err, devicedomain.ErrDeviceNotFound) || errors.Is(err, storage.ErrNotFound) {
		return devicedomain.Device{}, automationdomain.ErrDeviceNotFound
	}
	if err != nil {
		return devicedomain.Device{}, err
	}
	return item, nil
}

func warningForAction(
	action automationdomain.ActionInstance,
	message string,
) automationdomain.ActionExecutionWarning {
	return automationdomain.ActionExecutionWarning{
		ActionID: action.ID,
		TypeID:   action.TypeID,
		Message:  message,
	}
}

func normalizeDeviceID(raw string) string {
	value := strings.TrimSpace(strings.ToUpper(raw))
	value = strings.ReplaceAll(value, "%3A", ":")
	value = strings.ReplaceAll(value, "%3a", ":")
	value = strings.ReplaceAll(value, "-", ":")
	return value
}
