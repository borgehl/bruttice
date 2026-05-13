package api

import (
	"github.com/borgehl/bruttice/packages/go-pkg/configs"
	"github.com/borgehl/bruttice/packages/go-pkg/ingest"
	"github.com/borgehl/bruttice/packages/go-pkg/models"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/util"
	"encoding/json"
	"html"
	"net/http"
	"strings"

	"gorm.io/gorm"
)

func (h *APIHandler) handleIngest(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") || strings.TrimPrefix(auth, "Bearer ") != h.ingestKey {
		h.writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var payload ingest.IngestionOutput
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	for _, e := range payload.Errors {
		h.logger.Errorf("scraper error from %s: %s", payload.VenueInfo.Name, e)
	}

	tx := h.database.DB.Begin()

	if err := h.saveVenueInfo(payload.VenueInfo, tx); err != nil {
		tx.Rollback()
		h.logger.Errorf("failed to save venue %s: %s", payload.VenueInfo.Name, err)
		h.writeError(w, http.StatusInternalServerError, "failed to save venue")
		return
	}

	saved := 0
	var errs []string
	for _, event := range payload.Events {
		if err := h.saveIngestedEvent(event, payload.VenueInfo.IsAutoIngest, tx); err != nil {
			h.logger.Errorf("failed to save event %s: %s", event.Title, err)
			errs = append(errs, err.Error())
		} else {
			saved++
		}
	}

	tx.Commit()

	dedupTx := h.database.DB.Begin()
	if err := h.database.DeduplicateEvents(dedupTx); err != nil {
		h.logger.Errorf("deduplication failed: %s", err)
		dedupTx.Rollback()
	} else {
		dedupTx.Commit()
	}

	h.logger.Infof("ingest from %s: saved %d events, %d errors", payload.VenueInfo.Name, saved, len(errs))
	h.writeJSON(w, http.StatusOK, map[string]any{"saved": saved, "errors": errs})
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
		panic(err)
	}
	if eventUnscoped.ID != 0 && eventUnscoped.DeletedAt.Valid {
		return nil
	}

	event.Title = html.UnescapeString(event.Title)
	_, err = h.database.SaveEvent(event, venueIsAutoIngest, tx)
	return err
}
