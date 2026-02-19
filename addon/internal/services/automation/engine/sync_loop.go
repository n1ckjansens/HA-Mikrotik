package engine

import (
	"context"
	"time"
)

const defaultSyncInterval = 20 * time.Second

// RunSyncLoop runs periodic synchronization until context cancellation.
func (e *Engine) RunSyncLoop(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = defaultSyncInterval
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := e.SyncOnce(ctx); err != nil && e.logger != nil {
				e.logger.Warn("automation sync failed", "err", err)
			}
		}
	}
}
