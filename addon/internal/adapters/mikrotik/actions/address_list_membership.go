package actions

import (
	"context"
	"fmt"
	"strings"

	automationdomain "github.com/micro-ha/mikrotik-presence/addon/internal/domain/automation"
)

const (
	// ActionIDAddressListMembership toggles membership for one address-list entry.
	ActionIDAddressListMembership = "mikrotik.address_list.set_membership"
)

// AddressListMembershipAction toggles RouterOS address-list membership.
type AddressListMembershipAction struct{}

// NewAddressListMembershipAction creates MikroTik membership action.
func NewAddressListMembershipAction() *AddressListMembershipAction {
	return &AddressListMembershipAction{}
}

// ID returns unique action identifier.
func (a *AddressListMembershipAction) ID() string {
	return ActionIDAddressListMembership
}

// Metadata returns action descriptor for UI.
func (a *AddressListMembershipAction) Metadata() automationdomain.ActionMetadata {
	return automationdomain.ActionMetadata{
		ID:          ActionIDAddressListMembership,
		Label:       "MikroTik: Address-list membership",
		Description: "Add or remove a target value in a MikroTik firewall address-list",
		ParamSchema: []automationdomain.ParamField{
			{
				Key:         "list",
				Label:       "Address-list name",
				Kind:        automationdomain.ParamString,
				Required:    true,
				Description: "RouterOS firewall address-list name",
			},
			{
				Key:         "mode",
				Label:       "Mode",
				Kind:        automationdomain.ParamEnum,
				Required:    true,
				Options:     []string{"add", "remove"},
				Description: "Whether to add or remove target from the list",
			},
			{
				Key:         "target",
				Label:       "Target",
				Kind:        automationdomain.ParamEnum,
				Required:    true,
				Options:     []string{"device.ip", "device.mac", "literal_ip"},
				Description: "Source value to apply in address-list",
			},
			{
				Key:       "literal_ip",
				Label:     "Literal IP",
				Kind:      automationdomain.ParamString,
				Required:  true,
				VisibleIf: &automationdomain.VisibleIfCondition{Key: "target", Equals: "literal_ip"},
			},
		},
	}
}

// Validate validates action params against metadata schema.
func (a *AddressListMembershipAction) Validate(
	target automationdomain.AutomationTarget,
	params map[string]any,
) error {
	list, err := stringParam(params, "list")
	if err != nil {
		return err
	}
	if strings.TrimSpace(list) == "" {
		return fmt.Errorf("list is required")
	}

	mode, err := stringParam(params, "mode")
	if err != nil {
		return err
	}
	if mode != "add" && mode != "remove" {
		return fmt.Errorf("unsupported mode %q", mode)
	}

	targetParam, err := stringParam(params, "target")
	if err != nil {
		return err
	}
	scope := automationdomain.NormalizeCapabilityScope(target.Scope)
	switch targetParam {
	case "device.ip", "device.mac":
		if scope == automationdomain.ScopeGlobal {
			return fmt.Errorf("target %q is not available for global scope", targetParam)
		}
	case "literal_ip":
		literalIP, err := stringParam(params, "literal_ip")
		if err != nil {
			return err
		}
		if scope == automationdomain.ScopeGlobal && containsDevicePlaceholder(literalIP) {
			return fmt.Errorf("global scope does not support device placeholders")
		}
	default:
		return fmt.Errorf("unsupported target %q", targetParam)
	}
	return nil
}

// Execute applies add/remove action to RouterOS address-list.
func (a *AddressListMembershipAction) Execute(
	ctx context.Context,
	execCtx automationdomain.ActionExecutionContext,
	params map[string]any,
) error {
	if err := a.Validate(execCtx.Target, params); err != nil {
		return err
	}
	if execCtx.RouterClient == nil {
		return fmt.Errorf("router client is not configured")
	}

	listName, _ := stringParam(params, "list")
	mode, _ := stringParam(params, "mode")
	target, _ := stringParam(params, "target")

	address, err := resolveTargetAddress(target, params, execCtx)
	if err != nil {
		return err
	}
	switch mode {
	case "add":
		return execCtx.RouterClient.AddAddressListEntry(ctx, execCtx.RouterConfig, listName, address)
	case "remove":
		return execCtx.RouterClient.RemoveAddressListEntry(ctx, execCtx.RouterConfig, listName, address)
	default:
		return fmt.Errorf("unsupported mode %q", mode)
	}
}

func resolveTargetAddress(
	target string,
	params map[string]any,
	execCtx automationdomain.ActionExecutionContext,
) (string, error) {
	switch target {
	case "device.ip":
		if execCtx.Target.Device == nil || execCtx.Target.Device.LastIP == nil || strings.TrimSpace(*execCtx.Target.Device.LastIP) == "" {
			return "", fmt.Errorf("device IP is empty")
		}
		return strings.TrimSpace(*execCtx.Target.Device.LastIP), nil
	case "device.mac":
		if execCtx.Target.Device == nil || strings.TrimSpace(execCtx.Target.Device.MAC) == "" {
			return "", fmt.Errorf("device MAC is empty")
		}
		return execCtx.Target.Device.MAC, nil
	case "literal_ip":
		value, err := stringParam(params, "literal_ip")
		if err != nil {
			return "", err
		}
		return value, nil
	default:
		return "", fmt.Errorf("unsupported target %q", target)
	}
}

func containsDevicePlaceholder(raw string) bool {
	return strings.Contains(strings.ToLower(raw), "{{device.")
}

func stringParam(params map[string]any, key string) (string, error) {
	raw, ok := params[key]
	if !ok {
		return "", fmt.Errorf("missing param %q", key)
	}
	value, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("param %q must be string", key)
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("param %q is empty", key)
	}
	return value, nil
}
