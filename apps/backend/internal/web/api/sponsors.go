package api

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/borgehl/bruttice/apps/backend/internal/web/dto"
	"github.com/borgehl/bruttice/apps/backend/internal/web/mapper"
	"github.com/borgehl/bruttice/packages/go-pkg/models"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/dbx"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/util"
)

// sponsorStore is what sponsor cards need from the database. Its errors are *dbx.ConfigError when
// the admin asked for something that can't be done or the sponsor does not exist.
type sponsorStore interface {
	ListSponsorRows(today string) ([]dbx.SponsorRow, error)
	ListSponsors() ([]models.Sponsor, error)
	GetSponsor(id int64) (models.Sponsor, error)
	SponsorImage(id int64) ([]byte, string, error)
	CreateSponsor(in dbx.SponsorInput) error
	UpdateSponsor(id int64, in dbx.SponsorInput) error
	DeleteSponsor(id int64) error
	CountSponsor(id int64, kind dbx.SponsorCount, today string) error
}

// newSponsorLimiter is for the counters, which anyone can call: 120 a visitor an hour, 20000 in all.
// Past it the card still works, it is only not counted.
func newSponsorLimiter() *rateLimiter {
	return newRateLimiter(time.Hour, 120, 20000)
}

// sponsorToday is today's date in the site's time zone, which is how sponsors' dates are read.
func sponsorToday() string {
	return time.Now().In(util.TZLocation).Format(models.SponsorDateLayout)
}

// sponsorCountAllowed says if a request may be counted. It records the request.
func (h *APIHandler) sponsorCountAllowed(r *http.Request) bool {
	return h.sponsorLimit == nil || h.sponsorLimit.Allow(util.GetClientIPAddr(r), time.Now())
}

func (h *APIHandler) sponsorID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		h.writeError(w, http.StatusBadRequest, "Invalid id")
		return 0, false
	}
	return id, true
}

// activeSponsor is the sponsor a visitor's request is about, if its card is shown today. It writes
// the answer and returns false if not.
func (h *APIHandler) activeSponsor(w http.ResponseWriter, id int64) (models.Sponsor, bool) {
	sponsor, err := h.sponsors.GetSponsor(id)
	if err != nil {
		var configErr *dbx.ConfigError
		if errors.As(err, &configErr) && configErr.Kind == dbx.ConfigNotFound {
			h.writeError(w, http.StatusNotFound, "Not found")
			return sponsor, false
		}
		h.logger.Errorf("failed to load sponsor %d: %v", id, err)
		h.writeError(w, http.StatusInternalServerError, "Failed to load the sponsor")
		return sponsor, false
	}
	if !sponsor.ActiveOn(sponsorToday()) {
		h.writeError(w, http.StatusNotFound, "Not found")
		return sponsor, false
	}
	return sponsor, true
}

func (h *APIHandler) countSponsor(sponsor models.Sponsor, kind dbx.SponsorCount) {
	if err := h.sponsors.CountSponsor(sponsor.ID, kind, sponsorToday()); err != nil {
		h.logger.Errorf("failed to count sponsor %d: %v", sponsor.ID, err)
	}
}

// handleGetSponsors lists the cards shown today. The calendar picks one of them.
func (h *APIHandler) handleGetSponsors(w http.ResponseWriter, r *http.Request) {
	sponsors, err := h.sponsors.ListSponsors()
	if err != nil {
		h.logger.Errorf("failed to list sponsors: %v", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to fetch sponsors")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	h.writeJSON(w, http.StatusOK, mapper.SponsorResponses(sponsors, sponsorToday()))
}

// handleSponsorImage serves a sponsor's picture. The address carries the time of the last change,
// so a picture can be cached for good.
func (h *APIHandler) handleSponsorImage(w http.ResponseWriter, r *http.Request) {
	id, ok := h.sponsorID(w, r)
	if !ok {
		return
	}
	image, kind, err := h.sponsors.SponsorImage(id)
	if err != nil || len(image) == 0 {
		var configErr *dbx.ConfigError
		if err != nil && !errors.As(err, &configErr) {
			h.logger.Errorf("failed to load the picture of sponsor %d: %v", id, err)
			h.writeError(w, http.StatusInternalServerError, "Failed to load the picture")
			return
		}
		h.writeError(w, http.StatusNotFound, "Not found")
		return
	}
	w.Header().Set("Content-Type", kind)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(image)
}

// handleSponsorImpression counts that a card was seen. It records no cookie, address or visitor id.
func (h *APIHandler) handleSponsorImpression(w http.ResponseWriter, r *http.Request) {
	id, ok := h.sponsorID(w, r)
	if !ok {
		return
	}
	sponsor, ok := h.activeSponsor(w, id)
	if !ok {
		return
	}
	if h.sponsorCountAllowed(r) {
		h.countSponsor(sponsor, dbx.SponsorImpression)
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleSponsorClick counts a click and sends the visitor to the sponsor's link.
func (h *APIHandler) handleSponsorClick(w http.ResponseWriter, r *http.Request) {
	id, ok := h.sponsorID(w, r)
	if !ok {
		return
	}
	sponsor, ok := h.activeSponsor(w, id)
	if !ok {
		return
	}
	if h.sponsorCountAllowed(r) {
		h.countSponsor(sponsor, dbx.SponsorClick)
	}
	http.Redirect(w, r, sponsor.Link, http.StatusFound)
}

func (h *APIHandler) handleAdminSponsors(w http.ResponseWriter, r *http.Request) {
	if h.getAuthenticatedUser(r) == nil {
		h.writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	today := sponsorToday()
	rows, err := h.sponsors.ListSponsorRows(today)
	if err != nil {
		h.logger.Errorf("failed to list sponsors for the admin: %v", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to fetch the sponsors")
		return
	}
	h.writeJSON(w, http.StatusOK, mapper.AdminSponsorResponses(rows, today))
}

// sponsorInput reads a request. If the picture can't be read it writes a 400 and returns false.
func (h *APIHandler) sponsorInput(w http.ResponseWriter, req dto.SponsorRequest) (dbx.SponsorInput, bool) {
	in := dbx.SponsorInput{
		Name: req.Name, Text: req.Text, Link: req.Link,
		StartsOn: req.StartsOn, EndsOn: req.EndsOn,
		Enabled: req.Enabled, RemoveImage: req.RemoveImage,
	}
	if req.Image != nil && *req.Image != "" {
		image, err := base64.StdEncoding.DecodeString(*req.Image)
		if err != nil {
			h.writeError(w, http.StatusBadRequest, "The picture could not be read")
			return in, false
		}
		in.Image = image
	}
	return in, true
}

func (h *APIHandler) handleAdminSponsorCreate(w http.ResponseWriter, r *http.Request) {
	var req dto.SponsorRequest
	if !h.decodeJSON(w, r, &req) {
		return
	}
	in, ok := h.sponsorInput(w, req)
	if !ok {
		return
	}
	h.configChange(w, r, func() (dto.ConfigChangeResponse, error) {
		return dto.ConfigChangeResponse{}, h.sponsors.CreateSponsor(in)
	})
}

func (h *APIHandler) handleAdminSponsorUpdate(w http.ResponseWriter, r *http.Request) {
	id, ok := h.sponsorID(w, r)
	var req dto.SponsorRequest
	if !ok || !h.decodeJSON(w, r, &req) {
		return
	}
	in, ok := h.sponsorInput(w, req)
	if !ok {
		return
	}
	h.configChange(w, r, func() (dto.ConfigChangeResponse, error) {
		return dto.ConfigChangeResponse{}, h.sponsors.UpdateSponsor(id, in)
	})
}

func (h *APIHandler) handleAdminSponsorDelete(w http.ResponseWriter, r *http.Request) {
	id, ok := h.sponsorID(w, r)
	if !ok {
		return
	}
	h.configChange(w, r, func() (dto.ConfigChangeResponse, error) {
		return dto.ConfigChangeResponse{}, h.sponsors.DeleteSponsor(id)
	})
}
