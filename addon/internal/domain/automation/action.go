package automation

import (
	"context"
	"log/slog"

	devicedomain "github.com/micro-ha/mikrotik-presence/addon/internal/domain/device"
	"github.com/micro-ha/mikrotik-presence/addon/internal/model"
)

// AutomationTarget identifies runtime scope for action/state-source execution.
type AutomationTarget struct {
	Scope  CapabilityScope
	Device *devicedomain.Device
}

// AddressListClient is required by MikroTik address-list related actions.
type AddressListClient interface {
	AddAddressListEntry(ctx context.Context, cfg model.RouterConfig, list, address string) error
	RemoveAddressListEntry(ctx context.Context, cfg model.RouterConfig, list, address string) error
}

// FirewallRuleClient is required by MikroTik firewall rule actions.
type FirewallRuleClient interface {
	SetFirewallRuleDisabled(ctx context.Context, cfg model.RouterConfig, table, ruleID string, disabled bool) error
	SetFirewallRulesDisabledByComment(ctx context.Context, cfg model.RouterConfig, table, comment string, disabled bool) error
}

// RouterActionClient groups RouterOS operations used by automation actions.
type RouterActionClient interface {
	AddressListClient
	FirewallRuleClient
}

// ActionExecutionContext contains runtime dependencies for action execution.
type ActionExecutionContext struct {
	Target       AutomationTarget
	RouterClient RouterActionClient
	RouterConfig model.RouterConfig
	Logger       *slog.Logger
}

// ActionMetadata describes an action type and its parameters.
type ActionMetadata struct {
	ID          string       `json:"id"`
	Label       string       `json:"label"`
	Description string       `json:"description"`
	ParamSchema []ParamField `json:"param_schema"`
}

// Action is pluggable behavior executed on capability state transitions.
type Action interface {
	ID() string
	Metadata() ActionMetadata
	Validate(target AutomationTarget, params map[string]any) error
	Execute(ctx context.Context, execCtx ActionExecutionContext, params map[string]any) error
}
