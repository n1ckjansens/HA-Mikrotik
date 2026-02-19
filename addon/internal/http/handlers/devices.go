package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	devicedomain "github.com/micro-ha/mikrotik-presence/addon/internal/domain/device"
)

// ListDevices returns current device list.
func (a *API) ListDevices(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.config.Get(); !ok {
		writeError(w, http.StatusConflict, "integration_not_configured", "Integration not configured")
		return
	}
	filter := devicedomain.ListFilter{
		Status: r.URL.Query().Get("status"),
		Query:  r.URL.Query().Get("query"),
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("online")); raw != "" {
		value, err := strconv.ParseBool(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_online_filter", "online must be true or false")
			return
		}
		filter.Online = &value
	}

	items, err := a.devices.ListDevices(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

// GetDevice returns one device by MAC.
func (a *API) GetDevice(w http.ResponseWriter, r *http.Request, mac string) {
	device, err := a.devices.GetDevice(r.Context(), mac)
	if errors.Is(err, devicedomain.ErrDeviceNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "Device not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "get_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, device)
}

// RegisterDevice creates/updates registered device metadata.
func (a *API) RegisterDevice(w http.ResponseWriter, r *http.Request, mac string) {
	var payload devicedomain.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_payload", "Invalid JSON payload")
		return
	}
	if err := a.devices.RegisterDevice(r.Context(), mac, payload); err != nil {
		writeError(w, http.StatusInternalServerError, "register_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{"ok": true})
}

// PatchDevice partially updates registered device metadata.
func (a *API) PatchDevice(w http.ResponseWriter, r *http.Request, mac string) {
	var payload devicedomain.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_payload", "Invalid JSON payload")
		return
	}
	if err := a.devices.PatchDevice(r.Context(), mac, payload); err != nil {
		if errors.Is(err, devicedomain.ErrDeviceNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Registered device not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "patch_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{"ok": true})
}

// Refresh triggers immediate poll cycle asynchronously.
func (a *API) Refresh(w http.ResponseWriter, _ *http.Request) {
	a.poller.TriggerRefresh()
	writeJSON(w, http.StatusAccepted, map[string]any{"ok": true})
}
