package poller

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/micro-ha/mikrotik-presence/addon/internal/configsync"
	"github.com/micro-ha/mikrotik-presence/addon/internal/service"
)

type Poller struct {
	service   *service.Service
	config    *configsync.Manager
	refreshCh chan struct{}
	logger    *slog.Logger
}

func New(svc *service.Service, cfg *configsync.Manager, logger *slog.Logger) *Poller {
	return &Poller{service: svc, config: cfg, refreshCh: make(chan struct{}, 1), logger: logger}
}

func (p *Poller) TriggerRefresh() {
	select {
	case p.refreshCh <- struct{}{}:
	default:
	}
}

func (p *Poller) Run(ctx context.Context) {
	for {
		interval := 5 * time.Second
		if cfg, ok := p.config.Get(); ok {
			interval = cfg.PollInterval()
		}
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-p.refreshCh:
			timer.Stop()
		case <-timer.C:
		}
		if err := p.service.PollOnce(ctx); err != nil {
			if errors.Is(err, service.ErrIntegrationNotConfigured) {
				p.logger.Info("poll skipped; integration not configured")
				continue
			}
			p.logger.Error("poll failed", "err", err)
		}
	}
}
