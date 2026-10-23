package mapper

import (
	"testing"

	"github.com/borgehl/bruttice/packages/go-pkg/models"
)

func message(id int64, thread, source, email, subject string, at int64, read, archived int64) models.InboxMessage {
	return models.InboxMessage{
		BaseModel: models.BaseModel{ID: id, CreatedAt: at},
		ThreadID:  thread, Source: source, FromName: email, FromEmail: email, Subject: subject,
		ReadAt: read, ArchivedAt: archived,
	}
}

func TestInboxThreadResponses(t *testing.T) {
	threads := InboxThreadResponses([]models.InboxMessage{
		message(1, "a", models.InboxSourceContactForm, "sam@example.com", "Hi", 100, 5, 0),
		message(2, "b", models.InboxSourceContactForm, "kim@example.com", "Question", 150, 0, 0),
		message(3, "a", models.InboxSourceAdminReply, "hello@example.com", "Re: Hi", 200, 0, 0),
		message(4, "a", models.InboxSourceEmail, "sam@example.com", "Re: Re: Hi", 300, 0, 0),
	})
	if len(threads) != 2 {
		t.Fatalf("threads = %d", len(threads))
	}
	a, b := threads[0], threads[1]
	if a.ThreadID != "a" || b.ThreadID != "b" {
		t.Errorf("the thread with the latest message comes first: %s, %s", a.ThreadID, b.ThreadID)
	}
	if a.Subject != "Hi" || a.FromEmail != "sam@example.com" || a.LastAt != 300 || len(a.Messages) != 3 {
		t.Errorf("thread a = %+v", a)
	}
	if a.Messages[0].ID != 1 || a.Messages[2].ID != 4 {
		t.Errorf("messages must stay oldest first: %+v", a.Messages)
	}
	if !a.Unread || !b.Unread {
		t.Error("an unread message from outside makes the thread unread")
	}
	if a.Messages[1].Unread {
		t.Error("an admin's own reply is never unread")
	}
}

func TestInboxThreadReadAndArchived(t *testing.T) {
	threads := InboxThreadResponses([]models.InboxMessage{
		message(1, "a", models.InboxSourceContactForm, "sam@example.com", "Hi", 100, 5, 7),
		message(2, "a", models.InboxSourceAdminReply, "hello@example.com", "Re: Hi", 200, 0, 7),
		message(3, "b", models.InboxSourceContactForm, "kim@example.com", "Q", 150, 5, 7),
		message(4, "b", models.InboxSourceEmail, "kim@example.com", "Re: Q", 160, 0, 0),
	})
	byID := map[string]bool{}
	for _, th := range threads {
		byID[th.ThreadID] = th.Archived
		if th.ThreadID == "a" && th.Unread {
			t.Error("a thread whose outside messages are read is read")
		}
	}
	if !byID["a"] {
		t.Error("a thread with every message archived is archived")
	}
	if byID["b"] {
		t.Error("a new message brings an archived thread back")
	}
}

func TestBaseSubject(t *testing.T) {
	for in, want := range map[string]string{"Hi": "Hi", "Re: Hi": "Hi", "RE: re:  Hi there": "Hi there", "": "", "Regarding x": "Regarding x"} {
		if got := baseSubject(in); got != want {
			t.Errorf("baseSubject(%q) = %q, want %q", in, got, want)
		}
	}
}
