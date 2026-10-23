package dbx

import (
	"os"
	"testing"

	"github.com/borgehl/bruttice/packages/go-pkg/models"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/logx"
	"github.com/google/uuid"
)

// The repo has no database in its tests, so this one only runs when POSTGRES_* point at one that
// can be written to, such as a throwaway container:
//
//	POSTGRES_HOST=localhost POSTGRES_PORT=5433 POSTGRES_DB=postgres POSTGRES_USER=postgres \
//	  POSTGRES_PASSWORD=test go test ./platform/dbx
func testInboxDB(t *testing.T) *DatabaseClient {
	t.Helper()
	if os.Getenv("POSTGRES_HOST") == "" {
		t.Skip("POSTGRES_HOST is not set")
	}
	d := NewDatabaseClient(logx.NewLogger(nil, nil))
	if err := d.DB.AutoMigrate(&models.InboxMessage{}); err != nil {
		t.Fatal(err)
	}
	return d
}

func addInbox(t *testing.T, d *DatabaseClient, thread, source string, provider *string) models.InboxMessage {
	t.Helper()
	m := models.InboxMessage{ThreadID: thread, Source: source, FromEmail: "sam@example.com", Subject: "Hi", Body: "text", ProviderID: provider}
	if err := d.CreateInboxMessage(&m); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.DB.Unscoped().Where("thread_id = ?", thread).Delete(&models.InboxMessage{}) })
	return m
}

func threadIDs(messages []models.InboxMessage) map[string]int {
	ret := map[string]int{}
	for _, m := range messages {
		ret[m.ThreadID]++
	}
	return ret
}

func TestInboxThreadsListAndArchive(t *testing.T) {
	d := testInboxDB(t)
	a, b := uuid.NewString(), uuid.NewString()
	addInbox(t, d, a, models.InboxSourceContactForm, nil)
	addInbox(t, d, a, models.InboxSourceAdminReply, nil)
	addInbox(t, d, b, models.InboxSourceContactForm, nil)

	listed, err := d.ListInboxMessages(false)
	if err != nil {
		t.Fatal(err)
	}
	if got := threadIDs(listed); got[a] != 2 || got[b] != 1 {
		t.Errorf("both threads with all their messages must be listed: %v", got)
	}

	if err := d.SetInboxThreadArchived(a, true); err != nil {
		t.Fatal(err)
	}
	listed, _ = d.ListInboxMessages(false)
	if got := threadIDs(listed); got[a] != 0 || got[b] != 1 {
		t.Errorf("an archived thread must be left out: %v", got)
	}
	listed, _ = d.ListInboxMessages(true)
	if got := threadIDs(listed); got[a] != 2 {
		t.Errorf("includeArchived must bring it back: %v", got)
	}

	thread, err := d.GetInboxThread(a)
	if err != nil || len(thread) != 2 || thread[0].ID > thread[1].ID || thread[0].ReadAt == 0 {
		t.Errorf("archiving reads the thread, and messages come oldest first: %+v, %v", thread, err)
	}

	// a new message does not unarchive on its own, the handler does that
	if err := d.SetInboxThreadArchived(a, false); err != nil {
		t.Fatal(err)
	}
	listed, _ = d.ListInboxMessages(false)
	if got := threadIDs(listed); got[a] != 2 {
		t.Errorf("unarchived: %v", got)
	}
}

func TestInboxUnreadCountAndRead(t *testing.T) {
	d := testInboxDB(t)
	before, err := d.CountUnreadInboxThreads()
	if err != nil {
		t.Fatal(err)
	}
	a := uuid.NewString()
	addInbox(t, d, a, models.InboxSourceContactForm, nil)
	addInbox(t, d, a, models.InboxSourceEmail, nil)
	addInbox(t, d, a, models.InboxSourceAdminReply, nil)

	after, _ := d.CountUnreadInboxThreads()
	if after != before+1 {
		t.Errorf("unread threads = %d, want %d: two unread messages in one thread count once", after, before+1)
	}
	if err := d.MarkInboxThreadRead(a); err != nil {
		t.Fatal(err)
	}
	if now, _ := d.CountUnreadInboxThreads(); now != before {
		t.Errorf("unread threads = %d, want %d", now, before)
	}
}

func TestInboxProviderIDIsUnique(t *testing.T) {
	d := testInboxDB(t)
	id := uuid.NewString()
	addInbox(t, d, uuid.NewString(), models.InboxSourceEmail, &id)

	if seen, err := d.InboxProviderIDSeen(id); err != nil || !seen {
		t.Errorf("seen = %v, err = %v", seen, err)
	}
	if seen, _ := d.InboxProviderIDSeen(uuid.NewString()); seen {
		t.Error("an unknown id is not seen")
	}
	dup := models.InboxMessage{ThreadID: uuid.NewString(), Source: models.InboxSourceEmail, ProviderID: &id}
	if err := d.CreateInboxMessage(&dup); err == nil {
		t.Error("the same received email must not be saved twice")
	}
	// messages that were not received have no id, any number of them
	addInbox(t, d, uuid.NewString(), models.InboxSourceContactForm, nil)
	addInbox(t, d, uuid.NewString(), models.InboxSourceContactForm, nil)
}

func TestInboxFindByMessageID(t *testing.T) {
	d := testInboxDB(t)
	thread := uuid.NewString()
	m := models.InboxMessage{ThreadID: thread, Source: models.InboxSourceEmail, MessageID: "<" + uuid.NewString() + "@x>"}
	if err := d.CreateInboxMessage(&m); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.DB.Unscoped().Where("thread_id = ?", thread).Delete(&models.InboxMessage{}) })

	if got, err := d.FindInboxMessageByMessageID(m.MessageID); err != nil || got == nil || got.ThreadID != thread {
		t.Errorf("got = %+v, err = %v", got, err)
	}
	if got, err := d.FindInboxMessageByMessageID("<nope@x>"); err != nil || got != nil {
		t.Errorf("an unknown Message-ID must give nil: %+v, %v", got, err)
	}
}
