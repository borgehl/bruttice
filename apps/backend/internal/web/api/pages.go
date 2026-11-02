package api

import (
	"net/http"

	"github.com/borgehl/bruttice/apps/backend/internal/web/mapper"
)

func (h *APIHandler) handleGetPages(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, mapper.PageSummaryResponses(h.pages))
}

func (h *APIHandler) handleGetPage(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	for _, page := range h.pages {
		if page.Slug == slug {
			h.writeJSON(w, http.StatusOK, mapper.PageResponse(page))
			return
		}
	}
	h.writeError(w, http.StatusNotFound, "Page not found")
}
