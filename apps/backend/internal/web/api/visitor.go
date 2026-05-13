package api

import (
	"github.com/borgehl/bruttice/packages/go-pkg/models"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/util"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
)

type visitorConsentRequest struct {
	Path  string `json:"path"`
	Query string `json:"query"`
}

func (h *APIHandler) handleVisitorConsent(w http.ResponseWriter, r *http.Request) {
	var req visitorConsentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Errorf("visitor consent: failed to decode request body: %v", err)
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	visitorID := uuid.New().String()

	visitorIP := util.GetClientIPAddr(r)
	location, err := h.geoClient.GetLocation(visitorIP)
	if err != nil {
		h.logger.Errorf("visitor consent: geo lookup on %s failed: %v", visitorIP, err)
	}

	loc := models.VisitorLocation{
		VisitorID: visitorID,
		Country:   location.Country,
		StateCode: location.StateCode,
		City:      location.City,
	}

	pageQuery, _ := url.ParseQuery(req.Query)
	accessLog := models.VisitorLog{
		VisitorID:   visitorID,
		Hostname:    r.Host,
		PathName:    req.Path,
		QueryParams: req.Query,
		FilterKey:   pageQuery.Get("filterKey"),
	}

	tx := h.database.DB.Begin()
	if err := h.database.SaveVisitorLog(accessLog, loc, tx); err != nil {
		tx.Rollback()
		h.logger.Errorf("visitor consent: failed to save visitor log: %v", err)
		h.writeError(w, http.StatusInternalServerError, "failed to record consent")
		return
	}
	tx.Commit()

	http.SetCookie(w, &http.Cookie{
		Name:     "visitor-id",
		Value:    visitorID,
		Path:     "/",
		MaxAge:   int(365 * 24 * time.Hour / time.Second),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	h.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
