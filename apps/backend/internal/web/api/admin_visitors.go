package api

import (
	"net/http"
	"strconv"

	"github.com/borgehl/bruttice/apps/backend/internal/web/dto"
)

const (
	defaultVisitorStatsDays = 30
	maxVisitorStatsDays     = 365
	maxVisitorLocations     = 50
)

// handleAdminVisitors returns how many distinct visitors came each day, and from where.
// Only visitors who accepted the cookie banner are counted.
func (h *APIHandler) handleAdminVisitors(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("GET /api/admin/visitors called")

	if h.getAuthenticatedUser(r) == nil {
		h.writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	days := defaultVisitorStatsDays
	if raw := r.URL.Query().Get("days"); raw != "" {
		var err error
		days, err = strconv.Atoi(raw)
		if err != nil || days < 1 || days > maxVisitorStatsDays {
			h.writeError(w, http.StatusBadRequest, "Invalid days")
			return
		}
	}

	daily, err := h.database.GetVisitorsPerDay(days)
	if err != nil {
		h.logger.Errorf("failed to get visitors per day: %v", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to fetch visitors")
		return
	}

	locations, err := h.database.GetVisitorsPerLocation(days, maxVisitorLocations)
	if err != nil {
		h.logger.Errorf("failed to get visitors per location: %v", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to fetch visitors")
		return
	}

	resp := dto.VisitorStatsResponse{
		Days:      days,
		Daily:     make([]dto.VisitorDayResponse, 0, len(daily)),
		Locations: make([]dto.VisitorLocationResponse, 0, len(locations)),
	}
	for _, d := range daily {
		resp.Daily = append(resp.Daily, dto.VisitorDayResponse{
			Date:     d.VisitDate.Format("2006-01-02"),
			Visitors: d.NumVisitors,
		})
	}
	for _, l := range locations {
		resp.Locations = append(resp.Locations, dto.VisitorLocationResponse{
			Country:  l.Country,
			State:    l.State,
			City:     l.City,
			Visitors: l.NumVisitors,
		})
	}

	h.writeJSON(w, http.StatusOK, resp)
}
