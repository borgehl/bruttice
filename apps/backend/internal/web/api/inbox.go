package api

import (
	"net/http"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/borgehl/bruttice/apps/backend/internal/web/dto"
	"github.com/borgehl/bruttice/apps/backend/internal/web/mapper"
	"github.com/borgehl/bruttice/packages/go-pkg/models"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/emailx"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/util"
	"github.com/google/uuid"
)

// inboxStore is what the inbox needs from the database.
type inboxStore interface {
	CreateInboxMessage(message *models.InboxMessage) error
	InboxProviderIDSeen(providerID string) (bool, error)
	FindInboxMessageByMessageID(messageID string) (*models.InboxMessage, error)
	MarkInboxDelivered(sentID string, at int64) error
	GetInboxThread(threadID string) ([]models.InboxMessage, error)
	ListInboxMessages(includeArchived bool) ([]models.InboxMessage, error)
	CountUnreadInboxThreads() (int64, error)
	MarkInboxThreadRead(threadID string) error
	SetInboxThreadArchived(threadID string, archived bool) error
	DeleteInboxThread(threadID string) error
	DeleteArchivedInboxThreads() (int64, error)
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

// handleAdminInboxDelete removes a conversation for good, whatever its state.
func (h *APIHandler) handleAdminInboxDelete(w http.ResponseWriter, r *http.Request) {
	threadID, _, ok := h.adminThread(w, r)
	if !ok {
		return
	}
	if err := h.inbox.DeleteInboxThread(threadID); err != nil {
		h.logger.Errorf("failed to delete inbox thread %s: %v", threadID, err)
		h.writeError(w, http.StatusInternalServerError, "Failed to delete the conversation")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleAdminInboxDeleteArchived removes every archived conversation for good.
func (h *APIHandler) handleAdminInboxDeleteArchived(w http.ResponseWriter, r *http.Request) {
	if h.getAuthenticatedUser(r) == nil {
		h.writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	deleted, err := h.inbox.DeleteArchivedInboxThreads()
	if err != nil {
		h.logger.Errorf("failed to delete the archived inbox threads: %v", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to delete the archived conversations")
		return
	}
	h.logger.Infof("deleted %d archived inbox messages", deleted)
	h.writeJSON(w, http.StatusOK, dto.InboxDeleteArchivedResponse{Deleted: deleted})
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
	// an outreach nobody has answered yet: write to the person it was sent to
	to, inReplyTo := "", ""
	if target != nil {
		to, inReplyTo = target.FromEmail, target.MessageID
	} else {
		to = messages[0].ToEmail
	}
	if to == "" {
		h.writeError(w, http.StatusConflict, "There is nobody to reply to")
		return
	}

	subject := "Re: " + messages[0].Subject
	if strings.HasPrefix(strings.ToLower(messages[0].Subject), "re:") {
		subject = messages[0].Subject
	}
	// the email shows what is being answered, quoted under the reply; the conversation keeps only the reply
	text := body
	if target != nil {
		who := target.FromEmail
		if target.FromName != "" {
			who = target.FromName + " <" + target.FromEmail + ">"
		}
		when := time.Unix(target.CreatedAt, 0).In(util.TZLocation).Format("Mon, 2 Jan 2006 at 3:04 PM MST")
		text = emailx.QuoteReply(body, "On "+when+", "+who+" wrote", target.Body)
	}
	sentID, err := h.email.Send(h.email.ReplyMessage(to, user.Name, subject, text, threadID, inReplyTo))
	if err != nil {
		h.writeError(w, http.StatusBadGateway, "The reply could not be sent. Try again later")
		return
	}

	reply := &models.InboxMessage{
		ThreadID:  threadID,
		Source:    models.InboxSourceAdminReply,
		FromName:  user.DisplayName(),
		FromEmail: h.email.Contact(),
		ToEmail:   to,
		Subject:   subject,
		Body:      body,
		InReplyTo: inReplyTo,
		SentID:    optionalID(sentID),
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

const maxComposeSubject = 200

// handleAdminInboxCompose emails someone who hasn't written to the site, and starts a conversation
// with them. Their answer comes back to the conversation's reply address, like any reply.
func (h *APIHandler) handleAdminInboxCompose(w http.ResponseWriter, r *http.Request) {
	user := h.getAuthenticatedUser(r)
	if user == nil {
		h.writeError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	var req dto.InboxComposeRequest
	if !h.decodeJSON(w, r, &req) {
		return
	}
	to := strings.TrimSpace(req.To)
	subject := strings.TrimSpace(req.Subject)
	body := strings.TrimSpace(req.Body)

	// one plain address: no display name, no list, nothing that could add recipients or headers
	parsed, err := mail.ParseAddress(to)
	switch {
	case err != nil || parsed.Address != to || strings.ContainsAny(to, "\r\n"):
		h.writeError(w, http.StatusBadRequest, "Enter one email address")
		return
	case strings.EqualFold(to, h.email.Contact()):
		h.writeError(w, http.StatusBadRequest, "That is the site's own address")
		return
	case subject == "" || strings.ContainsAny(subject, "\r\n"):
		h.writeError(w, http.StatusBadRequest, "Write a subject on one line")
		return
	case utf8.RuneCountInString(subject) > maxComposeSubject:
		h.writeError(w, http.StatusBadRequest, "The subject is too long")
		return
	case body == "":
		h.writeError(w, http.StatusBadRequest, "Write a message first")
		return
	case utf8.RuneCountInString(body) > maxReply:
		h.writeError(w, http.StatusBadRequest, "The message is too long")
		return
	}
	if !h.email.Enabled() {
		h.writeError(w, http.StatusServiceUnavailable, "Email is not set up, the message can't be sent")
		return
	}

	threadID := uuid.NewString()
	sentID, err := h.email.Send(h.email.ReplyMessage(to, user.Name, subject, body, threadID, ""))
	if err != nil {
		h.writeError(w, http.StatusBadGateway, "The message could not be sent. Try again later")
		return
	}
	message := &models.InboxMessage{
		ThreadID:  threadID,
		Source:    models.InboxSourceAdminReply,
		FromName:  user.DisplayName(),
		FromEmail: h.email.Contact(),
		ToEmail:   to,
		Subject:   subject,
		Body:      body,
		SentID:    optionalID(sentID),
	}
	if err := h.inbox.CreateInboxMessage(message); err != nil {
		// it was sent, so say so: a second try would send it twice
		h.logger.Errorf("outreach %s was sent but not saved: %v", threadID, err)
		h.writeError(w, http.StatusInternalServerError, "The message was sent but could not be saved")
		return
	}
	h.writeJSON(w, http.StatusCreated, dto.InboxComposeResponse{ThreadID: threadID})
}

// optionalID is nil for an empty id, so a column that is unique or looked up never holds "".
func optionalID(id string) *string {
	if id == "" {
		return nil
	}
	return &id
}
