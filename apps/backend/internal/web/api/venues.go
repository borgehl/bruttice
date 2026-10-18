package api

import (
	"net/http"

	"github.com/borgehl/bruttice/apps/backend/internal/web/dto"
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

	venues := make([]dto.VenueResponse, 0, len(venueModels))
	for _, v := range venueModels {
		venues = append(venues, dto.VenueResponse{Name: v.Name, Slug: v.Slug, Link: v.Link})
	}
	h.writeJSON(w, http.StatusOK, venues)
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

	venues := make([]dto.VenueResponse, 0, len(venueModels))
	for _, v := range venueModels {
		venues = append(venues, dto.VenueResponse{Name: v.Name, Slug: v.Slug, Link: v.Link})
	}
	h.writeJSON(w, http.StatusOK, venues)
}
