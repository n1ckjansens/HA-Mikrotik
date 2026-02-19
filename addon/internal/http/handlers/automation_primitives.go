package handlers

import (
	"errors"
	"net/http"

	automationdomain "github.com/micro-ha/mikrotik-presence/addon/internal/domain/automation"
)

// ListActionTypes returns action metadata for frontend editors.
func (a *API) ListActionTypes(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, a.automation.ActionTypes())
}

// ListStateSourceTypes returns state-source metadata for frontend editors.
func (a *API) ListStateSourceTypes(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, a.automation.StateSourceTypes())
}

func writeAutomationServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, automationdomain.ErrCapabilityNotFound):
		writeError(w, http.StatusNotFound, "capability_not_found", err.Error())
	case errors.Is(err, automationdomain.ErrCapabilityConflict):
		writeError(w, http.StatusConflict, "capability_conflict", err.Error())
	case errors.Is(err, automationdomain.ErrCapabilityInvalid):
		writeError(w, http.StatusBadRequest, "capability_invalid", err.Error())
	case errors.Is(err, automationdomain.ErrCapabilityStateInvalid):
		writeError(w, http.StatusBadRequest, "capability_state_invalid", err.Error())
	case errors.Is(err, automationdomain.ErrDeviceNotFound):
		writeError(w, http.StatusNotFound, "device_not_found", err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "automation_failed", err.Error())
	}
}
