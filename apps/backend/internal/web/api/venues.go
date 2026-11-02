package api

import (
	"net/http"

	"github.com/borgehl/bruttice/apps/backend/internal/web/mapper"
)

func (h *APIHandler) handleGetVenues(w http.ResponseWriter, r *http.Request) {
	tx := h.database.DB.Begin()
	defer tx.Commit()

	// Get venues
	venueModels, err := h.database.GetVenues(tx)
	if err != nil {
		h.logger.Errorf("failed to get venues: %v", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to get venues")
		return
	}

	h.writeJSON(w, http.StatusOK, mapper.VenueResponses(venueModels))
}

func (h *APIHandler) handleGetSubmissionVenues(w http.ResponseWriter, r *http.Request) {
	tx := h.database.DB.Begin()
	defer tx.Commit()

	venueModels, err := h.database.GetSubmissionVenues(tx)
	if err != nil {
		h.logger.Errorf("failed to get submission venues: %v", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to get venues")
		return
	}

	h.writeJSON(w, http.StatusOK, mapper.VenueResponses(venueModels))
}
