package automation

import (
	"context"
	"log/slog"

	devicedomain "github.com/micro-ha/mikrotik-presence/addon/internal/domain/device"
	"github.com/micro-ha/mikrotik-presence/addon/internal/model"
)

// AddressListStateClient provides read operations for RouterOS address-lists.
type AddressListStateClient interface {
	AddressListContains(ctx context.Context, cfg model.RouterConfig, list, address string) (bool, error)
}

// StateSourceContext contains runtime dependencies for state reads.
type StateSourceContext struct {
	Device       devicedomain.Device
	RouterClient AddressListStateClient
	RouterConfig model.RouterConfig
	Logger       *slog.Logger
}

// StateSourceMetadata describes external state provider for sync.
type StateSourceMetadata struct {
	ID          string       `json:"id"`
	Label       string       `json:"label"`
	Description string       `json:"description"`
	OutputType  string       `json:"output_type"`
	ParamSchema []ParamField `json:"param_schema"`
}

// StateSource reads external truth for one capability/device pair.
type StateSource interface {
	ID() string
	Metadata() StateSourceMetadata
	Validate(params map[string]any) error
	Read(ctx context.Context, sourceCtx StateSourceContext, params map[string]any) (any, error)
}
