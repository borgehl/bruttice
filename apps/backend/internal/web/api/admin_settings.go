package api

import (
	"net/http"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/borgehl/bruttice/apps/backend/internal/web/dto"
	"github.com/borgehl/bruttice/apps/backend/internal/web/mapper"
	"github.com/borgehl/bruttice/packages/go-pkg/models"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/emailx"
)

// userStore is what the settings and the notifications need from the database.
type userStore interface {
	SetUserName(userID int64, name string) error
	// MergeUserSettings merges the patch into the user's settings and answers them as they are then
	MergeUserSettings(userID int64, patch map[string]any) (models.UserSettings, error)
	// NotificationRecipients are the emails of the admins who chose a notification (a
	// models.SettingNotify key), none if nobody did
	NotificationRecipients(kind string) ([]string, error)
}

const maxAdminName = 100

// dropControl is for strings.Map: it removes control characters, which have no place in a name.
func dropControl(r rune) rune {
	if unicode.IsControl(r) && !unicode.IsSpace(r) {
		return -1
	}
	return r
}

// handleAdminSettings changes the signed in admin's own settings: their name and which emails they
// get. A setting the request does not have is left as it is.
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
	name := user.Name
	if req.Name.Present {
		// one line: the name goes into messages and onto the screen of the other admins
		name = strings.Join(strings.Fields(strings.Map(dropControl, req.Name.Value)), " ")
		if utf8.RuneCountInString(name) > maxAdminName {
			h.writeError(w, http.StatusBadRequest, "The name is too long")
			return
		}
	}

	if req.Name.Present {
		if err := h.users.SetUserName(user.ID, name); err != nil {
			h.logger.Errorf("failed to set the name of user %d: %v", user.ID, err)
			h.writeError(w, http.StatusInternalServerError, "Failed to save the settings")
			return
		}
	}
	// merged in the database, so the other settings are neither read nor rewritten
	settings, err := h.users.MergeUserSettings(user.ID, mapper.NotificationsPatch(req.Notifications))
	if err != nil {
		h.logger.Errorf("failed to save the settings of user %d: %v", user.ID, err)
		h.writeError(w, http.StatusInternalServerError, "Failed to save the settings")
		return
	}
	h.writeJSON(w, http.StatusOK, dto.LoginResponse{
		Success: true, Email: user.Email, Name: name, Notifications: mapper.Notifications(settings),
	})
}

// notify sends a message to the admins who chose the notification (a models.SettingNotify key).
// If nobody did, or the lookup fails, it is not sent and a line says why.
func (h *APIHandler) notify(kind string, msg emailx.Message) {
	to, err := h.users.NotificationRecipients(kind)
	if err != nil {
		h.logger.Errorf("not sending %q: failed to look up who gets %s: %v", msg.Subject, kind, err)
		return
	}
	if len(to) == 0 {
		h.logger.Infof("not sending %q: no admin chose %s", msg.Subject, kind)
		return
	}
	msg.To = to
	_, _ = h.email.Send(msg)
}
