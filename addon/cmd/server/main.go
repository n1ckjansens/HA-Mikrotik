package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/micro-ha/mikrotik-presence/addon/internal/adapters/ha"
	mikrotikadapter "github.com/micro-ha/mikrotik-presence/addon/internal/adapters/mikrotik"
	mikrotikactions "github.com/micro-ha/mikrotik-presence/addon/internal/adapters/mikrotik/actions"
	mikrotikstatesources "github.com/micro-ha/mikrotik-presence/addon/internal/adapters/mikrotik/statesources"
	"github.com/micro-ha/mikrotik-presence/addon/internal/aggregator"
	"github.com/micro-ha/mikrotik-presence/addon/internal/config"
	"github.com/micro-ha/mikrotik-presence/addon/internal/configsync"
	"github.com/micro-ha/mikrotik-presence/addon/internal/http"
	"github.com/micro-ha/mikrotik-presence/addon/internal/http/handlers"
	"github.com/micro-ha/mikrotik-presence/addon/internal/logging"
	"github.com/micro-ha/mikrotik-presence/addon/internal/oui"
	"github.com/micro-ha/mikrotik-presence/addon/internal/poller"
	"github.com/micro-ha/mikrotik-presence/addon/internal/repository/sqlite"
	automationservice "github.com/micro-ha/mikrotik-presence/addon/internal/services/automation"
	automationengine "github.com/micro-ha/mikrotik-presence/addon/internal/services/automation/engine"
	automationregistry "github.com/micro-ha/mikrotik-presence/addon/internal/services/automation/registry"
	deviceservice "github.com/micro-ha/mikrotik-presence/addon/internal/services/device"
	"github.com/micro-ha/mikrotik-presence/addon/internal/subnet"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()
	logger := logging.New(cfg.LogLevel)

	if err := os.MkdirAll(cfg.DBDir(), 0o755); err != nil {
		logger.Error("failed to create db directory", "err", err)
		os.Exit(1)
	}

	db, err := sqlite.Open(ctx, cfg.DBPath, logger)
	if err != nil {
		logger.Error("failed to initialize storage", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	ouiDB, err := oui.LoadEmbedded()
	if err != nil {
		logger.Error("failed to load oui db", "err", err)
		os.Exit(1)
	}

	routerClient := mikrotikadapter.NewRestClient(logger.With("component", "routeros"))
	agg := aggregator.NewWithThresholds(subnet.New(), ouiDB, cfg.PresenceThresholds)

	cfgClient := configsync.NewClient(cfg.HABaseURL, cfg.SupervisorToken)
	cfgManager := configsync.NewManager(cfgClient, logger.With("component", "configsync"))
	haConfig := ha.NewManagerAdapter(cfgManager)

	if _, err := cfgManager.Refresh(ctx); err != nil {
		logger.Warn("initial config refresh failed", "err", err)
	}

	deviceRepo := sqlite.NewDeviceRepository(db)
	automationRepo := sqlite.NewAutomationRepository(db)

	deviceSvc := deviceservice.NewWithThresholds(
		deviceRepo,
		agg,
		routerClient,
		haConfig,
		logger.With("service", "device"),
		cfg.PresenceThresholds,
	)

	reg := automationregistry.New()
	reg.RegisterAction(mikrotikactions.NewAddressListMembershipAction())
	reg.RegisterStateSource(mikrotikstatesources.NewAddressListMembershipSource())

	engine := automationengine.New(
		automationRepo,
		deviceSvc,
		reg,
		haConfig,
		routerClient,
		logger.With("service", "automation_engine"),
	)
	automationSvc := automationservice.New(
		automationRepo,
		deviceSvc,
		engine,
		reg,
		logger.With("service", "automation"),
	)

	devicePoller := poller.New(deviceSvc, cfgManager, logger.With("component", "poller"))
	go runConfigFallbackRefresh(ctx, cfgManager, devicePoller, logger)
	go devicePoller.Run(ctx)
	devicePoller.TriggerRefresh()

	go engine.RunSyncLoop(ctx, cfg.AutomationSyncInterval)

	if cfg.SupervisorToken != "" {
		watcher := configsync.NewWatcher(cfg.HABaseURL, cfg.SupervisorToken, logger.With("component", "configsync-watcher"))
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

	api := handlers.New(
		deviceSvc,
		automationSvc,
		devicePoller,
		haConfig,
		logger.With("component", "http"),
		cfg.FrontendDist,
	)

	httpServer := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           httpapi.NewRouter(api),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	logger.Info("server starting", "addr", httpServer.Addr)
	if err := httpapi.RunServer(ctx, httpServer); err != nil && !errors.Is(err, context.Canceled) {
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
