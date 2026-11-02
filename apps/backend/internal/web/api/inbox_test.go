package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/borgehl/bruttice/apps/backend/internal/web/dto"
	"github.com/borgehl/bruttice/packages/go-pkg/models"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/emailx"
)

const replyAddress = "reply-" + testThread + "@mail.example.com"

func TestWebhookDoesNotExistWithoutASecret(t *testing.T) {
	hr := newHarness(t, true, false)
	if rec := hr.do(http.MethodPost, "/api/webhooks/resend", `{}`); rec.Code != http.StatusNotFound {
		t.Errorf("status = %d", rec.Code)
	}
}

func TestWebhookRefusesUnsignedAndOldRequests(t *testing.T) {
	hr := newHarness(t, true, true)
	body := `{"type":"email.received","data":{"email_id":"e1","from":"sam@example.com","to":["hello@mail.example.com"]}}`
	if rec := hr.do(http.MethodPost, "/api/webhooks/resend", body); rec.Code != http.StatusBadRequest {
		t.Errorf("unsigned: status = %d", rec.Code)
	}
	if rec := hr.signedWebhook(body, time.Now().Add(-time.Hour)); rec.Code != http.StatusBadRequest {
		t.Errorf("an hour old: status = %d", rec.Code)
	}
	if len(hr.inbox.all()) != 0 {
		t.Error("nothing may be saved from a refused request")
	}
}

func TestWebhookIgnoresOtherEvents(t *testing.T) {
	hr := newHarness(t, true, true)
	rec := hr.signedWebhook(`{"type":"email.bounced","data":{"email_id":"e1"}}`, time.Now())
	if rec.Code != http.StatusOK || len(hr.inbox.all()) != 0 {
		t.Errorf("status = %d, saved = %d", rec.Code, len(hr.inbox.all()))
	}
}

func TestInboundReplyJoinsTheThreadItWasAddressedTo(t *testing.T) {
	hr := newHarness(t, true, true)
	hr.seedThread(t)
	hr.resend.received["e1"] = emailx.Received{ID: "e1", Subject: "Re: Contact form message", MessageID: "<in-1@mail.example>", Text: "Yes, all ages!", Headers: map[string]any{"In-Reply-To": "<unknown@x>"}}

	rec := hr.webhook("e1", `Sam <sam@example.com>`, replyAddress)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
	}
	saved := hr.inbox.all()
	if len(saved) != 2 {
		t.Fatalf("saved = %d", len(saved))
	}
	got := saved[1]
	if got.ThreadID != testThread || got.Source != models.InboxSourceEmail || got.FromEmail != "sam@example.com" || got.FromName != "Sam" || got.Body != "Yes, all ages!" {
		t.Errorf("message = %+v", got)
	}
	if got.ProviderID == nil || *got.ProviderID != "e1" || got.MessageID != "<in-1@mail.example>" {
		t.Errorf("ids = %v, %q", got.ProviderID, got.MessageID)
	}
	hr.resend.waitForSent(t, 1)
}

func TestInboundReplyJoinsTheThreadByInReplyTo(t *testing.T) {
	hr := newHarness(t, true, true)
	hr.seedThread(t)
	// a mail client that ignored Reply-To still answers the Message-ID of the message
	hr.inbox.messages[0].MessageID = "<ours-1@mail.example>"
	hr.resend.received["e2"] = emailx.Received{ID: "e2", Text: "Thanks", Headers: map[string]any{"in-reply-to": "<ours-1@mail.example>"}}

	hr.webhook("e2", "sam@example.com", "hello@mail.example.com")
	saved := hr.inbox.all()
	if len(saved) != 2 || saved[1].ThreadID != testThread {
		t.Errorf("saved = %+v", saved)
	}
}

func TestInboundWithoutAThreadStartsOne(t *testing.T) {
	hr := newHarness(t, true, true)
	hr.seedThread(t)
	hr.resend.received["e3"] = emailx.Received{ID: "e3", Subject: "Booking", Text: "Can we book the hall?"}

	hr.webhook("e3", "kim@example.com", "hello@mail.example.com")
	saved := hr.inbox.all()
	if len(saved) != 2 || saved[1].ThreadID == testThread || saved[1].ThreadID == "" || saved[1].Subject != "Booking" {
		t.Errorf("saved = %+v", saved)
	}
}

func TestInboundDoesNotTrustAMadeUpThreadAddress(t *testing.T) {
	hr := newHarness(t, true, true)
	hr.seedThread(t)
	hr.resend.received["e4"] = emailx.Received{ID: "e4", Text: "hi"}

	hr.webhook("e4", "kim@example.com", "reply-11111111-2222-4333-8444-555555555555@mail.example.com")
	saved := hr.inbox.all()
	if len(saved) != 2 || saved[1].ThreadID == testThread || saved[1].ThreadID == "11111111-2222-4333-8444-555555555555" {
		t.Errorf("a thread id from outside must not create or join a thread: %+v", saved)
	}
}

func TestInboundIsOnlyForTheSitesAddresses(t *testing.T) {
	hr := newHarness(t, true, true)
	hr.resend.received["e5"] = emailx.Received{ID: "e5", Text: "hi"}
	if rec := hr.webhook("e5", "kim@example.com", "someone-else@mail.example.com"); rec.Code != http.StatusOK {
		t.Errorf("status = %d", rec.Code)
	}
	if len(hr.inbox.all()) != 0 {
		t.Error("mail for another address on the domain is not ours")
	}
}

func TestInboundIgnoresTheSitesOwnMail(t *testing.T) {
	hr := newHarness(t, true, true)
	hr.resend.received["e6"] = emailx.Received{ID: "e6", Text: "notice"}
	hr.webhook("e6", "Events <hello@mail.example.com>", "hello@mail.example.com")
	if len(hr.inbox.all()) != 0 {
		t.Error("mail from the site's own address would loop")
	}
}

func TestInboundDeliveredTwiceIsSavedOnce(t *testing.T) {
	hr := newHarness(t, true, true)
	hr.resend.received["e7"] = emailx.Received{ID: "e7", Text: "hello"}
	for i := 0; i < 2; i++ {
		if rec := hr.webhook("e7", "kim@example.com", "hello@mail.example.com"); rec.Code != http.StatusOK {
			t.Fatalf("delivery %d: status = %d", i+1, rec.Code)
		}
	}
	if len(hr.inbox.all()) != 1 {
		t.Errorf("saved = %d, want 1", len(hr.inbox.all()))
	}
}

// Resend delivers the event again on an error, so a failure to fetch the body must be one.
func TestInboundFailsSoResendTriesAgain(t *testing.T) {
	hr := newHarness(t, true, true)
	rec := hr.webhook("missing", "kim@example.com", "hello@mail.example.com")
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d", rec.Code)
	}
	if len(hr.inbox.all()) != 0 {
		t.Error("nothing is saved without the body")
	}
}

func TestInboundHTMLIsStoredAsText(t *testing.T) {
	hr := newHarness(t, true, true)
	hr.resend.received["e8"] = emailx.Received{ID: "e8", HTML: `<p>Hi <b>there</b></p><script>steal()</script><img src=x onerror=alert(1)>`}
	hr.webhook("e8", "kim@example.com", "hello@mail.example.com")
	body := hr.inbox.all()[0].Body
	if strings.ContainsAny(body, "<>") || strings.Contains(body, "steal") || !strings.Contains(body, "Hi there") {
		t.Errorf("body = %q", body)
	}
}

func TestInboundReopensAnArchivedThread(t *testing.T) {
	hr := newHarness(t, true, true)
	hr.seedThread(t)
	_ = hr.inbox.SetInboxThreadArchived(testThread, true)
	hr.resend.received["e9"] = emailx.Received{ID: "e9", Text: "one more thing"}

	hr.webhook("e9", "sam@example.com", replyAddress)
	for _, m := range hr.inbox.all() {
		if m.ArchivedAt != 0 {
			t.Errorf("the thread must be open again: %+v", m)
		}
	}
}

func TestAdminInboxNeedsSignIn(t *testing.T) {
	hr := newHarness(t, true, true)
	hr.seedThread(t)
	for _, tc := range []struct{ method, path string }{
		{http.MethodGet, "/api/admin/inbox"},
		{http.MethodPost, "/api/admin/inbox/" + testThread + "/reply"},
		{http.MethodPost, "/api/admin/inbox/" + testThread + "/read"},
		{http.MethodPost, "/api/admin/inbox/" + testThread + "/archive"},
		{http.MethodDelete, "/api/admin/inbox/" + testThread},
		{http.MethodDelete, "/api/admin/inbox/archived"},
	} {
		if rec := hr.do(tc.method, tc.path, `{"body":"hi"}`); rec.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: status = %d", tc.method, tc.path, rec.Code)
		}
	}
}

func TestAdminInboxListsThreads(t *testing.T) {
	hr := newHarness(t, true, true)
	hr.signedIn = true
	hr.seedThread(t)
	rec := hr.do(http.MethodGet, "/api/admin/inbox", "")
	var threads []dto.InboxThreadResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &threads); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK || len(threads) != 1 || threads[0].ThreadID != testThread || !threads[0].Unread || len(threads[0].Messages) != 1 {
		t.Errorf("status = %d, threads = %+v", rec.Code, threads)
	}
}

func TestAdminReplyIsEmailedAndKeptInTheThread(t *testing.T) {
	hr := newHarness(t, true, true)
	hr.signedIn = true
	hr.seedThread(t)
	hr.inbox.messages[0].MessageID = "<ours-1@mail.example>"

	rec := hr.do(http.MethodPost, "/api/admin/inbox/"+testThread+"/reply", `{"body":"  Yes, all ages.  "}`)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
	}

	sent := hr.resend.messages()
	if len(sent) != 1 {
		t.Fatalf("sent = %d", len(sent))
	}
	to, _ := sent[0]["to"].([]any)
	replyTo, _ := sent[0]["reply_to"].([]any)
	headers, _ := sent[0]["headers"].(map[string]any)
	if len(to) != 1 || to[0] != "sam@example.com" || sent[0]["subject"] != "Re: Contact form message" {
		t.Errorf("sent = %+v", sent[0])
	}
	// the email quotes what is being answered under the reply
	text, _ := sent[0]["text"].(string)
	if !strings.HasPrefix(text, "Yes, all ages.\n\nOn ") || !strings.HasSuffix(text, "Sam <sam@example.com> wrote:\n> Is Friday all ages?") {
		t.Errorf("text = %q", text)
	}
	if len(replyTo) != 1 || replyTo[0] != replyAddress {
		t.Errorf("their answer must come back to the thread: %v", replyTo)
	}
	if headers["In-Reply-To"] != "<ours-1@mail.example>" {
		t.Errorf("headers = %v", headers)
	}

	saved := hr.inbox.all()
	if len(saved) == 2 && saved[1].Body != "Yes, all ages." {
		t.Errorf("the conversation keeps only the reply, got %q", saved[1].Body)
	}
	if len(saved) != 2 || saved[1].Source != models.InboxSourceAdminReply || saved[1].ThreadID != testThread || saved[1].FromName != "admin@example.com" {
		t.Fatalf("saved = %+v", saved)
	}
	if saved[0].ReadAt == 0 {
		t.Error("answering reads the thread")
	}
}

func TestAdminReplyValidation(t *testing.T) {
	hr := newHarness(t, true, true)
	hr.signedIn = true
	hr.seedThread(t)
	path := "/api/admin/inbox/" + testThread + "/reply"
	if rec := hr.do(http.MethodPost, path, `{"body":"   "}`); rec.Code != http.StatusBadRequest {
		t.Errorf("empty: status = %d", rec.Code)
	}
	if rec := hr.do(http.MethodPost, path, `{"body":"`+strings.Repeat("a", maxReply+1)+`"}`); rec.Code != http.StatusBadRequest {
		t.Errorf("long: status = %d", rec.Code)
	}
	if rec := hr.do(http.MethodPost, "/api/admin/inbox/not-a-uuid/reply", `{"body":"hi"}`); rec.Code != http.StatusBadRequest {
		t.Errorf("bad id: status = %d", rec.Code)
	}
	other := "/api/admin/inbox/11111111-2222-4333-8444-555555555555/reply"
	if rec := hr.do(http.MethodPost, other, `{"body":"hi"}`); rec.Code != http.StatusNotFound {
		t.Errorf("unknown thread: status = %d", rec.Code)
	}
	if len(hr.resend.messages()) != 0 {
		t.Error("nothing may be sent")
	}
}

func TestAdminReplyThatFailsToSendIsNotSaved(t *testing.T) {
	hr := newHarness(t, true, true)
	hr.signedIn = true
	hr.seedThread(t)
	hr.resend.status = http.StatusInternalServerError
	rec := hr.do(http.MethodPost, "/api/admin/inbox/"+testThread+"/reply", `{"body":"Yes"}`)
	if rec.Code != http.StatusBadGateway || len(hr.inbox.all()) != 1 {
		t.Errorf("status = %d, saved = %d", rec.Code, len(hr.inbox.all()))
	}
}

func TestAdminReplyNeedsEmail(t *testing.T) {
	hr := newHarness(t, false, false)
	hr.signedIn = true
	hr.seedThread(t)
	if rec := hr.do(http.MethodPost, "/api/admin/inbox/"+testThread+"/reply", `{"body":"Yes"}`); rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d", rec.Code)
	}
}

func TestAdminReadAndArchive(t *testing.T) {
	hr := newHarness(t, true, true)
	hr.signedIn = true
	hr.seedThread(t)
	base := "/api/admin/inbox/" + testThread

	if rec := hr.do(http.MethodPost, base+"/read", ""); rec.Code != http.StatusNoContent || hr.inbox.all()[0].ReadAt == 0 {
		t.Errorf("read: status = %d", rec.Code)
	}
	if rec := hr.do(http.MethodPost, base+"/archive", ""); rec.Code != http.StatusNoContent || hr.inbox.all()[0].ArchivedAt == 0 {
		t.Errorf("archive: status = %d", rec.Code)
	}
	if rec := hr.do(http.MethodPost, base+"/unarchive", ""); rec.Code != http.StatusNoContent || hr.inbox.all()[0].ArchivedAt != 0 {
		t.Errorf("unarchive: status = %d", rec.Code)
	}
}

func TestAdminComposeStartsAThreadAndIsAnswerable(t *testing.T) {
	hr := newHarness(t, true, true)
	hr.signedIn = true

	rec := hr.do(http.MethodPost, "/api/admin/inbox", `{"to":"venue@example.org","subject":"Booking","body":"  Hello!  "}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
	}
	var created struct{ ThreadID string }
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil || created.ThreadID == "" {
		t.Fatalf("response = %s", rec.Body)
	}

	sent := hr.resend.messages()
	if len(sent) != 1 {
		t.Fatalf("sent = %d", len(sent))
	}
	to, _ := sent[0]["to"].([]any)
	replyTo, _ := sent[0]["reply_to"].([]any)
	if len(to) != 1 || to[0] != "venue@example.org" || sent[0]["subject"] != "Booking" || sent[0]["text"] != "Hello!" {
		t.Errorf("sent = %+v", sent[0])
	}
	if len(replyTo) != 1 || !strings.HasPrefix(replyTo[0].(string), "reply-"+created.ThreadID+"@") {
		t.Errorf("their answer must come back to the thread: %v", replyTo)
	}

	saved := hr.inbox.all()
	if len(saved) != 1 || saved[0].ThreadID != created.ThreadID || saved[0].ToEmail != "venue@example.org" || saved[0].Source != models.InboxSourceAdminReply {
		t.Fatalf("saved = %+v", saved)
	}

	// before they answer, a second message goes to the same person
	rec = hr.do(http.MethodPost, "/api/admin/inbox/"+created.ThreadID+"/reply", `{"body":"Following up"}`)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("reply status = %d, body = %s", rec.Code, rec.Body)
	}
	sent = hr.resend.messages()
	to, _ = sent[1]["to"].([]any)
	if len(to) != 1 || to[0] != "venue@example.org" || sent[1]["subject"] != "Re: Booking" {
		t.Errorf("follow-up = %+v", sent[1])
	}
}

func TestAdminComposeValidation(t *testing.T) {
	hr := newHarness(t, true, true)
	hr.signedIn = true
	long := strings.Repeat("a", maxReply+1)
	cases := map[string]string{
		"no address":      `{"to":"","subject":"s","body":"b"}`,
		"not an address":  `{"to":"nobody","subject":"s","body":"b"}`,
		"display name":    `{"to":"Sam <sam@example.org>","subject":"s","body":"b"}`,
		"two addresses":   `{"to":"a@example.org, b@example.org","subject":"s","body":"b"}`,
		"the site itself": `{"to":"` + "HELLO@MAIL.EXAMPLE.COM" + `","subject":"s","body":"b"}`,
		"no subject":      `{"to":"a@example.org","subject":" ","body":"b"}`,
		"subject lines":   `{"to":"a@example.org","subject":"s\nBcc: x@example.org","body":"b"}`,
		"long subject":    `{"to":"a@example.org","subject":"` + strings.Repeat("s", maxComposeSubject+1) + `","body":"b"}`,
		"no body":         `{"to":"a@example.org","subject":"s","body":"  "}`,
		"long body":       `{"to":"a@example.org","subject":"s","body":"` + long + `"}`,
	}
	for name, body := range cases {
		if rec := hr.do(http.MethodPost, "/api/admin/inbox", body); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d", name, rec.Code)
		}
	}
	if len(hr.resend.messages()) != 0 || len(hr.inbox.all()) != 0 {
		t.Error("nothing may be sent or saved")
	}
}

func TestAdminComposeNeedsSignInAndEmail(t *testing.T) {
	hr := newHarness(t, true, true)
	body := `{"to":"a@example.org","subject":"s","body":"b"}`
	if rec := hr.do(http.MethodPost, "/api/admin/inbox", body); rec.Code != http.StatusUnauthorized {
		t.Errorf("signed out: status = %d", rec.Code)
	}
	off := newHarness(t, false, false)
	off.signedIn = true
	if rec := off.do(http.MethodPost, "/api/admin/inbox", body); rec.Code != http.StatusServiceUnavailable {
		t.Errorf("no email: status = %d", rec.Code)
	}
}

func TestAdminComposeThatFailsToSendIsNotSaved(t *testing.T) {
	hr := newHarness(t, true, true)
	hr.signedIn = true
	hr.resend.status = http.StatusInternalServerError
	rec := hr.do(http.MethodPost, "/api/admin/inbox", `{"to":"a@example.org","subject":"s","body":"b"}`)
	if rec.Code != http.StatusBadGateway || len(hr.inbox.all()) != 0 {
		t.Errorf("status = %d, saved = %d", rec.Code, len(hr.inbox.all()))
	}
}

// an outreach sent through the stand-in Resend, whose id for it is sent_1
func (hr *harness) sendOutreach(t *testing.T) {
	t.Helper()
	hr.signedIn = true
	if rec := hr.do(http.MethodPost, "/api/admin/inbox", `{"to":"venue@example.org","subject":"Booking","body":"Hello"}`); rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
	}
	if got := hr.inbox.all(); len(got) != 1 || got[0].SentID == nil || *got[0].SentID != "sent_1" {
		t.Fatalf("the id Resend gave must be kept: %+v", got)
	}
}

func TestDeliveredIsKeptOnTheMessage(t *testing.T) {
	hr := newHarness(t, true, true)
	hr.sendOutreach(t)

	delivered := `{"type":"email.delivered","created_at":"2026-02-22T23:41:12.126Z","data":{"email_id":"sent_1","to":["venue@example.org"]}}`
	later := `{"type":"email.delivered","created_at":"2026-02-23T08:00:00Z","data":{"email_id":"sent_1"}}`
	for _, body := range []string{delivered, later} {
		if rec := hr.signedWebhook(body, time.Now()); rec.Code != http.StatusOK {
			t.Fatalf("status = %d", rec.Code)
		}
	}
	if want := time.Date(2026, 2, 22, 23, 41, 12, 0, time.UTC).Unix(); hr.inbox.all()[0].DeliveredAt != want {
		t.Errorf("only the first delivery counts: delivered at = %d (want %d)", hr.inbox.all()[0].DeliveredAt, want)
	}

	rec := hr.do(http.MethodGet, "/api/admin/inbox", "")
	var threads []struct{ Messages []struct{ DeliveredAt int64 } }
	if err := json.Unmarshal(rec.Body.Bytes(), &threads); err != nil || len(threads) != 1 || threads[0].Messages[0].DeliveredAt == 0 {
		t.Errorf("the thread must show it: %s", rec.Body)
	}
}

func TestOpenedIsIgnored(t *testing.T) {
	hr := newHarness(t, true, true)
	hr.sendOutreach(t)
	if rec := hr.signedWebhook(`{"type":"email.opened","data":{"email_id":"sent_1"}}`, time.Now()); rec.Code != http.StatusOK {
		t.Errorf("status = %d", rec.Code)
	}
}

func TestEventsOfOtherEmailsAreAccepted(t *testing.T) {
	hr := newHarness(t, true, true)
	hr.sendOutreach(t)
	// the site's own notices are reported too, and so is an event without an id
	for _, body := range []string{
		`{"type":"email.delivered","data":{"email_id":"some-notice"}}`,
		`{"type":"email.delivered","data":{}}`,
	} {
		if rec := hr.signedWebhook(body, time.Now()); rec.Code != http.StatusOK {
			t.Errorf("%s: status = %d", body, rec.Code)
		}
	}
	if m := hr.inbox.all()[0]; m.DeliveredAt != 0 {
		t.Errorf("message = %+v", m)
	}
}

func TestDeliveredFailsSoResendTriesAgain(t *testing.T) {
	hr := newHarness(t, true, true)
	hr.sendOutreach(t)
	hr.inbox.failCreate = true
	if rec := hr.signedWebhook(`{"type":"email.delivered","data":{"email_id":"sent_1"}}`, time.Now()); rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d", rec.Code)
	}
}

func TestAdminInboxDeleteRemovesAnyConversation(t *testing.T) {
	hr := newHarness(t, true, true)
	hr.signedIn = true
	hr.seedThread(t)
	other := "11111111-1111-4111-8111-111111111111"
	if err := hr.inbox.CreateInboxMessage(&models.InboxMessage{ThreadID: other, Source: models.InboxSourceContactForm, FromEmail: "kim@example.com", Body: "keep me"}); err != nil {
		t.Fatal(err)
	}

	// an open, unread conversation can be deleted, and only that one
	if rec := hr.do(http.MethodDelete, "/api/admin/inbox/"+testThread, ""); rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}
	if left := hr.inbox.all(); len(left) != 1 || left[0].ThreadID != other {
		t.Errorf("messages left = %+v", left)
	}
	if rec := hr.do(http.MethodDelete, "/api/admin/inbox/"+testThread, ""); rec.Code != http.StatusNotFound {
		t.Errorf("deleting it again: status = %d, want 404", rec.Code)
	}
	if rec := hr.do(http.MethodDelete, "/api/admin/inbox/not-a-uuid", ""); rec.Code != http.StatusBadRequest {
		t.Errorf("bad id: status = %d, want 400", rec.Code)
	}
}

func TestAdminInboxDeleteArchivedKeepsTheRest(t *testing.T) {
	hr := newHarness(t, true, true)
	hr.signedIn = true
	hr.seedThread(t)
	second := "22222222-2222-4222-8222-222222222222"
	third := "33333333-3333-4333-8333-333333333333"
	for _, id := range []string{second, third} {
		if err := hr.inbox.CreateInboxMessage(&models.InboxMessage{ThreadID: id, Source: models.InboxSourceContactForm, FromEmail: "kim@example.com", Body: "hello"}); err != nil {
			t.Fatal(err)
		}
	}
	_ = hr.inbox.SetInboxThreadArchived(testThread, true)
	_ = hr.inbox.SetInboxThreadArchived(second, true)
	// a new message since the archiving brings the whole conversation back, so it must stay
	if err := hr.inbox.CreateInboxMessage(&models.InboxMessage{ThreadID: second, Source: models.InboxSourceEmail, FromEmail: "kim@example.com", Body: "again"}); err != nil {
		t.Fatal(err)
	}

	rec := hr.do(http.MethodDelete, "/api/admin/inbox/archived", "")
	var res dto.InboxDeleteArchivedResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK || res.Deleted != 1 {
		t.Fatalf("status = %d, deleted = %d", rec.Code, res.Deleted)
	}
	if left := hr.inbox.all(); len(left) != 3 {
		t.Errorf("messages left = %+v, want the two of the second conversation and the third", left)
	}
}
