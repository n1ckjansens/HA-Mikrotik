package configsync

import (
	"context"
	"log/slog"
	"sync"

	"github.com/micro-ha/mikrotik-presence/addon/internal/model"
)

type Manager struct {
	client *Client
	logger *slog.Logger

	mu         sync.RWMutex
	configured bool
	config     model.RouterConfig
}

func NewManager(client *Client, logger *slog.Logger) *Manager {
	return &Manager{client: client, logger: logger}
}

func (m *Manager) Refresh(ctx context.Context) (bool, error) {
	res, err := m.client.FetchConfig(ctx)
	if err != nil {
		return false, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	changed := false
	if !res.Configured {
		if m.configured {
			changed = true
		}
		m.configured = false
		m.config = model.RouterConfig{}
		return changed, nil
	}

	if !m.configured || res.Config.Version != m.config.Version {
		changed = true
	}
	m.configured = true
	m.config = res.Config
	return changed, nil
}

func (m *Manager) Get() (model.RouterConfig, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if !m.configured {
		return model.RouterConfig{}, false
	}
	return m.config, true
}
