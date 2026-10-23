package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/borgehl/bruttice/packages/go-pkg/models"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/emailx"
	"github.com/google/uuid"
)

const (
	maxWebhookBody = 1 << 20
	maxInboundBody = 50000
)

// resendEvent is what Resend posts for every event of its webhook. For email.received the data
// has who the email is from and for, but not its body or its headers.
type resendEvent struct {
	Type string `json:"type"`
	Data struct {
		EmailID   string   `json:"email_id"`
		From      string   `json:"from"`
		To        []string `json:"to"`
		Cc        []string `json:"cc"`
		Subject   string   `json:"subject"`
		MessageID string   `json:"message_id"`
	} `json:"data"`
}

// handleResendWebhook receives mail sent to the site's domain. Anyone can post to it, so nothing
// is believed without Resend's signature, and with no signing secret the endpoint doesn't exist.
func (h *APIHandler) handleResendWebhook(w http.ResponseWriter, r *http.Request) {
	if !h.email.InboundEnabled() {
		h.writeError(w, http.StatusNotFound, "Not found")
		return
	}

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxWebhookBody))
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if err := h.email.VerifyWebhook(r.Header.Get("svix-id"), r.Header.Get("svix-timestamp"), r.Header.Get("svix-signature"), body, time.Now()); err != nil {
		h.logger.Warnf("resend webhook refused: %v", err)
		h.writeError(w, http.StatusBadRequest, "Invalid signature")
		return
	}

	var event resendEvent
	if err := json.Unmarshal(body, &event); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	// delivery events for sent mail come here too, they aren't needed
	if event.Type != "email.received" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// An error makes Resend deliver the event again later, which receiveEmail allows.
	if err := h.receiveEmail(event); err != nil {
		h.logger.Errorf("failed to save received email %s: %v", event.Data.EmailID, err)
		h.writeError(w, http.StatusInternalServerError, "Failed to save the email")
		return
	}
	w.WriteHeader(http.StatusOK)
}

// receiveEmail saves an email sent to the site as a message in the inbox. It is safe to call
// again for the same email.
func (h *APIHandler) receiveEmail(event resendEvent) error {
	d := event.Data
	if d.EmailID == "" {
		return nil
	}

	// Resend sends every email to the domain, only the site's own addresses are for the inbox.
	ours := false
	for _, address := range append(append([]string{}, d.To...), d.Cc...) {
		ours = ours || h.email.ForUs(address)
	}
	if !ours {
		h.logger.Debugf("received email %s is not for the site, ignored", d.EmailID)
		return nil
	}

	fromName, fromEmail := emailx.Sender(d.From)
	// the site's own mail, such as a notice, must not come back as a message
	if strings.EqualFold(fromEmail, h.email.Contact()) {
		return nil
	}

	seen, err := h.inbox.InboxProviderIDSeen(d.EmailID)
	if err != nil || seen {
		return err
	}

	received, err := h.email.FetchReceived(d.EmailID)
	if err != nil {
		return err
	}

	threadID, existing := h.threadFor(received, d.To, d.Cc)
	body := received.PlainText()
	if utf8.RuneCountInString(body) > maxInboundBody {
		body = string([]rune(body)[:maxInboundBody])
	}
	subject := received.Subject
	if subject == "" {
		subject = d.Subject
	}
	providerID := d.EmailID
	message := &models.InboxMessage{
		ThreadID:   threadID,
		Source:     models.InboxSourceEmail,
		FromName:   fromName,
		FromEmail:  fromEmail,
		Subject:    subject,
		Body:       body,
		MessageID:  firstNonEmpty(received.MessageID, d.MessageID),
		InReplyTo:  received.Header("In-Reply-To"),
		ProviderID: &providerID,
	}
	if err := h.inbox.CreateInboxMessage(message); err != nil {
		return err
	}

	// they wrote back, so a conversation that was put away is open again
	if existing {
		if err := h.inbox.SetInboxThreadArchived(threadID, false); err != nil {
			h.logger.Errorf("failed to reopen inbox thread %s: %v", threadID, err)
		}
	}
	go func() { _, _ = h.email.Send(h.email.InboxNoticeMessage(fromName, fromEmail, subject, body)) }()
	return nil
}

// threadFor finds the conversation a received email belongs to, and says if it already existed.
// First the reply address it was sent to, which the site put in the Reply-To of its message.
// Then the Message-ID it answers. Otherwise it starts a new conversation.
func (h *APIHandler) threadFor(received emailx.Received, to, cc []string) (string, bool) {
	for _, address := range append(append([]string{}, to...), cc...) {
		id, ok := h.email.ThreadFromAddress(address)
		if !ok {
			continue
		}
		// the id comes from outside, only believe it if the conversation is real
		if messages, err := h.inbox.GetInboxThread(id); err == nil && len(messages) > 0 {
			return id, true
		}
	}

	if inReplyTo := strings.TrimSpace(received.Header("In-Reply-To")); inReplyTo != "" {
		if message, err := h.inbox.FindInboxMessageByMessageID(inReplyTo); err == nil && message != nil {
			return message.ThreadID, true
		}
	}
	return uuid.NewString(), false
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
