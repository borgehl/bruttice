package api

import (
	"net/http"
	"strings"
	"testing"

	"github.com/borgehl/bruttice/packages/go-pkg/models"
)

const validContact = `{"name":"Sam","email":"sam@example.com","message":"Is the Friday show all ages?"}`

func TestContactIsSavedInTheInboxAndTheAdminIsTold(t *testing.T) {
	hr := newHarness(t, true, true)
	rec := hr.do(http.MethodPost, "/api/contact", validContact)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
	}

	saved := hr.inbox.all()
	if len(saved) != 1 {
		t.Fatalf("saved = %d messages", len(saved))
	}
	m := saved[0]
	if m.Source != models.InboxSourceContactForm || m.FromEmail != "sam@example.com" || m.Body != "Is the Friday show all ages?" || m.ThreadID == "" {
		t.Errorf("message = %+v", m)
	}

	sent := hr.resend.waitForSent(t, 1)[0]
	to, _ := sent["to"].([]any)
	if len(to) != 1 || to[0] != "root@example.com" {
		t.Errorf("the notice goes to the root admin, not the site's own address: %v", to)
	}
	if _, ok := sent["reply_to"]; ok {
		t.Errorf("the notice has no reply-to: the answer is written in the inbox")
	}
}

func TestContactIsUnavailableWithoutEmail(t *testing.T) {
	hr := newHarness(t, false, false)
	if rec := hr.do(http.MethodPost, "/api/contact", validContact); rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d", rec.Code)
	}
	if len(hr.inbox.all()) != 0 {
		t.Error("nothing may be saved")
	}
}

func TestContactValidation(t *testing.T) {
	hr := newHarness(t, true, true)
	for name, body := range map[string]string{
		"no name":       `{"name":"  ","email":"sam@example.com","message":"A long enough message"}`,
		"long name":     `{"name":"` + strings.Repeat("a", 101) + `","email":"sam@example.com","message":"A long enough message"}`,
		"bad email":     `{"name":"Sam","email":"sam@","message":"A long enough message"}`,
		"email newline": `{"name":"Sam","email":"sam@example.com\nBcc: x@example.com","message":"A long enough message"}`,
		"short message": `{"name":"Sam","email":"sam@example.com","message":"hi"}`,
		"long message":  `{"name":"Sam","email":"sam@example.com","message":"` + strings.Repeat("a", 5001) + `"}`,
		"not json":      `nope`,
	} {
		if rec := hr.do(http.MethodPost, "/api/contact", body); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d, want 400", name, rec.Code)
		}
	}
	if len(hr.inbox.all()) != 0 || len(hr.resend.messages()) != 0 {
		t.Error("an invalid message must not be saved or announced")
	}
}

func TestContactHoneypotLooksLikeSuccessButSavesNothing(t *testing.T) {
	hr := newHarness(t, true, true)
	rec := hr.do(http.MethodPost, "/api/contact", `{"name":"Bot","email":"bot@example.com","message":"Buy things now please","website":"http://spam.example"}`)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d", rec.Code)
	}
	if len(hr.inbox.all()) != 0 {
		t.Error("a bot's message must not be saved")
	}
}

func TestContactRateLimit(t *testing.T) {
	hr := newHarness(t, true, true)
	for i := 0; i < 5; i++ {
		if rec := hr.do(http.MethodPost, "/api/contact", validContact); rec.Code != http.StatusOK {
			t.Fatalf("message %d: status = %d", i+1, rec.Code)
		}
	}
	if rec := hr.do(http.MethodPost, "/api/contact", validContact); rec.Code != http.StatusTooManyRequests {
		t.Errorf("status = %d, want 429", rec.Code)
	}
	if len(hr.inbox.all()) != 5 {
		t.Errorf("saved = %d, want 5", len(hr.inbox.all()))
	}
}

// The notice is only a courtesy. A message must not be lost because Resend is down.
func TestContactSurvivesAFailedNotice(t *testing.T) {
	hr := newHarness(t, true, true)
	hr.resend.status = http.StatusInternalServerError
	if rec := hr.do(http.MethodPost, "/api/contact", validContact); rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if len(hr.inbox.all()) != 1 {
		t.Error("the message must be saved")
	}
}

func TestContactReportsAFailedSave(t *testing.T) {
	hr := newHarness(t, true, true)
	hr.inbox.failCreate = true
	rec := hr.do(http.MethodPost, "/api/contact", validContact)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "database") {
		t.Errorf("the visitor must not see why: %s", rec.Body)
	}
}
