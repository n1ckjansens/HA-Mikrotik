package handlers

import (
	"encoding/json"
	"net/http"

	automationdomain "github.com/micro-ha/mikrotik-presence/addon/internal/domain/automation"
)

// ListCapabilities returns capability templates with optional filters.
func (a *API) ListCapabilities(w http.ResponseWriter, r *http.Request) {
	items, err := a.automation.ListCapabilities(
		r.Context(),
		r.URL.Query().Get("search"),
		r.URL.Query().Get("category"),
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "capability_list_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, items)
}

// GetCapability returns capability template by ID.
func (a *API) GetCapability(w http.ResponseWriter, r *http.Request, capabilityID string) {
	item, err := a.automation.GetCapability(r.Context(), capabilityID)
	if err != nil {
		writeAutomationServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

// CreateCapability validates and creates capability template.
func (a *API) CreateCapability(w http.ResponseWriter, r *http.Request) {
	var template automationdomain.CapabilityTemplate
	if err := json.NewDecoder(r.Body).Decode(&template); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_payload", "Invalid JSON payload")
		return
	}
	if err := a.automation.CreateCapability(r.Context(), template); err != nil {
		writeAutomationServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, template)
}

// UpdateCapability validates and updates capability template.
func (a *API) UpdateCapability(w http.ResponseWriter, r *http.Request, capabilityID string) {
	var template automationdomain.CapabilityTemplate
	if err := json.NewDecoder(r.Body).Decode(&template); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_payload", "Invalid JSON payload")
		return
	}
	if err := a.automation.UpdateCapability(r.Context(), capabilityID, template); err != nil {
		writeAutomationServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, template)
}

// DeleteCapability removes capability template by ID.
func (a *API) DeleteCapability(w http.ResponseWriter, r *http.Request, capabilityID string) {
	if err := a.automation.DeleteCapability(r.Context(), capabilityID); err != nil {
		writeAutomationServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
