package api

import (
	"crypto/subtle"
	"github.com/borgehl/bruttice/apps/backend/internal/web/dto"
	"github.com/borgehl/bruttice/apps/backend/internal/web/mapper"
	"github.com/borgehl/bruttice/packages/go-pkg/configs"
	"github.com/borgehl/bruttice/packages/go-pkg/ingest"
	"github.com/borgehl/bruttice/packages/go-pkg/models"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/util"
	"html"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"
)

// maxIngestBody is the largest ingest payload accepted.
const maxIngestBody = 16 << 20

// handleIngest saves one scraper's output for one venue.
//
// Payloads that carry a run_id are recorded under that run (see handleStartIngestRun).
// Payloads without one are recorded as a run of their own.
func (h *APIHandler) handleIngest(w http.ResponseWriter, r *http.Request) {
	if !h.ingestAuthorized(r) {
		h.writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var payload ingest.IngestionOutput
	// strict, so a payload in an older format is refused instead of ingested with fields missing
	if !h.decodeStrictJSON(w, r, &payload, maxIngestBody) {
		return
	}

	for _, e := range payload.Errors {
		h.logger.Errorf("scraper error from %s: %s", payload.VenueInfo.Name, e)
	}

	runID := payload.RunID
	standalone := runID == 0
	if standalone {
		run, err := h.database.CreateIngestRun()
		if err != nil {
			h.logger.Errorf("failed to create ingest run record: %s", err)
		}
		runID = run.ID
	}

	source, dedupeErr := h.ingestPayload(payload)
	h.recordIngestSource(runID, source)

	if dedupeErr != nil && runID != 0 {
		if err := h.database.SetIngestRunDedupeError(runID, dedupeErr); err != nil {
			h.logger.Errorf("failed to record dedupe error: %s", err)
		}
	}
	if standalone && runID != 0 {
		if err := h.database.FinishIngestRun(runID, dedupeErr); err != nil {
			h.logger.Errorf("failed to finish ingest run record: %s", err)
		}
		h.emailRunErrors(runID, false)
	}

	status := http.StatusOK
	if source.Status == models.IngestSourceStatusFailed {
		status = http.StatusInternalServerError
	}
	h.logger.Infof("ingest from %s: saved %d of %d events, %d errors", payload.VenueInfo.Name, source.SavedCount, source.ScrapedCount, len(source.Errors))
	h.writeJSON(w, status, dto.IngestResultResponse{Saved: source.SavedCount, Errors: source.Errors})
}

// ingestPayload saves the venue and events of a payload, then deduplicates that
// venue's events. The returned error is a deduplication failure, which does not
// affect the saved events.
func (h *APIHandler) ingestPayload(payload ingest.IngestionOutput) (models.IngestRunSource, error) {
	source := models.IngestRunSource{
		Source:       payload.VenueInfo.Name,
		Status:       models.IngestSourceStatusOk,
		ScrapedCount: len(payload.Events),
		Errors:       []string{},
	}

	tx := h.database.DB.Begin()

	if err := h.saveVenueInfo(payload.VenueInfo, tx); err != nil {
		tx.Rollback()
		h.logger.Errorf("failed to save venue %s: %s", payload.VenueInfo.Name, err)
		source.Status = models.IngestSourceStatusFailed
		source.Errors = []string{err.Error()}
		return source, nil
	}

	// a scraper that reported errors may have returned partial data. Saving it is
	// safe: guids and deduplication make ingestion idempotent, and events missing
	// from a partial payload are not removed
	source.Errors = append(source.Errors, payload.Errors...)
	if len(payload.Errors) > 0 {
		source.Status = models.IngestSourceStatusErrors
	}

	now := time.Now()
	for _, scraped := range payload.Events {
		event := mapper.IngestEvent(scraped, payload.VenueInfo, now)
		// each event gets a savepoint so one bad event does not abort the rest
		err := tx.Transaction(func(eventTx *gorm.DB) error {
			return h.saveIngestedEvent(event, payload.VenueInfo.IsAutoIngest, eventTx)
		})
		if err != nil {
			h.logger.Errorf("failed to save event %s: %s", event.Title, err)
			source.FailedCount++
			source.Errors = append(source.Errors, err.Error())
			continue
		}
		source.SavedCount++
	}
	tx.Commit()

	if source.FailedCount > 0 {
		source.Status = models.IngestSourceStatusErrors
	}

	dedupTx := h.database.DB.Begin()
	if err := h.database.DeduplicateEvents(h.ingestedVenueIDs(payload, dedupTx), dedupTx); err != nil {
		h.logger.Errorf("deduplication failed: %s", err)
		dedupTx.Rollback()
		return source, err
	}
	dedupTx.Commit()
	return source, nil
}

func (h *APIHandler) recordIngestSource(runID int64, source models.IngestRunSource) {
	if runID == 0 {
		return
	}
	source.IngestRunID = runID
	if err := h.database.SaveIngestRunSource(source); err != nil {
		h.logger.Errorf("failed to save ingest run source %s: %s", source.Source, err)
	}
}

func (h *APIHandler) ingestAuthorized(r *http.Request) bool {
	auth := r.Header.Get("Authorization")
	key := strings.TrimPrefix(auth, "Bearer ")
	return h.ingestKey != "" && strings.HasPrefix(auth, "Bearer ") && subtle.ConstantTimeCompare([]byte(key), []byte(h.ingestKey)) == 1
}

// ingestedVenueIDs returns the IDs of the venues touched by a payload: the
// payload's own venue plus any venue named on its events.
func (h *APIHandler) ingestedVenueIDs(payload ingest.IngestionOutput, tx *gorm.DB) []int64 {
	names := map[string]struct{}{}
	if payload.VenueInfo.Name != "" {
		names[payload.VenueInfo.Name] = struct{}{}
	}
	for _, event := range payload.Events {
		if event.Venue != "" {
			names[event.Venue] = struct{}{}
		}
	}

	ids := make([]int64, 0, len(names))
	for name := range names {
		venue, err := h.database.GetVenueByName(name, tx)
		if err != nil {
			h.logger.Errorf("failed to look up venue %s for dedup: %s", name, err)
			continue
		}
		if venue.ID != 0 {
			ids = append(ids, int64(venue.ID))
		}
	}
	return ids
}

func (h *APIHandler) saveVenueInfo(venueInfo ingest.VenueInfo, tx *gorm.DB) error {
	if venueInfo.Name == "" || venueInfo.SkipVenueSave {
		return nil
	}
	venue := models.Venue{
		Name:          venueInfo.Name,
		Link:          venueInfo.Link,
		IsScraped:     venueInfo.IsAutoIngest,
		DefaultRegion: venueInfo.DefaultRegion,
	}
	venue.InternalUpdate()
	_, err := h.database.GetSaveVenueByName(venue, tx)
	return err
}

func (h *APIHandler) saveIngestedEvent(event models.Event, venueIsAutoIngest bool, tx *gorm.DB) error {
	titleLower := strings.ToLower(event.Title)
	if strings.Contains(titleLower, "private party") || strings.Contains(titleLower, "private event") {
		return nil
	}

	event.UpdateGuid(nil)
	if util.IsFree(event.CoverCost) {
		event.Tags = append(event.Tags, models.Tag{Name: configs.FreeTag})
	}

	eventUnscoped, err := h.database.GetEventByGuidUnscoped(event.Guid, tx)
	if err != nil {
		return err
	}
	if eventUnscoped.ID != 0 && eventUnscoped.DeletedAt.Valid {
		return nil
	}

	event.Title = html.UnescapeString(event.Title)
	_, err = h.database.SaveEvent(event, venueIsAutoIngest, tx)
	return err
}
