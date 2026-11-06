package api

import (
	"github.com/borgehl/bruttice/apps/backend/internal/web/dto"
	"github.com/borgehl/bruttice/apps/backend/internal/web/mapper"
	"github.com/borgehl/bruttice/packages/go-pkg/models"
	"net/http"
)

func (h *APIHandler) handleGetFilterOptions(w http.ResponseWriter, r *http.Request) {
	tx := h.database.DB.Begin()
	defer tx.Commit()

	tagNames, err := h.database.GetTagNames(tx)
	if err != nil {
		h.logger.Errorf("failed to get tags: %v", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to get tags")
		return
	}

	venues, err := h.database.GetVenues(tx)
	if err != nil {
		h.logger.Errorf("failed to get venues: %v", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to get venues")
		return
	}

	ages, err := h.database.GetAges(tx)
	if err != nil {
		h.logger.Errorf("failed to get ages: %v", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to get ages")
		return
	}

	regionNames, err := h.database.GetRegionNames(tx)
	if err != nil {
		h.logger.Errorf("failed to get regions: %v", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to get regions")
		return
	}

	h.writeJSON(w, http.StatusOK, mapper.FilterOptionsResponse(tagNames, venues, ages, regionNames))
}

func (h *APIHandler) handleGetFilter(w http.ResponseWriter, r *http.Request) {
	filterKey := r.URL.Query().Get("filterKey")
	if filterKey == "" {
		h.writeJSON(w, http.StatusOK, dto.GetFilterResponse{})
		return
	}

	tx := h.database.DB.Begin()
	defer tx.Commit()

	filterModel, err := h.database.GetFilterByKey(filterKey, tx)
	if err != nil {
		h.logger.Errorf("failed to get filter by key: %v", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to get filter")
		return
	}
	if filterModel.ID == 0 {
		h.writeJSON(w, http.StatusOK, dto.GetFilterResponse{})
		return
	}

	filter := filterModel.GetBackendFilter()
	h.writeJSON(w, http.StatusOK, dto.GetFilterResponse{
		Tags:          filter.Tags,
		TagsInclude:   filter.TagsInclude,
		Age:           filter.Age,
		AgesLessThan:  filter.AgesLessThan,
		Venues:        filter.Venues,
		VenuesInclude: filter.VenuesInclude,
		Regions:       filter.Regions,
	})
}

func (h *APIHandler) handleCreateFilter(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateFilterRequest
	if !h.decodeJSON(w, r, &req) {
		return
	}

	// Find the minimum age value from the selected ages
	var ageValue uint8
	if len(req.Ages) > 0 {
		ageValue = req.Ages[0]
		for _, age := range req.Ages {
			if age < ageValue {
				ageValue = age
			}
		}
	}

	filter := models.CalendarFilterBackend{
		Tags:          req.Tags,
		TagsInclude:   req.TagsInclude,
		Age:           ageValue,
		AgesLessThan:  req.AgesLessThan,
		Venues:        req.Venues,
		VenuesInclude: req.VenuesInclude,
		Regions:       regionsFromQuery(req.Regions),
	}

	dbFilter := filter.GetDBFilter()

	tx := h.database.DB.Begin()
	defer tx.Commit()

	filterRecord, err := h.database.SaveFilter(dbFilter, tx)
	if err != nil {
		h.logger.Errorf("failed to save filter: %v", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to create filter")
		return
	}

	h.writeJSON(w, http.StatusOK, dto.CreateFilterResponse{FilterKey: filterRecord.HashKey})
}
