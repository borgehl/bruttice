package api

import "net/http"

func (h *APIHandler) handleGetSite(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, h.site)
}
