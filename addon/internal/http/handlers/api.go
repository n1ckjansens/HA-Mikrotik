package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	automationdomain "github.com/micro-ha/mikrotik-presence/addon/internal/domain/automation"
	devicedomain "github.com/micro-ha/mikrotik-presence/addon/internal/domain/device"
	"github.com/micro-ha/mikrotik-presence/addon/internal/model"
)

// Poller triggers asynchronous presence refresh.
type Poller interface {
	TriggerRefresh()
}

// ConfigProvider exposes current add-on router config status.
type ConfigProvider interface {
	Get() (model.RouterConfig, bool)
}

// API groups HTTP handlers and dependencies.
type API struct {
	devices    devicedomain.Service
	automation automationdomain.Service
	poller     Poller
	config     ConfigProvider
	logger     *slog.Logger
	staticDir  string
}

// New creates HTTP handlers with explicit dependencies.
func New(
	devices devicedomain.Service,
	automation automationdomain.Service,
	poller Poller,
	config ConfigProvider,
	logger *slog.Logger,
	staticDir string,
) *API {
	return &API{
		devices:    devices,
		automation: automation,
		poller:     poller,
		config:     config,
		logger:     logger,
		staticDir:  staticDir,
	}
}

// Logger returns request logger used by HTTP middleware.
func (a *API) Logger() *slog.Logger {
	return a.logger
}

// Health reports service liveness and router config status.
func (a *API) Health(w http.ResponseWriter, _ *http.Request) {
	_, configured := a.config.Get()
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "configured": configured})
}

// Static serves frontend assets and SPA fallback.
func (a *API) Static(w http.ResponseWriter, r *http.Request) {
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

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]any{
			"code":    code,
			"message": message,
		},
	})
}
