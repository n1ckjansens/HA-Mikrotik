package automation

import (
	"fmt"
	"regexp"
	"strings"

	automationdomain "github.com/micro-ha/mikrotik-presence/addon/internal/domain/automation"
	"github.com/micro-ha/mikrotik-presence/addon/internal/services/automation/registry"
)

var capabilityIDPattern = regexp.MustCompile(`^[a-z0-9]+(\.[a-z0-9_]+)+$`)

func validateTemplate(template automationdomain.CapabilityTemplate, reg *registry.Registry) error {
	template.ID = strings.TrimSpace(template.ID)
	if template.ID == "" {
		return fmt.Errorf("id is required")
	}
	if !capabilityIDPattern.MatchString(template.ID) {
		return fmt.Errorf("id must match %s", capabilityIDPattern.String())
	}
	if strings.TrimSpace(template.Label) == "" {
		return fmt.Errorf("label is required")
	}
	if len(template.States) == 0 {
		return fmt.Errorf("states are required")
	}
	if strings.TrimSpace(template.DefaultState) == "" {
		return fmt.Errorf("default_state is required")
	}

	optionValues, err := validateControl(template.Control)
	if err != nil {
		return err
	}
	for _, optionValue := range optionValues {
		if _, ok := template.States[optionValue]; !ok {
			return fmt.Errorf("state %q is required for control option", optionValue)
		}
	}
	if _, ok := template.States[template.DefaultState]; !ok {
		return fmt.Errorf("default_state %q is not declared in states", template.DefaultState)
	}

	for stateID, state := range template.States {
		if strings.TrimSpace(stateID) == "" {
			return fmt.Errorf("state key cannot be empty")
		}
		for _, action := range state.ActionsOnEnter {
			actionType, ok := reg.Action(action.TypeID)
			if !ok {
				return fmt.Errorf("state %q has unknown action type %q", stateID, action.TypeID)
			}
			if err := actionType.Validate(action.Params); err != nil {
				return fmt.Errorf("state %q action %q: %w", stateID, action.TypeID, err)
			}
		}
	}

	if template.Sync != nil && template.Sync.Enabled {
		if strings.TrimSpace(template.Sync.Source.TypeID) == "" {
			return fmt.Errorf("sync.source.type_id is required when sync is enabled")
		}
		source, ok := reg.StateSource(template.Sync.Source.TypeID)
		if !ok {
			return fmt.Errorf("sync source type %q is not registered", template.Sync.Source.TypeID)
		}
		if err := source.Validate(template.Sync.Source.Params); err != nil {
			return fmt.Errorf("sync source params: %w", err)
		}
		if strings.TrimSpace(template.Sync.Mapping.WhenTrue) != "" {
			if _, ok := template.States[template.Sync.Mapping.WhenTrue]; !ok {
				return fmt.Errorf("sync.mapping.when_true %q is not declared in states", template.Sync.Mapping.WhenTrue)
			}
		}
		if strings.TrimSpace(template.Sync.Mapping.WhenFalse) != "" {
			if _, ok := template.States[template.Sync.Mapping.WhenFalse]; !ok {
				return fmt.Errorf("sync.mapping.when_false %q is not declared in states", template.Sync.Mapping.WhenFalse)
			}
		}
		mode := strings.TrimSpace(strings.ToLower(template.Sync.Mode))
		if mode != "" && mode != "external_truth" && mode != "internal_truth" {
			return fmt.Errorf("sync.mode must be external_truth or internal_truth")
		}
	}

	if template.HAExpose.Enabled {
		entityType := strings.TrimSpace(template.HAExpose.EntityType)
		if entityType != "switch" && entityType != "select" {
			return fmt.Errorf("ha_expose.entity_type must be switch or select")
		}
		if strings.TrimSpace(template.HAExpose.EntitySuffix) == "" {
			return fmt.Errorf("ha_expose.entity_suffix is required when HA expose is enabled")
		}
	}
	return nil
}

func normalizeTemplate(template automationdomain.CapabilityTemplate) automationdomain.CapabilityTemplate {
	template.ID = strings.TrimSpace(template.ID)
	template.Label = strings.TrimSpace(template.Label)
	template.Description = strings.TrimSpace(template.Description)
	template.Category = strings.TrimSpace(template.Category)
	template.DefaultState = strings.TrimSpace(template.DefaultState)
	if template.Control.Type == automationdomain.ControlSwitch && len(template.Control.Options) == 0 {
		template.Control.Options = []automationdomain.CapabilityControlOption{
			{Value: "on", Label: "On"},
			{Value: "off", Label: "Off"},
		}
	}
	if template.States == nil {
		template.States = map[string]automationdomain.CapabilityStateConfig{}
	}
	for stateID, state := range template.States {
		state.Label = strings.TrimSpace(state.Label)
		if state.ActionsOnEnter == nil {
			state.ActionsOnEnter = []automationdomain.ActionInstance{}
		}
		template.States[stateID] = state
	}
	if template.Sync != nil {
		template.Sync.Mode = normalizeSyncMode(template.Sync.Mode)
		if template.Sync.Source.Params == nil {
			template.Sync.Source.Params = map[string]any{}
		}
	}
	return template
}

func normalizeSyncMode(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "internal_truth":
		return "internal_truth"
	default:
		return "external_truth"
	}
}

func validateControl(control automationdomain.CapabilityControl) ([]string, error) {
	switch control.Type {
	case automationdomain.ControlSwitch:
		if len(control.Options) == 0 {
			return []string{"on", "off"}, nil
		}
		if len(control.Options) != 2 {
			return nil, fmt.Errorf("switch control must define exactly two options")
		}
		values := map[string]struct{}{}
		for _, option := range control.Options {
			values[option.Value] = struct{}{}
		}
		if _, ok := values["on"]; !ok {
			return nil, fmt.Errorf("switch control requires option value \"on\"")
		}
		if _, ok := values["off"]; !ok {
			return nil, fmt.Errorf("switch control requires option value \"off\"")
		}
		return []string{"on", "off"}, nil
	case automationdomain.ControlSelect:
		if len(control.Options) < 2 {
			return nil, fmt.Errorf("select control must define at least two options")
		}
		values := make([]string, 0, len(control.Options))
		seen := map[string]struct{}{}
		for _, option := range control.Options {
			value := strings.TrimSpace(option.Value)
			if value == "" {
				return nil, fmt.Errorf("select control option value cannot be empty")
			}
			if _, ok := seen[value]; ok {
				return nil, fmt.Errorf("duplicate control option value %q", value)
			}
			seen[value] = struct{}{}
			values = append(values, value)
		}
		return values, nil
	default:
		return nil, fmt.Errorf("unsupported control type %q", control.Type)
	}
}
