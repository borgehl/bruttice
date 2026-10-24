package api

import (
	"fmt"
	"github.com/borgehl/bruttice/apps/backend/internal/web/dto"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/emailx"
	"net/http"
	"time"
)

// handleStartIngestRun opens a run that a scraper service can attach its payloads to.
func (h *APIHandler) handleStartIngestRun(w http.ResponseWriter, r *http.Request) {
	if !h.ingestAuthorized(r) {
		h.writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	run, err := h.database.CreateIngestRun()
	if err != nil {
		h.logger.Errorf("failed to create ingest run record: %s", err)
		h.writeError(w, http.StatusInternalServerError, "failed to create run")
		return
	}
	h.writeJSON(w, http.StatusCreated, dto.IngestRunCreatedResponse{ID: run.ID})
}

// handleFinishIngestRun marks a run as finished once the scraper service has posted all its payloads.
func (h *APIHandler) handleFinishIngestRun(w http.ResponseWriter, r *http.Request) {
	if !h.ingestAuthorized(r) {
		h.writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var runID int64
	if _, err := fmt.Sscanf(r.PathValue("id"), "%d", &runID); err != nil || runID <= 0 {
		h.writeError(w, http.StatusBadRequest, "invalid run id")
		return
	}

	if err := h.database.FinishIngestRun(runID, nil); err != nil {
		h.logger.Errorf("failed to finish ingest run record: %s", err)
		h.writeError(w, http.StatusInternalServerError, "failed to finish run")
		return
	}
	h.emailRunErrors(runID, true)
	w.WriteHeader(http.StatusNoContent)
}

// emailRunErrors emails the root admin the errors a run recorded.
//
// With weeklyCheck, a run with no errors is still reported on Mondays, so a
// silent inbox means ingestion is broken rather than healthy.
func (h *APIHandler) emailRunErrors(runID int64, weeklyCheck bool) {
	run, err := h.database.GetIngestRun(runID)
	if err != nil {
		h.logger.Errorf("failed to load ingest run %d for its error email: %s", runID, err)
		return
	}

	report := emailx.ErrorsReport{RunID: runID, DedupeError: run.DedupeError}
	for _, source := range run.Sources {
		report.Sources = append(report.Sources, emailx.SourceReport{
			Name:    source.Source,
			Scraped: source.ScrapedCount,
			Saved:   source.SavedCount,
			Failed:  source.FailedCount,
			Errors:  source.Errors,
		})
	}

	if report.Total() == 0 && report.DedupeError == "" && !(weeklyCheck && time.Now().Weekday() == time.Monday) {
		h.logger.Infof("no email: no errors in ingest run %d", runID)
		return
	}

	// Send says if it did, or why not
	_, _ = h.email.Send(h.email.ErrorsMessage(report))
}
