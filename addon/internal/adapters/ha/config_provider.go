package ha

import (
	"context"

	"github.com/micro-ha/mikrotik-presence/addon/internal/configsync"
	"github.com/micro-ha/mikrotik-presence/addon/internal/model"
)

// ConfigProvider exposes current integration config from HA supervisor.
type ConfigProvider interface {
	Get() (model.RouterConfig, bool)
	Refresh(ctx context.Context) (bool, error)
}

// ManagerAdapter adapts configsync manager to adapter contract.
type ManagerAdapter struct {
	manager *configsync.Manager
}

// NewManagerAdapter wraps config manager for DI in new services.
func NewManagerAdapter(manager *configsync.Manager) *ManagerAdapter {
	return &ManagerAdapter{manager: manager}
}

// Get returns cached config and configured flag.
func (a *ManagerAdapter) Get() (model.RouterConfig, bool) {
	return a.manager.Get()
}

// Refresh fetches latest config snapshot from HA.
func (a *ManagerAdapter) Refresh(ctx context.Context) (bool, error) {
	return a.manager.Refresh(ctx)
}
