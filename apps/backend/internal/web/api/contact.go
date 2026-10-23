package api

import (
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/borgehl/bruttice/apps/backend/internal/web/dto"
	"github.com/borgehl/bruttice/packages/go-pkg/models"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/util"
	"github.com/google/uuid"
)

const (
	maxContactName    = 100
	maxContactEmail   = 254
	minContactMessage = 10
	maxContactMessage = 5000
)

// handleContact saves a visitor's message in the admin inbox and tells the root admin about it.
// The site's address is never shown. Answers are written in the inbox and emailed from there.
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
	text := strings.TrimSpace(req.Message)
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
	case utf8.RuneCountInString(text) < minContactMessage:
		h.writeError(w, http.StatusBadRequest, "Message is too short")
		return
	case utf8.RuneCountInString(text) > maxContactMessage:
		h.writeError(w, http.StatusBadRequest, "Message is too long")
		return
	}

	if !h.contactLimit.Allow(util.GetClientIPAddr(r), time.Now()) {
		h.writeError(w, http.StatusTooManyRequests, "Too many messages. Try again later")
		return
	}

	// Keep it first: whatever happens to the notice, the message must not be lost.
	saved := &models.InboxMessage{
		ThreadID:  uuid.NewString(),
		Source:    models.InboxSourceContactForm,
		FromName:  name,
		FromEmail: email,
		Subject:   "Contact form message",
		Body:      text,
	}
	if err := h.inbox.CreateInboxMessage(saved); err != nil {
		h.logger.Errorf("failed to save contact message: %v", err)
		h.writeError(w, http.StatusInternalServerError, "Your message could not be sent. Try again later")
		return
	}
	go func() { _, _ = h.email.Send(h.email.InboxNoticeMessage(name, email, saved.Subject, saved.Body)) }()
	h.writeJSON(w, http.StatusOK, dto.AddEventResponse{Success: true, Message: contactSent})
}

const contactSent = "Thanks, your message was sent."
