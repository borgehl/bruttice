package api

import (
	"net/http"
	"strconv"
)

const adminIngestRunsPageSize = 15

type IngestSourceResponse struct {
	Source       string   `json:"source"`
	Status       string   `json:"status"`
	ScrapedCount int      `json:"scrapedCount"`
	SavedCount   int      `json:"savedCount"`
	FailedCount  int      `json:"failedCount"`
	Errors       []string `json:"errors"`
}

type IngestRunsResponse struct {
	Runs    []IngestRunResponse `json:"runs"`
	HasMore bool                `json:"hasMore"`
}

type IngestRunResponse struct {
	ID          int64                  `json:"id"`
	StartedAt   int64                  `json:"startedAt"`
	FinishedAt  int64                  `json:"finishedAt"`
	DedupeError string                 `json:"dedupeError,omitempty"`
	Scraped     int                    `json:"scraped"`
	Saved       int                    `json:"saved"`
	Failed      int                    `json:"failed"`
	Sources     []IngestSourceResponse `json:"sources"`
}

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

	ret := make([]IngestRunResponse, 0, len(runs))
	for _, run := range runs {
		resp := IngestRunResponse{
			ID:          run.ID,
			StartedAt:   run.StartedAt,
			FinishedAt:  run.FinishedAt,
			DedupeError: run.DedupeError,
			Sources:     make([]IngestSourceResponse, 0, len(run.Sources)),
		}
		for _, s := range run.Sources {
			resp.Scraped += s.ScrapedCount
			resp.Saved += s.SavedCount
			resp.Failed += s.FailedCount
			errs := s.Errors
			if errs == nil {
				errs = []string{}
			}
			resp.Sources = append(resp.Sources, IngestSourceResponse{
				Source:       s.Source,
				Status:       s.Status,
				ScrapedCount: s.ScrapedCount,
				SavedCount:   s.SavedCount,
				FailedCount:  s.FailedCount,
				Errors:       errs,
			})
		}
		ret = append(ret, resp)
	}

	h.writeJSON(w, http.StatusOK, IngestRunsResponse{Runs: ret, HasMore: hasMore})
}

type StartIngestResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
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
	h.writeJSON(w, http.StatusAccepted, StartIngestResponse{Success: true})
}
