package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/micro-ha/mikrotik-presence/addon/internal/automation/actions"
	"github.com/micro-ha/mikrotik-presence/addon/internal/automation/domain"
	"github.com/micro-ha/mikrotik-presence/addon/internal/automation/repository"
)

const actionExecutionTimeout = 12 * time.Second

func (s *Service) SetCapabilityState(
	ctx context.Context,
	deviceID string,
	capabilityID string,
	newState string,
) (domain.SetStateResult, error) {
	deviceID = normalizeDeviceID(deviceID)
	capabilityID = strings.TrimSpace(capabilityID)
	newState = strings.TrimSpace(newState)
	if newState == "" {
		return domain.SetStateResult{}, fmt.Errorf("%w: state is required", ErrCapabilityStateInvalid)
	}

	device, err := s.requireDevice(ctx, deviceID)
	if err != nil {
		return domain.SetStateResult{}, err
	}
	template, err := s.repo.GetTemplate(ctx, capabilityID)
	if errors.Is(err, repository.ErrNotFound) {
		return domain.SetStateResult{}, ErrCapabilityNotFound
	}
	if err != nil {
		return domain.SetStateResult{}, err
	}
	stateConfig, ok := template.States[newState]
	if !ok {
		return domain.SetStateResult{}, fmt.Errorf("%w: unknown state %q", ErrCapabilityStateInvalid, newState)
	}

	current, exists, err := s.repo.GetDeviceCapabilityState(ctx, deviceID, capabilityID)
	if err != nil {
		return domain.SetStateResult{}, err
	}
	if !exists {
		current = domain.DeviceCapabilityState{
			DeviceID:     deviceID,
			CapabilityID: capabilityID,
			Enabled:      true,
			State:        template.DefaultState,
		}
	}

	if current.Enabled && current.State == newState {
		return domain.SetStateResult{OK: true}, nil
	}

	result := domain.SetStateResult{OK: true}
	routerConfig, configured := s.config.Get()

	for index, actionInstance := range stateConfig.ActionsOnEnter {
		actionType, ok := actions.TypeByID(actionInstance.TypeID)
		if !ok {
			result.Warnings = append(result.Warnings, warningForAction(
				actionInstance,
				fmt.Sprintf("action type %q is not registered", actionInstance.TypeID),
			))
			continue
		}

		resolvedParams, err := actions.ResolveParams(actionInstance.Params, device)
		if err != nil {
			result.Warnings = append(result.Warnings, warningForAction(actionInstance, err.Error()))
			continue
		}
		if err := actions.ValidateActionParams(actionType, resolvedParams); err != nil {
			result.Warnings = append(result.Warnings, warningForAction(actionInstance, err.Error()))
			continue
		}
		handler, ok := actions.HandlerByID(actionInstance.TypeID)
		if !ok {
			result.Warnings = append(result.Warnings, warningForAction(
				actionInstance,
				fmt.Sprintf("handler is not registered for action type %q", actionInstance.TypeID),
			))
			continue
		}
		if !configured {
			result.Warnings = append(result.Warnings, warningForAction(
				actionInstance,
				"router integration is not configured",
			))
			continue
		}

		actionLogger := s.logger
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
		err = handler(actionCtx, actions.ExecutionContext{
			Device:       device,
			RouterClient: s.router,
			RouterConfig: routerConfig,
			Logger:       actionLogger,
		}, resolvedParams)
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
	if err := s.repo.UpsertDeviceCapabilityState(ctx, current); err != nil {
		return domain.SetStateResult{}, err
	}
	return result, nil
}

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
	if errors.Is(err, repository.ErrNotFound) {
		return ErrCapabilityNotFound
	}
	if err != nil {
		return err
	}

	current, exists, err := s.repo.GetDeviceCapabilityState(ctx, deviceID, capabilityID)
	if err != nil {
		return err
	}
	if !exists {
		current = domain.DeviceCapabilityState{
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

func warningForAction(action domain.ActionInstance, message string) domain.ActionExecutionWarning {
	return domain.ActionExecutionWarning{
		ActionID: action.ID,
		TypeID:   action.TypeID,
		Message:  message,
	}
}
