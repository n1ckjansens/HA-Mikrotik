package statesources

import (
	"context"
	"fmt"
	"strings"

	automationdomain "github.com/micro-ha/mikrotik-presence/addon/internal/domain/automation"
)

const (
	// StateSourceIDAddressListMembership reads membership in RouterOS address-list.
	StateSourceIDAddressListMembership = "mikrotik.address_list.membership"
)

// AddressListMembershipSource checks if target is currently in address-list.
type AddressListMembershipSource struct{}

// NewAddressListMembershipSource creates state-source implementation.
func NewAddressListMembershipSource() *AddressListMembershipSource {
	return &AddressListMembershipSource{}
}

// ID returns unique state-source identifier.
func (s *AddressListMembershipSource) ID() string {
	return StateSourceIDAddressListMembership
}

// Metadata returns state-source descriptor for UI.
func (s *AddressListMembershipSource) Metadata() automationdomain.StateSourceMetadata {
	return automationdomain.StateSourceMetadata{
		ID:          StateSourceIDAddressListMembership,
		Label:       "MikroTik: Address-list membership",
		Description: "Checks whether target value currently exists in MikroTik firewall address-list",
		OutputType:  "boolean",
		ParamSchema: []automationdomain.ParamField{
			{
				Key:         "list",
				Label:       "Address-list name",
				Kind:        automationdomain.ParamString,
				Required:    true,
				Description: "RouterOS firewall address-list name",
			},
			{
				Key:         "target",
				Label:       "Target",
				Kind:        automationdomain.ParamEnum,
				Required:    true,
				Options:     []string{"device.ip", "device.mac", "literal_ip"},
				Description: "Source value used to lookup in address-list",
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

// Validate validates state-source params against schema.
func (s *AddressListMembershipSource) Validate(
	target automationdomain.AutomationTarget,
	params map[string]any,
) error {
	if _, err := stringParam(params, "list"); err != nil {
		return err
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

// Read checks presence of target value in RouterOS address-list.
func (s *AddressListMembershipSource) Read(
	ctx context.Context,
	sourceCtx automationdomain.StateSourceContext,
	params map[string]any,
) (any, error) {
	if err := s.Validate(sourceCtx.Target, params); err != nil {
		return nil, err
	}
	if sourceCtx.RouterClient == nil {
		return nil, fmt.Errorf("router client is not configured")
	}

	listName, _ := stringParam(params, "list")
	target, _ := stringParam(params, "target")
	address, err := resolveTargetAddress(target, params, sourceCtx)
	if err != nil {
		return nil, err
	}

	contains, err := sourceCtx.RouterClient.AddressListContains(
		ctx,
		sourceCtx.RouterConfig,
		listName,
		address,
	)
	if err != nil {
		return nil, err
	}
	return contains, nil
}

func resolveTargetAddress(
	target string,
	params map[string]any,
	sourceCtx automationdomain.StateSourceContext,
) (string, error) {
	switch target {
	case "device.ip":
		if sourceCtx.Target.Device == nil || sourceCtx.Target.Device.LastIP == nil || strings.TrimSpace(*sourceCtx.Target.Device.LastIP) == "" {
			return "", fmt.Errorf("device IP is empty")
		}
		return strings.TrimSpace(*sourceCtx.Target.Device.LastIP), nil
	case "device.mac":
		if sourceCtx.Target.Device == nil || strings.TrimSpace(sourceCtx.Target.Device.MAC) == "" {
			return "", fmt.Errorf("device MAC is empty")
		}
		return sourceCtx.Target.Device.MAC, nil
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
