package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/borgehl/bruttice/apps/backend/internal/web/dto"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/dbx"
)

// configStore is what the admin settings for the site's lists need from the database. Its
// errors are *dbx.ConfigError when the admin asked for something that can't be done.
type configStore interface {
	ListRegionRows() ([]dbx.RegionRow, error)
	ListTagRows() ([]dbx.TagRow, error)
	ListAgeRows() ([]dbx.AgeRow, error)
	ListVenueRows() ([]dbx.VenueRow, error)
	CreateRegion(name string) error
	RenameRegion(id int64, name string) error
	RetireRegion(id int64, moveToID *int64, deleteEvents bool) (dbx.RetireResult, error)
	CreateTag(name string) error
	RenameTag(id int64, name string) error
	DeleteTag(id int64) error
	CreateAge(name string, value int) error
	UpdateAge(id int64, name string, value int) error
	DeleteAge(id int64, replacementID *int64) (int64, error)
	CreateVenue(name, link, region string) error
	UpdateVenue(id int64, name, link, region string) error
	DeleteVenue(id int64) error
}

// configChange runs one change of the lists for a signed in admin. fn returns what the change did
// to events, if it did anything, and its error is shown to the admin if it is a ConfigError.
func (h *APIHandler) configChange(w http.ResponseWriter, r *http.Request, fn func() (dto.ConfigChangeResponse, error)) {
	if h.getAuthenticatedUser(r) == nil {
		h.writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	result, err := fn()
	if err != nil {
		var configErr *dbx.ConfigError
		if !errors.As(err, &configErr) {
			h.logger.Errorf("failed to change the site's lists: %v", err)
			h.writeError(w, http.StatusInternalServerError, "The change could not be saved")
			return
		}
		status := map[dbx.ConfigErrorKind]int{
			dbx.ConfigInvalid:  http.StatusBadRequest,
			dbx.ConfigNotFound: http.StatusNotFound,
			dbx.ConfigConflict: http.StatusConflict,
		}[configErr.Kind]
		h.writeError(w, status, configErr.Message)
		return
	}
	h.writeJSON(w, http.StatusOK, result)
}

func (h *APIHandler) configID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		h.writeError(w, http.StatusBadRequest, "Invalid id")
		return 0, false
	}
	return id, true
}

func (h *APIHandler) handleAdminConfig(w http.ResponseWriter, r *http.Request) {
	if h.getAuthenticatedUser(r) == nil {
		h.writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	regions, err := h.config.ListRegionRows()
	var tags []dbx.TagRow
	var ages []dbx.AgeRow
	var venues []dbx.VenueRow
	if err == nil {
		tags, err = h.config.ListTagRows()
	}
	if err == nil {
		ages, err = h.config.ListAgeRows()
	}
	if err == nil {
		venues, err = h.config.ListVenueRows()
	}
	if err != nil {
		h.logger.Errorf("failed to list the site's lists: %v", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to fetch the settings")
		return
	}

	resp := dto.ConfigResponse{
		Regions: make([]dto.ConfigRegion, 0, len(regions)),
		Tags:    make([]dto.ConfigTag, 0, len(tags)),
		Ages:    make([]dto.ConfigAge, 0, len(ages)),
		Venues:  make([]dto.ConfigVenue, 0, len(venues)),
	}
	for _, v := range regions {
		resp.Regions = append(resp.Regions, dto.ConfigRegion{ID: v.ID, Name: v.Name, Events: v.Events, Venues: v.Venues, Retired: v.RetiredAt > 0, MergedInto: v.MergedInto})
	}
	for _, v := range tags {
		resp.Tags = append(resp.Tags, dto.ConfigTag{ID: v.ID, Name: v.Name, Events: v.Events})
	}
	for _, v := range ages {
		resp.Ages = append(resp.Ages, dto.ConfigAge{ID: v.ID, Name: v.Name, Value: v.Value, Events: v.Events})
	}
	for _, v := range venues {
		resp.Venues = append(resp.Venues, dto.ConfigVenue{ID: v.ID, Name: v.Name, Link: v.Link, Region: v.DefaultRegion, Scraped: v.IsScraped, Events: v.Events})
	}
	h.writeJSON(w, http.StatusOK, resp)
}

func (h *APIHandler) handleAdminConfigRegionCreate(w http.ResponseWriter, r *http.Request) {
	var req dto.ConfigNameRequest
	if !h.decodeJSON(w, r, &req) {
		return
	}
	h.configChange(w, r, func() (dto.ConfigChangeResponse, error) {
		return dto.ConfigChangeResponse{}, h.config.CreateRegion(req.Name)
	})
}

func (h *APIHandler) handleAdminConfigRegionRename(w http.ResponseWriter, r *http.Request) {
	id, ok := h.configID(w, r)
	var req dto.ConfigNameRequest
	if !ok || !h.decodeJSON(w, r, &req) {
		return
	}
	h.configChange(w, r, func() (dto.ConfigChangeResponse, error) {
		return dto.ConfigChangeResponse{}, h.config.RenameRegion(id, req.Name)
	})
}

func (h *APIHandler) handleAdminConfigRegionRetire(w http.ResponseWriter, r *http.Request) {
	id, ok := h.configID(w, r)
	var req dto.ConfigRetireRegionRequest
	if !ok || !h.decodeJSON(w, r, &req) {
		return
	}
	h.configChange(w, r, func() (dto.ConfigChangeResponse, error) {
		result, err := h.config.RetireRegion(id, req.MoveToID, req.DeleteEvents)
		return dto.ConfigChangeResponse{EventsMoved: result.EventsMoved, EventsDeleted: result.EventsDeleted}, err
	})
}

func (h *APIHandler) handleAdminConfigTagCreate(w http.ResponseWriter, r *http.Request) {
	var req dto.ConfigNameRequest
	if !h.decodeJSON(w, r, &req) {
		return
	}
	h.configChange(w, r, func() (dto.ConfigChangeResponse, error) {
		return dto.ConfigChangeResponse{}, h.config.CreateTag(req.Name)
	})
}

func (h *APIHandler) handleAdminConfigTagRename(w http.ResponseWriter, r *http.Request) {
	id, ok := h.configID(w, r)
	var req dto.ConfigNameRequest
	if !ok || !h.decodeJSON(w, r, &req) {
		return
	}
	h.configChange(w, r, func() (dto.ConfigChangeResponse, error) {
		return dto.ConfigChangeResponse{}, h.config.RenameTag(id, req.Name)
	})
}

func (h *APIHandler) handleAdminConfigTagDelete(w http.ResponseWriter, r *http.Request) {
	id, ok := h.configID(w, r)
	if !ok {
		return
	}
	h.configChange(w, r, func() (dto.ConfigChangeResponse, error) { return dto.ConfigChangeResponse{}, h.config.DeleteTag(id) })
}

func (h *APIHandler) handleAdminConfigAgeCreate(w http.ResponseWriter, r *http.Request) {
	var req dto.ConfigAgeRequest
	if !h.decodeJSON(w, r, &req) {
		return
	}
	h.configChange(w, r, func() (dto.ConfigChangeResponse, error) {
		return dto.ConfigChangeResponse{}, h.config.CreateAge(req.Name, req.Value)
	})
}

func (h *APIHandler) handleAdminConfigAgeUpdate(w http.ResponseWriter, r *http.Request) {
	id, ok := h.configID(w, r)
	var req dto.ConfigAgeRequest
	if !ok || !h.decodeJSON(w, r, &req) {
		return
	}
	h.configChange(w, r, func() (dto.ConfigChangeResponse, error) {
		return dto.ConfigChangeResponse{}, h.config.UpdateAge(id, req.Name, req.Value)
	})
}

// handleAdminConfigAgeDelete deletes an age group. replacementId says which age group takes its events.
func (h *APIHandler) handleAdminConfigAgeDelete(w http.ResponseWriter, r *http.Request) {
	id, ok := h.configID(w, r)
	if !ok {
		return
	}
	var replacement *int64
	if raw := r.URL.Query().Get("replacementId"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || value <= 0 {
			h.writeError(w, http.StatusBadRequest, "Invalid id")
			return
		}
		replacement = &value
	}
	h.configChange(w, r, func() (dto.ConfigChangeResponse, error) {
		moved, err := h.config.DeleteAge(id, replacement)
		return dto.ConfigChangeResponse{EventsMoved: moved}, err
	})
}

func (h *APIHandler) handleAdminConfigVenueCreate(w http.ResponseWriter, r *http.Request) {
	var req dto.ConfigVenueRequest
	if !h.decodeJSON(w, r, &req) {
		return
	}
	h.configChange(w, r, func() (dto.ConfigChangeResponse, error) {
		return dto.ConfigChangeResponse{}, h.config.CreateVenue(req.Name, req.Link, req.Region)
	})
}

func (h *APIHandler) handleAdminConfigVenueUpdate(w http.ResponseWriter, r *http.Request) {
	id, ok := h.configID(w, r)
	var req dto.ConfigVenueRequest
	if !ok || !h.decodeJSON(w, r, &req) {
		return
	}
	h.configChange(w, r, func() (dto.ConfigChangeResponse, error) {
		return dto.ConfigChangeResponse{}, h.config.UpdateVenue(id, req.Name, req.Link, req.Region)
	})
}

func (h *APIHandler) handleAdminConfigVenueDelete(w http.ResponseWriter, r *http.Request) {
	id, ok := h.configID(w, r)
	if !ok {
		return
	}
	h.configChange(w, r, func() (dto.ConfigChangeResponse, error) { return dto.ConfigChangeResponse{}, h.config.DeleteVenue(id) })
}
