package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/micro-ha/mikrotik-presence/addon/internal/configsync"
	"github.com/micro-ha/mikrotik-presence/addon/internal/poller"
	"github.com/micro-ha/mikrotik-presence/addon/internal/service"
	"github.com/micro-ha/mikrotik-presence/addon/internal/storage"
)

type API struct {
	service   *service.Service
	poller    *poller.Poller
	config    *configsync.Manager
	logger    *slog.Logger
	staticDir string
}

func New(svc *service.Service, p *poller.Poller, cfg *configsync.Manager, logger *slog.Logger, staticDir string) *API {
	return &API{service: svc, poller: p, config: cfg, logger: logger, staticDir: staticDir}
}

func (a *API) Handler() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(20 * time.Second))
	r.Use(stripIngressPrefix)

	r.Get("/healthz", a.health)
	r.Route("/api", func(api chi.Router) {
		api.Get("/devices", a.listDevices)
		api.Get("/devices/{mac}", a.getDevice)
		api.Post("/devices/{mac}/register", a.registerDevice)
		api.Patch("/devices/{mac}", a.patchDevice)
		api.Post("/refresh", a.refresh)
	})

	r.Get("/*", a.static)
	r.Get("/", a.static)
	return r
}

func (a *API) health(w http.ResponseWriter, _ *http.Request) {
	_, configured := a.config.Get()
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "configured": configured})
}

func (a *API) listDevices(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.config.Get(); !ok {
		writeError(w, http.StatusConflict, "integration_not_configured", "Integration not configured")
		return
	}
	filter := service.ListFilter{Status: r.URL.Query().Get("status"), Query: r.URL.Query().Get("query")}
	if raw := strings.TrimSpace(r.URL.Query().Get("online")); raw != "" {
		value, err := strconv.ParseBool(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_online_filter", "online must be true or false")
			return
		}
		filter.Online = &value
	}
	items, err := a.service.ListDevices(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (a *API) getDevice(w http.ResponseWriter, r *http.Request) {
	device, err := a.service.GetDevice(r.Context(), chi.URLParam(r, "mac"))
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Device not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "get_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, device)
}

func (a *API) registerDevice(w http.ResponseWriter, r *http.Request) {
	var payload service.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_payload", "Invalid JSON payload")
		return
	}
	if err := a.service.RegisterDevice(r.Context(), chi.URLParam(r, "mac"), payload); err != nil {
		writeError(w, http.StatusInternalServerError, "register_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{"ok": true})
}

func (a *API) patchDevice(w http.ResponseWriter, r *http.Request) {
	var payload service.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_payload", "Invalid JSON payload")
		return
	}
	if err := a.service.PatchDevice(r.Context(), chi.URLParam(r, "mac"), payload); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Registered device not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "patch_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{"ok": true})
}

func (a *API) refresh(w http.ResponseWriter, _ *http.Request) {
	a.poller.TriggerRefresh()
	writeJSON(w, http.StatusAccepted, map[string]any{"ok": true})
}

func (a *API) static(w http.ResponseWriter, r *http.Request) {
	if a.staticDir == "" {
		writeError(w, http.StatusNotFound, "frontend_missing", "Frontend dist not found")
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		path = "index.html"
	}
	cleanPath := strings.TrimPrefix(filepath.Clean("/"+path), "/")
	fullPath := filepath.Join(a.staticDir, cleanPath)
	if info, err := os.Stat(fullPath); err == nil && !info.IsDir() {
		http.ServeFile(w, r, fullPath)
		return
	}
	http.ServeFile(w, r, filepath.Join(a.staticDir, "index.html"))
}

func stripIngressPrefix(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		prefix := strings.TrimSpace(r.Header.Get("X-Ingress-Path"))
		if prefix != "" && strings.HasPrefix(r.URL.Path, prefix) {
			r.URL.Path = strings.TrimPrefix(r.URL.Path, prefix)
			if r.URL.Path == "" {
				r.URL.Path = "/"
			}
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]any{
			"code":    code,
			"message": message,
		},
	})
}

func RunServer(ctx context.Context, server *http.Server, logger *slog.Logger) error {
	errCh := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case err := <-errCh:
		if err != nil {
			logger.Error("http server failed", "err", err)
			return err
		}
		return nil
	}
}
