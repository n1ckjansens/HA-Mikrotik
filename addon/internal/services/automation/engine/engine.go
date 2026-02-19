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

// RouterConfigProvider exposes current add-on router config.
type RouterConfigProvider interface {
	Get() (model.RouterConfig, bool)
}

// RouterClient groups action/state-source dependencies.
type RouterClient interface {
	automationdomain.RouterActionClient
	automationdomain.RouterStateClient
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
	targetRef automationdomain.CapabilityTargetRef,
	capabilityID string,
	newState string,
) (automationdomain.SetStateResult, error) {
	targetRef, err := normalizeTargetRef(targetRef)
	if err != nil {
		return automationdomain.SetStateResult{}, err
	}

	capabilityID = strings.TrimSpace(capabilityID)
	newState = strings.TrimSpace(newState)
	if newState == "" {
		return automationdomain.SetStateResult{}, fmt.Errorf("%w: state is required", automationdomain.ErrCapabilityStateInvalid)
	}

	template, err := e.repo.GetTemplate(ctx, capabilityID)
	if errors.Is(err, automationdomain.ErrNotFound) {
		return automationdomain.SetStateResult{}, automationdomain.ErrCapabilityNotFound
	}
	if err != nil {
		return automationdomain.SetStateResult{}, err
	}
	template.Scope = automationdomain.NormalizeCapabilityScope(template.Scope)
	if template.Scope != targetRef.Scope {
		return automationdomain.SetStateResult{}, fmt.Errorf(
			"%w: template scope %q target scope %q",
			automationdomain.ErrCapabilityScopeMismatch,
			template.Scope,
			targetRef.Scope,
		)
	}

	stateConfig, ok := template.States[newState]
	if !ok {
		return automationdomain.SetStateResult{}, fmt.Errorf("%w: unknown state %q", automationdomain.ErrCapabilityStateInvalid, newState)
	}

	automationTarget, err := e.resolveAutomationTarget(ctx, targetRef)
	if err != nil {
		return automationdomain.SetStateResult{}, err
	}

	current, err := e.currentCapabilityState(ctx, targetRef, capabilityID, template.DefaultState)
	if err != nil {
		return automationdomain.SetStateResult{}, err
	}
	if current.Enabled && current.State == newState {
		return automationdomain.SetStateResult{OK: true}, nil
	}

	result := automationdomain.SetStateResult{OK: true}
	result.Warnings = append(result.Warnings, e.executeStateActions(
		ctx,
		automationTarget,
		capabilityID,
		newState,
		stateConfig.ActionsOnEnter,
	)...)

	current.Enabled = true
	current.State = newState
	if err := e.persistCapabilityState(ctx, targetRef, capabilityID, current); err != nil {
		return automationdomain.SetStateResult{}, err
	}
	return result, nil
}

// SyncOnce reads external state-sources and aligns capability states.
func (e *Engine) SyncOnce(ctx context.Context) error {
	routerConfig, configured := e.config.Get()
	if !configured {
		return automationdomain.ErrAddonNotConfigured
	}

	templates, err := e.repo.ListTemplates(ctx, "", "")
	if err != nil {
		return err
	}

	var syncErrors []error
	for _, template := range templates {
		template.Scope = automationdomain.NormalizeCapabilityScope(template.Scope)
		if template.Sync == nil || !template.Sync.Enabled {
			continue
		}

		source, ok := e.registry.StateSource(template.Sync.Source.TypeID)
		if !ok {
			syncErrors = append(syncErrors, fmt.Errorf("capability %s: statesource %q not found", template.ID, template.Sync.Source.TypeID))
			continue
		}

		targets, err := e.syncTargets(ctx, template.Scope)
		if err != nil {
			syncErrors = append(syncErrors, fmt.Errorf("capability %s: resolve targets: %w", template.ID, err))
			continue
		}

		for _, target := range targets {
			current, err := e.currentCapabilityState(ctx, target.Ref, template.ID, template.DefaultState)
			if err != nil {
				syncErrors = append(syncErrors, fmt.Errorf("capability %s target %s: current state: %w", template.ID, target.Label, err))
				continue
			}
			if !current.Enabled {
				continue
			}

			// `internal_truth` keeps local state as source of truth.
			if strings.EqualFold(strings.TrimSpace(template.Sync.Mode), "internal_truth") {
				continue
			}

			if err := source.Validate(target.Target, template.Sync.Source.Params); err != nil {
				syncErrors = append(syncErrors, fmt.Errorf("capability %s target %s: invalid sync source params: %w", template.ID, target.Label, err))
				continue
			}

			rawValue, err := source.Read(ctx, automationdomain.StateSourceContext{
				Target:       target.Target,
				RouterClient: e.routerClient,
				RouterConfig: routerConfig,
				Logger:       e.logger,
			}, template.Sync.Source.Params)
			if err != nil {
				syncErrors = append(syncErrors, fmt.Errorf("capability %s target %s: read sync source: %w", template.ID, target.Label, err))
				continue
			}

			boolValue, ok := rawValue.(bool)
			if !ok {
				syncErrors = append(syncErrors, fmt.Errorf("capability %s target %s: expected boolean source output", template.ID, target.Label))
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
				if err := e.persistCapabilityState(ctx, target.Ref, template.ID, current); err != nil {
					syncErrors = append(syncErrors, fmt.Errorf("capability %s target %s: upsert sync state: %w", template.ID, target.Label, err))
				}
				continue
			}

			if _, err := e.SetCapabilityState(ctx, target.Ref, template.ID, targetState); err != nil {
				syncErrors = append(syncErrors, fmt.Errorf("capability %s target %s: apply sync state: %w", template.ID, target.Label, err))
			}
		}
	}
	return errors.Join(syncErrors...)
}

type targetCapabilityState struct {
	Enabled bool
	State   string
}

type resolvedSyncTarget struct {
	Ref    automationdomain.CapabilityTargetRef
	Target automationdomain.AutomationTarget
	Label  string
}

func (e *Engine) syncTargets(
	ctx context.Context,
	scope automationdomain.CapabilityScope,
) ([]resolvedSyncTarget, error) {
	scope = automationdomain.NormalizeCapabilityScope(scope)
	if scope == automationdomain.ScopeGlobal {
		return []resolvedSyncTarget{{
			Ref:    automationdomain.CapabilityTargetRef{Scope: automationdomain.ScopeGlobal},
			Target: automationdomain.AutomationTarget{Scope: automationdomain.ScopeGlobal},
			Label:  "global",
		}}, nil
	}

	devices, err := e.devices.ListDevices(ctx, devicedomain.ListFilter{})
	if err != nil {
		return nil, err
	}
	items := make([]resolvedSyncTarget, 0, len(devices))
	for _, device := range devices {
		deviceCopy := device
		items = append(items, resolvedSyncTarget{
			Ref: automationdomain.CapabilityTargetRef{
				Scope:    automationdomain.ScopeDevice,
				DeviceID: normalizeDeviceID(device.MAC),
			},
			Target: automationdomain.AutomationTarget{
				Scope:  automationdomain.ScopeDevice,
				Device: &deviceCopy,
			},
			Label: normalizeDeviceID(device.MAC),
		})
	}
	return items, nil
}

func (e *Engine) currentCapabilityState(
	ctx context.Context,
	targetRef automationdomain.CapabilityTargetRef,
	capabilityID string,
	defaultState string,
) (targetCapabilityState, error) {
	defaultState = strings.TrimSpace(defaultState)
	targetRef.Scope = automationdomain.NormalizeCapabilityScope(targetRef.Scope)

	switch targetRef.Scope {
	case automationdomain.ScopeDevice:
		current, exists, err := e.repo.GetDeviceCapabilityState(ctx, targetRef.DeviceID, capabilityID)
		if err != nil {
			return targetCapabilityState{}, err
		}
		if !exists {
			return targetCapabilityState{Enabled: true, State: defaultState}, nil
		}
		state := strings.TrimSpace(current.State)
		if state == "" {
			state = defaultState
		}
		return targetCapabilityState{Enabled: current.Enabled, State: state}, nil
	case automationdomain.ScopeGlobal:
		current, err := e.repo.GetGlobalCapability(ctx, capabilityID)
		if err != nil {
			return targetCapabilityState{}, err
		}
		if current == nil {
			return targetCapabilityState{Enabled: true, State: defaultState}, nil
		}
		state := strings.TrimSpace(current.State)
		if state == "" {
			state = defaultState
		}
		return targetCapabilityState{Enabled: current.Enabled, State: state}, nil
	default:
		return targetCapabilityState{}, fmt.Errorf("%w: unsupported scope %q", automationdomain.ErrCapabilityScopeInvalid, targetRef.Scope)
	}
}

func (e *Engine) persistCapabilityState(
	ctx context.Context,
	targetRef automationdomain.CapabilityTargetRef,
	capabilityID string,
	state targetCapabilityState,
) error {
	targetRef.Scope = automationdomain.NormalizeCapabilityScope(targetRef.Scope)
	switch targetRef.Scope {
	case automationdomain.ScopeDevice:
		return e.repo.UpsertDeviceCapabilityState(ctx, automationdomain.DeviceCapability{
			DeviceID:     targetRef.DeviceID,
			CapabilityID: capabilityID,
			Enabled:      state.Enabled,
			State:        state.State,
			UpdatedAt:    time.Now().UTC(),
		})
	case automationdomain.ScopeGlobal:
		return e.repo.SaveGlobalCapability(ctx, &automationdomain.GlobalCapability{
			CapabilityID: capabilityID,
			Enabled:      state.Enabled,
			State:        state.State,
		})
	default:
		return fmt.Errorf("%w: unsupported scope %q", automationdomain.ErrCapabilityScopeInvalid, targetRef.Scope)
	}
}

func (e *Engine) resolveAutomationTarget(
	ctx context.Context,
	targetRef automationdomain.CapabilityTargetRef,
) (automationdomain.AutomationTarget, error) {
	targetRef.Scope = automationdomain.NormalizeCapabilityScope(targetRef.Scope)
	if targetRef.Scope == automationdomain.ScopeGlobal {
		return automationdomain.AutomationTarget{Scope: automationdomain.ScopeGlobal}, nil
	}

	device, err := e.requireDevice(ctx, targetRef.DeviceID)
	if err != nil {
		return automationdomain.AutomationTarget{}, err
	}
	deviceCopy := device
	return automationdomain.AutomationTarget{Scope: automationdomain.ScopeDevice, Device: &deviceCopy}, nil
}

func (e *Engine) executeStateActions(
	ctx context.Context,
	target automationdomain.AutomationTarget,
	capabilityID string,
	newState string,
	actions []automationdomain.ActionInstance,
) []automationdomain.ActionExecutionWarning {
	warnings := make([]automationdomain.ActionExecutionWarning, 0)
	routerConfig, configured := e.config.Get()

	for index, actionInstance := range actions {
		action, ok := e.registry.Action(actionInstance.TypeID)
		if !ok {
			warnings = append(warnings, warningForAction(
				actionInstance,
				fmt.Sprintf("action type %q is not registered", actionInstance.TypeID),
			))
			continue
		}
		if err := action.Validate(target, actionInstance.Params); err != nil {
			warnings = append(warnings, warningForAction(actionInstance, err.Error()))
			continue
		}
		if !configured {
			warnings = append(warnings, warningForAction(
				actionInstance,
				"router is not configured in add-on options",
			))
			continue
		}

		actionLogger := e.logger
		if actionLogger != nil {
			fields := []any{
				"scope", target.Scope,
				"capability_id", capabilityID,
				"state", newState,
				"action_type", actionInstance.TypeID,
				"action_index", index,
			}
			if target.Device != nil {
				fields = append(fields, "device_mac", target.Device.MAC)
			}
			actionLogger = actionLogger.With(fields...)
		}

		startedAt := time.Now()
		actionCtx, cancel := context.WithTimeout(ctx, actionExecutionTimeout)
		err := action.Execute(actionCtx, automationdomain.ActionExecutionContext{
			Target:       target,
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
			warnings = append(warnings, warningForAction(actionInstance, err.Error()))
			continue
		}
		if actionLogger != nil {
			actionLogger.Info("automation action succeeded", "duration_ms", duration.Milliseconds())
		}
	}

	return warnings
}

func normalizeTargetRef(
	target automationdomain.CapabilityTargetRef,
) (automationdomain.CapabilityTargetRef, error) {
	target.Scope = automationdomain.NormalizeCapabilityScope(target.Scope)
	if target.Scope == automationdomain.ScopeGlobal {
		target.DeviceID = ""
		return target, nil
	}

	target.DeviceID = normalizeDeviceID(target.DeviceID)
	if target.DeviceID == "" {
		return automationdomain.CapabilityTargetRef{}, fmt.Errorf("%w: device id is required", automationdomain.ErrCapabilityInvalid)
	}
	return target, nil
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
