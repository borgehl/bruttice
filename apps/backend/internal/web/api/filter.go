package api

import (
	"github.com/borgehl/bruttice/apps/backend/internal/web/dto"
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
	tags := make([]dto.TagResponse, 0, len(tagNames))
	for i, name := range tagNames {
		tags = append(tags, dto.TagResponse{ID: int64(i + 1), Name: name})
	}

	// Get venues
	venueModels, err := h.database.GetVenues(tx)
	if err != nil {
		h.logger.Errorf("failed to get venues: %v", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to get venues")
		return
	}
	venues := make([]dto.VenueResponse, 0, len(venueModels))
	for _, v := range venueModels {
		venues = append(venues, dto.VenueResponse{Name: v.Name, Slug: v.Slug, Link: v.Link})
	}

	// Get ages
	ageModels, err := h.database.GetAges(tx)
	if err != nil {
		h.logger.Errorf("failed to get ages: %v", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to get ages")
		return
	}
	ages := make([]dto.AgeResponse, 0, len(ageModels))
	for _, age := range ageModels {
		if age.Value > 100 {
			// skip unknown ages for filters/event creation
			continue
		}
		ages = append(ages, dto.AgeResponse{ID: int64(age.ID), Name: age.Name, Value: age.Value})
	}

	// Get regions
	regionNames, err := h.database.GetRegionNames(tx)
	if err != nil {
		h.logger.Errorf("failed to get regions: %v", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to get regions")
		return
	}
	regions := make([]dto.RegionResponse, 0, len(regionNames))
	for i, name := range regionNames {
		regions = append(regions, dto.RegionResponse{ID: int64(i + 1), Name: name})
	}

	h.writeJSON(w, http.StatusOK, dto.FilterOptionsResponse{
		Tags: tags, Venues: venues, Ages: ages, Regions: regions,
	})
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
