package api

import (
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/borgehl/bruttice/apps/backend/internal/web/dto"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/util"
)

const (
	maxContactName    = 100
	maxContactEmail   = 254
	minContactMessage = 10
	maxContactMessage = 5000
)

// handleContact sends a visitor's message to the site's contact address. The address itself is
// never shown, a reply goes to the visitor because their address is the message's Reply-To.
func (h *APIHandler) handleContact(w http.ResponseWriter, r *http.Request) {
	if !h.email.Enabled() {
		h.writeError(w, http.StatusServiceUnavailable, "The contact form is not available")
		return
	}

	var req dto.ContactRequest
	if !h.decodeJSON(w, r, &req) {
		return
	}

	// A bot fills in every field. Answer as if it worked, so it has nothing to learn from.
	if req.Website != "" {
		h.writeJSON(w, http.StatusOK, dto.AddEventResponse{Success: true, Message: contactSent})
		return
	}

	name := strings.TrimSpace(req.Name)
	email := strings.TrimSpace(req.Email)
	message := strings.TrimSpace(req.Message)
	switch {
	case name == "":
		h.writeError(w, http.StatusBadRequest, "Name is required")
		return
	case utf8.RuneCountInString(name) > maxContactName:
		h.writeError(w, http.StatusBadRequest, "Name is too long")
		return
	case len(email) > maxContactEmail || !emailRegex.MatchString(email):
		h.writeError(w, http.StatusBadRequest, "Invalid email address")
		return
	case utf8.RuneCountInString(message) < minContactMessage:
		h.writeError(w, http.StatusBadRequest, "Message is too short")
		return
	case utf8.RuneCountInString(message) > maxContactMessage:
		h.writeError(w, http.StatusBadRequest, "Message is too long")
		return
	}

	if !h.contactLimit.Allow(util.GetClientIPAddr(r), time.Now()) {
		h.writeError(w, http.StatusTooManyRequests, "Too many messages. Try again later")
		return
	}

	if err := h.email.Send(h.email.ContactMessage(name, email, message)); err != nil {
		h.writeError(w, http.StatusBadGateway, "Your message could not be sent. Try again later")
		return
	}
	h.writeJSON(w, http.StatusOK, dto.AddEventResponse{Success: true, Message: contactSent})
}

const contactSent = "Thanks, your message was sent."
