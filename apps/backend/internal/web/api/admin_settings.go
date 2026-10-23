package api

import (
	"net/http"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/borgehl/bruttice/apps/backend/internal/web/dto"
)

// userStore is what the settings need from the database.
type userStore interface {
	SetUserName(userID int64, name string) error
}

const maxAdminName = 100

// dropControl is for strings.Map: it removes control characters, which have no place in a name.
func dropControl(r rune) rune {
	if unicode.IsControl(r) && !unicode.IsSpace(r) {
		return -1
	}
	return r
}

// handleAdminSettings changes the signed in admin's own settings, which today is their name.
func (h *APIHandler) handleAdminSettings(w http.ResponseWriter, r *http.Request) {
	user := h.getAuthenticatedUser(r)
	if user == nil {
		h.writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req dto.AdminSettingsRequest
	if !h.decodeJSON(w, r, &req) {
		return
	}
	// one line: the name goes into messages and onto the screen of the other admins
	name := strings.Join(strings.Fields(strings.Map(dropControl, req.Name)), " ")
	if utf8.RuneCountInString(name) > maxAdminName {
		h.writeError(w, http.StatusBadRequest, "The name is too long")
		return
	}

	if err := h.users.SetUserName(user.ID, name); err != nil {
		h.logger.Errorf("failed to set the name of user %d: %v", user.ID, err)
		h.writeError(w, http.StatusInternalServerError, "Failed to save the settings")
		return
	}
	h.writeJSON(w, http.StatusOK, dto.LoginResponse{Success: true, Email: user.Email, Name: name})
}
