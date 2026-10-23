package api

import (
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/borgehl/bruttice/apps/backend/internal/web/dto"
	"github.com/borgehl/bruttice/apps/backend/internal/web/mapper"
	"github.com/borgehl/bruttice/packages/go-pkg/models"
	"github.com/google/uuid"
)

// inboxStore is what the inbox needs from the database.
type inboxStore interface {
	CreateInboxMessage(message *models.InboxMessage) error
	InboxProviderIDSeen(providerID string) (bool, error)
	FindInboxMessageByMessageID(messageID string) (*models.InboxMessage, error)
	GetInboxThread(threadID string) ([]models.InboxMessage, error)
	ListInboxMessages(includeArchived bool) ([]models.InboxMessage, error)
	CountUnreadInboxThreads() (int64, error)
	MarkInboxThreadRead(threadID string) error
	SetInboxThreadArchived(threadID string, archived bool) error
}

const maxReply = 10000

func (h *APIHandler) handleAdminInbox(w http.ResponseWriter, r *http.Request) {
	if h.getAuthenticatedUser(r) == nil {
		h.writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	messages, err := h.inbox.ListInboxMessages(r.URL.Query().Get("archived") == "true")
	if err != nil {
		h.logger.Errorf("failed to list inbox messages: %v", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to fetch messages")
		return
	}
	h.writeJSON(w, http.StatusOK, mapper.InboxThreadResponses(messages))
}

// adminThread is the conversation named in the path, for a signed in admin. If it can't be
// returned it has written why.
func (h *APIHandler) adminThread(w http.ResponseWriter, r *http.Request) (string, []models.InboxMessage, bool) {
	if h.getAuthenticatedUser(r) == nil {
		h.writeError(w, http.StatusUnauthorized, "Authentication required")
		return "", nil, false
	}
	threadID := r.PathValue("threadId")
	if _, err := uuid.Parse(threadID); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid conversation")
		return "", nil, false
	}
	messages, err := h.inbox.GetInboxThread(threadID)
	if err != nil {
		h.logger.Errorf("failed to get inbox thread %s: %v", threadID, err)
		h.writeError(w, http.StatusInternalServerError, "Failed to fetch messages")
		return "", nil, false
	}
	if len(messages) == 0 {
		h.writeError(w, http.StatusNotFound, "Conversation not found")
		return "", nil, false
	}
	return threadID, messages, true
}

func (h *APIHandler) handleAdminInboxRead(w http.ResponseWriter, r *http.Request) {
	threadID, _, ok := h.adminThread(w, r)
	if !ok {
		return
	}
	if err := h.inbox.MarkInboxThreadRead(threadID); err != nil {
		h.logger.Errorf("failed to mark inbox thread %s read: %v", threadID, err)
		h.writeError(w, http.StatusInternalServerError, "Failed to update the conversation")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *APIHandler) handleAdminInboxArchive(archived bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		threadID, _, ok := h.adminThread(w, r)
		if !ok {
			return
		}
		if err := h.inbox.SetInboxThreadArchived(threadID, archived); err != nil {
			h.logger.Errorf("failed to set inbox thread %s archived=%v: %v", threadID, archived, err)
			h.writeError(w, http.StatusInternalServerError, "Failed to update the conversation")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// handleAdminInboxReply emails an answer to the person a conversation is with, and keeps it in
// the conversation. Their reply to it comes back to the conversation's reply address.
func (h *APIHandler) handleAdminInboxReply(w http.ResponseWriter, r *http.Request) {
	threadID, messages, ok := h.adminThread(w, r)
	if !ok {
		return
	}
	user := h.getAuthenticatedUser(r)

	var req dto.InboxReplyRequest
	if !h.decodeJSON(w, r, &req) {
		return
	}
	body := strings.TrimSpace(req.Body)
	switch {
	case body == "":
		h.writeError(w, http.StatusBadRequest, "Write a reply first")
		return
	case utf8.RuneCountInString(body) > maxReply:
		h.writeError(w, http.StatusBadRequest, "The reply is too long")
		return
	}
	if !h.email.Enabled() {
		h.writeError(w, http.StatusServiceUnavailable, "Email is not set up, the reply can't be sent")
		return
	}

	// answer the latest message from the other side: that is who it is with, and its
	// Message-ID is what a mail client threads the answer under
	var target *models.InboxMessage
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Source != models.InboxSourceAdminReply {
			target = &messages[i]
			break
		}
	}
	if target == nil || target.FromEmail == "" {
		h.writeError(w, http.StatusConflict, "There is nobody to reply to")
		return
	}

	subject := "Re: " + messages[0].Subject
	if strings.HasPrefix(strings.ToLower(messages[0].Subject), "re:") {
		subject = messages[0].Subject
	}
	if _, err := h.email.Send(h.email.ReplyMessage(target.FromEmail, subject, body, threadID, target.MessageID)); err != nil {
		h.writeError(w, http.StatusBadGateway, "The reply could not be sent. Try again later")
		return
	}

	reply := &models.InboxMessage{
		ThreadID:  threadID,
		Source:    models.InboxSourceAdminReply,
		FromName:  user.DisplayName(),
		FromEmail: h.email.Contact(),
		Subject:   subject,
		Body:      body,
		InReplyTo: target.MessageID,
	}
	if err := h.inbox.CreateInboxMessage(reply); err != nil {
		// it was sent, so say so: a second try would send it twice
		h.logger.Errorf("reply to %s was sent but not saved: %v", threadID, err)
		h.writeError(w, http.StatusInternalServerError, "The reply was sent but could not be saved")
		return
	}
	// answering means the conversation has been read
	if err := h.inbox.MarkInboxThreadRead(threadID); err != nil {
		h.logger.Errorf("failed to mark inbox thread %s read: %v", threadID, err)
	}
	w.WriteHeader(http.StatusNoContent)
}
