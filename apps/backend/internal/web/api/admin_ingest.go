package api

import (
	"github.com/borgehl/bruttice/apps/backend/internal/web/dto"
	"github.com/borgehl/bruttice/apps/backend/internal/web/mapper"
	"net/http"
	"strconv"
)

const adminIngestRunsPageSize = 15

func (h *APIHandler) handleAdminIngestRuns(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("GET /api/admin/ingest-runs called")

	if h.getAuthenticatedUser(r) == nil {
		h.writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// optional cursor: only runs older than this id
	var beforeID int64
	if raw := r.URL.Query().Get("before"); raw != "" {
		var err error
		beforeID, err = strconv.ParseInt(raw, 10, 64)
		if err != nil || beforeID < 0 {
			h.writeError(w, http.StatusBadRequest, "Invalid before")
			return
		}
	}

	// fetch one extra to know whether there is another page
	runs, err := h.database.GetRecentIngestRuns(adminIngestRunsPageSize+1, beforeID)
	if err != nil {
		h.logger.Errorf("failed to get ingest runs: %v", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to fetch ingest runs")
		return
	}

	hasMore := len(runs) > adminIngestRunsPageSize
	if hasMore {
		runs = runs[:adminIngestRunsPageSize]
	}

	h.writeJSON(w, http.StatusOK, dto.IngestRunsResponse{Runs: mapper.IngestRunResponses(runs), HasMore: hasMore})
}

// handleAdminStartIngest asks the scraper service to start a run.
func (h *APIHandler) handleAdminStartIngest(w http.ResponseWriter, r *http.Request) {
	h.logger.Infof("POST /api/admin/ingest-run called")

	user := h.getAuthenticatedUser(r)
	if user == nil {
		h.writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	if h.ingestTrigger == nil {
		h.writeError(w, http.StatusNotImplemented, "No scraper service is configured")
		return
	}

	started, err := h.ingestTrigger.TryRunAll()
	if err != nil {
		h.logger.Errorf("failed to start ingestion: %v", err)
		h.writeError(w, http.StatusBadGateway, "Failed to reach the scraper service")
		return
	}
	if !started {
		h.writeError(w, http.StatusConflict, "An ingestion is already running")
		return
	}

	h.logger.Infof("ingestion started by %s", user.Email)
	h.writeJSON(w, http.StatusAccepted, dto.StartIngestResponse{Success: true})
}
