package emailx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/borgehl/bruttice/packages/go-pkg/configs"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/logx"
)

func testClient(t *testing.T, cfg Config, handler http.HandlerFunc) *EmailClient {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	c := NewEmailClient(logx.NewLogger(nil, nil), cfg)
	c.baseURL = server.URL
	return c
}

func TestSend(t *testing.T) {
	var got sendRequest
	var auth, path string
	c := testClient(t, Config{APIKey: "re_key", From: "hello@example.com", SiteName: "Do It Together"}, func(w http.ResponseWriter, r *http.Request) {
		auth, path = r.Header.Get("Authorization"), r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Error(err)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"abc"}`))
	})

	id, err := c.Send(Message{To: []string{"you@example.com"}, Subject: "Hi", Text: "Body", ReplyTo: "them@example.com", Headers: map[string]string{"In-Reply-To": "<a@b>"}})
	if err != nil {
		t.Fatal(err)
	}
	if id != "abc" {
		t.Errorf("id = %q", id)
	}
	if got.Headers["In-Reply-To"] != "<a@b>" {
		t.Errorf("headers = %v", got.Headers)
	}
	if auth != "Bearer re_key" || path != "/emails" {
		t.Errorf("auth = %q, path = %q", auth, path)
	}
	if got.From != `"Do It Together" <hello@example.com>` {
		t.Errorf("from = %q", got.From)
	}
	if got.Subject != "Hi" || got.Text != "Body" || len(got.To) != 1 || got.To[0] != "you@example.com" {
		t.Errorf("request = %+v", got)
	}
	if len(got.ReplyTo) != 1 || got.ReplyTo[0] != "them@example.com" {
		t.Errorf("reply_to = %v", got.ReplyTo)
	}
}

func TestSendWithoutReplyTo(t *testing.T) {
	var raw map[string]any
	c := testClient(t, Config{APIKey: "k", From: "hello@example.com"}, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&raw)
	})
	if _, err := c.Send(Message{To: []string{"a@example.com"}, Subject: "s", Text: "t"}); err != nil {
		t.Fatal(err)
	}
	if _, ok := raw["reply_to"]; ok {
		t.Errorf("reply_to must be left out: %v", raw)
	}
}

func TestSendReportsResendsReason(t *testing.T) {
	c := testClient(t, Config{APIKey: "re_secret", From: "hello@example.com"}, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"name":"validation_error","message":"The example.com domain is not verified."}`))
	})
	_, err := c.Send(Message{To: []string{"a@example.com"}, Subject: "s", Text: "t"})
	if err == nil || !strings.Contains(err.Error(), "403") || !strings.Contains(err.Error(), "not verified") {
		t.Errorf("err = %v", err)
	}
	if strings.Contains(err.Error(), "re_secret") {
		t.Errorf("the key must not appear in an error: %v", err)
	}
}

func TestSendDoesNothingWhenNotConfigured(t *testing.T) {
	for name, cfg := range map[string]Config{
		"no key":     {From: "hello@example.com"},
		"no address": {APIKey: "k"},
	} {
		called := false
		c := testClient(t, cfg, func(w http.ResponseWriter, r *http.Request) { called = true })
		if c.Enabled() {
			t.Errorf("%s: enabled", name)
		}
		if _, err := c.Send(Message{To: []string{"a@example.com"}}); err != nil || called {
			t.Errorf("%s: err = %v, called = %v", name, err, called)
		}
	}
}

func TestVerifyEventMessage(t *testing.T) {
	c := NewEmailClient(logx.NewLogger(nil, nil), Config{Host: "https://events.example.com"})
	m := c.VerifyEventMessage("Show", "a@example.com", "key123")
	if m.To[0] != "a@example.com" || !strings.Contains(m.Text, "https://events.example.com/verify-event?verifyKey=key123") {
		t.Errorf("message = %+v", m)
	}
	c = NewEmailClient(logx.NewLogger(nil, nil), Config{Host: "localhost:8080"})
	if !strings.Contains(c.VerifyEventMessage("S", "a@example.com", "k").Text, "http://localhost:8080/verify-event") {
		t.Error("a host without a scheme must get http://")
	}
}

func TestInboxNoticeMessage(t *testing.T) {
	configs.SetRootEmail("root@example.com")
	c := NewEmailClient(logx.NewLogger(nil, nil), Config{From: "hello@example.com", Host: "https://events.example.com"})
	m := c.InboxNoticeMessage("Sam\r\nBcc: x@example.com", "sam@example.com", "Hi", "Hello there")
	if m.To[0] != "root@example.com" || m.ReplyTo != "" {
		t.Errorf("message = %+v", m)
	}
	if strings.ContainsAny(m.Subject, "\r\n") || m.Subject != "New message from Sam Bcc: x@example.com" {
		t.Errorf("subject = %q", m.Subject)
	}
	if !strings.Contains(m.Text, "Hello there") || !strings.Contains(m.Text, "https://events.example.com/admin") {
		t.Errorf("text = %q", m.Text)
	}
}

func TestSendUsesTheSendersName(t *testing.T) {
	var got sendRequest
	c := testClient(t, Config{APIKey: "k", From: "hello@example.com", SiteName: "Do It Together"}, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&got)
		_, _ = w.Write([]byte(`{"id":"abc"}`))
	})
	if _, err := c.Send(Message{To: []string{"you@example.com"}, FromName: "Pat Lee", Subject: "Hi", Text: "Body"}); err != nil {
		t.Fatal(err)
	}
	if got.From != `"Pat Lee" <hello@example.com>` {
		t.Errorf("from = %q", got.From)
	}
}

func TestReplyMessage(t *testing.T) {
	c := NewEmailClient(logx.NewLogger(nil, nil), Config{From: "hello@mail.example.com"})
	m := c.ReplyMessage("sam@example.com", "Pat", "Re: Hi", "Yes", "0b8c1f3e-4d6a-4e1b-9f3a-2c7d8e9f0a1b", "<abc@x>")
	if m.ReplyTo != "reply-0b8c1f3e-4d6a-4e1b-9f3a-2c7d8e9f0a1b@mail.example.com" {
		t.Errorf("reply-to = %q", m.ReplyTo)
	}
	if m.FromName != "Pat" {
		t.Errorf("from name = %q", m.FromName)
	}
	if m.Headers["In-Reply-To"] != "<abc@x>" || m.Headers["References"] != "<abc@x>" {
		t.Errorf("headers = %v", m.Headers)
	}
	if c.ReplyMessage("a@b.c", "", "s", "t", "id", "").Headers != nil {
		t.Error("no headers without a Message-ID to answer")
	}
}

func TestErrorsMessageGoesToTheRootAdmin(t *testing.T) {
	configs.SetRootEmail("root@example.com")
	c := NewEmailClient(logx.NewLogger(nil, nil), Config{})
	m := c.ErrorsMessage([]error{errString("a: boom")})
	if m.To[0] != "root@example.com" || !strings.Contains(m.Text, "0: a: boom") || !strings.Contains(m.Subject, "Errors: 1") {
		t.Errorf("message = %+v", m)
	}
}

type errString string

func (e errString) Error() string { return string(e) }
