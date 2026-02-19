package actions

import (
	"context"
	"log/slog"

	"github.com/micro-ha/mikrotik-presence/addon/internal/model"
)

type AddressListClient interface {
	AddAddressListEntry(ctx context.Context, cfg model.RouterConfig, list, address string) error
	RemoveAddressListEntry(ctx context.Context, cfg model.RouterConfig, list, address string) error
}

type ExecutionContext struct {
	Device       model.DeviceView
	RouterClient AddressListClient
	RouterConfig model.RouterConfig
	Logger       *slog.Logger
}

type Handler func(ctx context.Context, execCtx ExecutionContext, params map[string]any) error
