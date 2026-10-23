package mapper

import (
	"sort"
	"strings"

	"github.com/borgehl/bruttice/apps/backend/internal/web/dto"
	"github.com/borgehl/bruttice/packages/go-pkg/models"
)

func InboxMessageResponse(message models.InboxMessage) dto.InboxMessageResponse {
	return dto.InboxMessageResponse{
		ID:        message.ID,
		Source:    message.Source,
		FromName:  message.FromName,
		FromEmail: message.FromEmail,
		Body:      message.Body,
		CreatedAt: message.CreatedAt,
		Unread:    message.NeedsReading(),
	}
}

// InboxThreadResponses groups messages, oldest first, into conversations. The conversation
// with the latest message comes first.
func InboxThreadResponses(messages []models.InboxMessage) []dto.InboxThreadResponse {
	byThread := map[string]*dto.InboxThreadResponse{}
	order := make([]string, 0)
	archived := map[string]bool{}

	for _, message := range messages {
		thread, ok := byThread[message.ThreadID]
		if !ok {
			thread = &dto.InboxThreadResponse{
				ThreadID: message.ThreadID,
				Subject:  baseSubject(message.Subject),
				Messages: make([]dto.InboxMessageResponse, 0, 1),
			}
			byThread[message.ThreadID] = thread
			order = append(order, message.ThreadID)
			archived[message.ThreadID] = true
		}
		thread.Messages = append(thread.Messages, InboxMessageResponse(message))
		thread.Unread = thread.Unread || message.NeedsReading()
		thread.LastAt = max(thread.LastAt, message.CreatedAt)
		archived[message.ThreadID] = archived[message.ThreadID] && message.ArchivedAt > 0
		// who it is with: the latest person who wrote, not the admins who answered
		if message.Source != models.InboxSourceAdminReply {
			thread.FromName, thread.FromEmail = message.FromName, message.FromEmail
		}
	}

	ret := make([]dto.InboxThreadResponse, 0, len(order))
	for _, id := range order {
		byThread[id].Archived = archived[id]
		ret = append(ret, *byThread[id])
	}
	sort.SliceStable(ret, func(i, j int) bool { return ret[i].LastAt > ret[j].LastAt })
	return ret
}

// baseSubject drops the Re: prefixes answers pile up, so a conversation has one title.
func baseSubject(subject string) string {
	subject = strings.TrimSpace(subject)
	for {
		lower := strings.ToLower(subject)
		if !strings.HasPrefix(lower, "re:") {
			return subject
		}
		subject = strings.TrimSpace(subject[3:])
	}
}
