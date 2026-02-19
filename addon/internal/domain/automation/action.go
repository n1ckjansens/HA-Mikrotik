package automation

import (
	"context"
	"log/slog"

	devicedomain "github.com/micro-ha/mikrotik-presence/addon/internal/domain/device"
	"github.com/micro-ha/mikrotik-presence/addon/internal/model"
)

// AddressListClient is required by MikroTik address-list related actions.
type AddressListClient interface {
	AddAddressListEntry(ctx context.Context, cfg model.RouterConfig, list, address string) error
	RemoveAddressListEntry(ctx context.Context, cfg model.RouterConfig, list, address string) error
}

// ActionExecutionContext contains runtime dependencies for action execution.
type ActionExecutionContext struct {
	Device       devicedomain.Device
	RouterClient AddressListClient
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
	Validate(params map[string]any) error
	Execute(ctx context.Context, execCtx ActionExecutionContext, params map[string]any) error
}
