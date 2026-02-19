package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
)

type patchCapabilityPayload struct {
	State   *string `json:"state"`
	Enabled *bool   `json:"enabled"`
}

// ListDeviceCapabilities returns capabilities bound to one device.
func (a *API) ListDeviceCapabilities(w http.ResponseWriter, r *http.Request, mac string) {
	items, err := a.automation.GetDeviceCapabilities(r.Context(), mac)
	if err != nil {
		writeAutomationServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

// PatchDeviceCapability updates state/enabled for one device capability.
func (a *API) PatchDeviceCapability(
	w http.ResponseWriter,
	r *http.Request,
	mac string,
	capabilityID string,
) {
	payload, ok := decodePatchCapabilityPayload(w, r)
	if !ok {
		return
	}
	result, err := a.automation.PatchDeviceCapability(
		r.Context(),
		mac,
		capabilityID,
		payload.State,
		payload.Enabled,
	)
	if err != nil {
		writeAutomationServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// ListCapabilityDevices returns assignments for all devices for one capability.
func (a *API) ListCapabilityDevices(w http.ResponseWriter, r *http.Request, capabilityID string) {
	items, err := a.automation.ListCapabilityAssignments(r.Context(), capabilityID)
	if err != nil {
		writeAutomationServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

// PatchCapabilityDevice updates one device assignment via capability-centric route.
func (a *API) PatchCapabilityDevice(w http.ResponseWriter, r *http.Request, capabilityID string, mac string) {
	payload, ok := decodePatchCapabilityPayload(w, r)
	if !ok {
		return
	}
	result, err := a.automation.PatchDeviceCapability(
		r.Context(),
		mac,
		capabilityID,
		payload.State,
		payload.Enabled,
	)
	if err != nil {
		writeAutomationServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func decodePatchCapabilityPayload(
	w http.ResponseWriter,
	r *http.Request,
) (patchCapabilityPayload, bool) {
	var payload patchCapabilityPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_payload", "Invalid JSON payload")
		return patchCapabilityPayload{}, false
	}
	if payload.State != nil {
		trimmed := strings.TrimSpace(*payload.State)
		payload.State = &trimmed
	}
	if payload.State == nil && payload.Enabled == nil {
		writeError(w, http.StatusBadRequest, "invalid_payload", "Either state or enabled must be provided")
		return patchCapabilityPayload{}, false
	}
	return payload, true
}
