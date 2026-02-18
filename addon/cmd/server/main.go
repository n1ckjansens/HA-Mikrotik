package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/micro-ha/mikrotik-presence/addon/internal/aggregator"
	"github.com/micro-ha/mikrotik-presence/addon/internal/configsync"
	httpapi "github.com/micro-ha/mikrotik-presence/addon/internal/http"
	"github.com/micro-ha/mikrotik-presence/addon/internal/oui"
	"github.com/micro-ha/mikrotik-presence/addon/internal/poller"
	"github.com/micro-ha/mikrotik-presence/addon/internal/routeros"
	"github.com/micro-ha/mikrotik-presence/addon/internal/service"
	"github.com/micro-ha/mikrotik-presence/addon/internal/storage"
	"github.com/micro-ha/mikrotik-presence/addon/internal/subnet"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	dbPath := env("DB_PATH", "/data/mikrotik_presence.db")
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		logger.Error("failed to create db directory", "err", err)
		os.Exit(1)
	}

	repo, err := storage.New(ctx, dbPath, logger)
	if err != nil {
		logger.Error("failed to initialize storage", "err", err)
		os.Exit(1)
	}
	defer repo.Close()

	ouiDB, err := oui.LoadEmbedded()
	if err != nil {
		logger.Error("failed to load oui db", "err", err)
		os.Exit(1)
	}

	routerClient := routeros.NewClient()
	agg := aggregator.New(subnet.New(), ouiDB)

	haBaseURL := env("HA_BASE_URL", "http://supervisor/core")
	supervisorToken := os.Getenv("SUPERVISOR_TOKEN")
	cfgClient := configsync.NewClient(haBaseURL, supervisorToken)
	cfgManager := configsync.NewManager(cfgClient, logger)

	if _, err := cfgManager.Refresh(ctx); err != nil {
		logger.Warn("initial config refresh failed", "err", err)
	}

	svc := service.New(repo, agg, routerClient, cfgManager, logger)
	devicePoller := poller.New(svc, cfgManager, logger)

	go runConfigFallbackRefresh(ctx, cfgManager, devicePoller, logger)

	if supervisorToken != "" {
		watcher := configsync.NewWatcher(haBaseURL, supervisorToken, logger)
		go watcher.Run(ctx, func() {
			refreshCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			changed, err := cfgManager.Refresh(refreshCtx)
			if err != nil {
				logger.Warn("config refresh from event failed", "err", err)
				return
			}
			if changed {
				devicePoller.TriggerRefresh()
			}
		})
	} else {
		logger.Warn("SUPERVISOR_TOKEN is empty; config sync watcher disabled")
	}

	go devicePoller.Run(ctx)
	devicePoller.TriggerRefresh()

	api := httpapi.New(
		svc,
		devicePoller,
		cfgManager,
		logger,
		env("FRONTEND_DIST", "/app/frontend/dist"),
	)

	httpServer := &http.Server{
		Addr:              env("HTTP_ADDR", ":8099"),
		Handler:           api.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	logger.Info("server starting", "addr", httpServer.Addr)
	if err := httpapi.RunServer(ctx, httpServer, logger); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("server terminated with error", "err", err)
		os.Exit(1)
	}
	logger.Info("server stopped")
}

func runConfigFallbackRefresh(ctx context.Context, cfg *configsync.Manager, p *poller.Poller, logger *slog.Logger) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			refreshCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			changed, err := cfg.Refresh(refreshCtx)
			cancel()
			if err != nil {
				logger.Warn("periodic config refresh failed", "err", err)
				continue
			}
			if changed {
				p.TriggerRefresh()
			}
		}
	}
}

func env(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
