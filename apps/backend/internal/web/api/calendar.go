package api

import (
	"github.com/borgehl/bruttice/apps/backend/internal/web/mapper"
	"github.com/borgehl/bruttice/packages/go-pkg/models"
	"net/http"
)

func (h *APIHandler) handleGetCalendar(w http.ResponseWriter, r *http.Request) {
	tx := h.database.DB.Begin()
	defer tx.Commit()

	filter, err := h.getFilter(r)
	if err != nil {
		h.logger.Errorf("failed to get filter: %v", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to get filter")
		return
	}

	events, err := h.database.GetIndexEvents(filter, tx)
	if err != nil {
		h.logger.Errorf("failed to get events: %v", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to get events")
		return
	}

	h.writeJSON(w, http.StatusOK, mapper.EventResponses(events))
}

func (h *APIHandler) getFilter(r *http.Request) (models.CalendarFilterBackend, error) {
	var ret models.CalendarFilterBackend
	filterKey := r.URL.Query().Get("filterKey")

	filterModel, err := h.database.GetFilterByKey(filterKey, h.database.DB)
	if err != nil {
		return ret, err
	}
	filter := filterModel.GetBackendFilter()

	filter.CustomOffsetsToDates(
		r.URL.Query().Get("from-date"),
		r.URL.Query().Get("to-date"),
	)
	return filter, nil
}
