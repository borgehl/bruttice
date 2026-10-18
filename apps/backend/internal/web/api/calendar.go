package api

import (
	"github.com/borgehl/bruttice/apps/backend/internal/web/dto"
	"github.com/borgehl/bruttice/packages/go-pkg/models"
	"net/http"
	"time"
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

	responses := make([]dto.EventResponse, 0, len(events))
	for _, event := range events {
		responses = append(responses, h.eventToResponse(event))
	}
	h.writeJSON(w, http.StatusOK, responses)
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

func (h *APIHandler) eventToResponse(event models.Event) dto.EventResponse {
	resp := dto.EventResponse{
		Guid:      event.Guid,
		Title:     event.Title,
		Link:      event.Link,
		Venue:     event.Venue.Name,
		VenueLink: event.Venue.Link,
		Age:       event.Age.Name,
		Region:    event.Region.Name,
		StartAt:   event.StartAt.Format(time.RFC3339),
		CoverCost: event.CoverCost,
		Tags:      make([]string, 0, len(event.Tags)),
	}
	if event.DoorsAt != nil && *event.DoorsAt != event.StartAt {
		resp.DoorsAt = event.DoorsAt.UTC().Format(time.RFC3339)
	}
	for _, tag := range event.Tags {
		resp.Tags = append(resp.Tags, tag.Name)
	}
	return resp
}
