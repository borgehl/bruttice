package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/borgehl/bruttice/packages/go-pkg/platform/emailx"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/logx"
)

// fakeResend stands in for api.resend.com and records what was sent.
type fakeResend struct {
	mu     sync.Mutex
	sent   []map[string]any
	status int
}

func (f *fakeResend) handler(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	_ = json.NewDecoder(r.Body).Decode(&body)
	f.mu.Lock()
	f.sent = append(f.sent, body)
	f.mu.Unlock()
	if f.status != 0 {
		w.WriteHeader(f.status)
	}
}

func contactServer(t *testing.T, enabled bool) (*http.ServeMux, *fakeResend) {
	t.Helper()
	fake := &fakeResend{}
	resend := httptest.NewServer(http.HandlerFunc(fake.handler))
	t.Cleanup(resend.Close)

	cfg := emailx.Config{Host: "https://events.example.com", SiteName: "Events", BaseURL: resend.URL}
	if enabled {
		cfg.APIKey, cfg.From = "re_key", "hello@example.com"
	}
	client := emailx.NewEmailClient(logx.NewLogger(nil, nil), cfg)

	h := &APIHandler{logger: logx.NewLogger(nil, nil), email: client, contactLimit: newContactLimiter()}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/contact", h.handleContact)
	return mux, fake
}

func postContact(mux *http.ServeMux, body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/contact", strings.NewReader(body))
	req.RemoteAddr = "203.0.113.7:4000"
	mux.ServeHTTP(rec, req)
	return rec
}

const validContact = `{"name":"Sam","email":"sam@example.com","message":"Is the Friday show all ages?"}`

func TestContactSendsToTheSiteWithTheVisitorAsReplyTo(t *testing.T) {
	mux, fake := contactServer(t, true)
	rec := postContact(mux, validContact)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
	}
	if len(fake.sent) != 1 {
		t.Fatalf("sent = %d messages", len(fake.sent))
	}
	got := fake.sent[0]
	to, _ := got["to"].([]any)
	replyTo, _ := got["reply_to"].([]any)
	if len(to) != 1 || to[0] != "hello@example.com" || len(replyTo) != 1 || replyTo[0] != "sam@example.com" {
		t.Errorf("to = %v, reply_to = %v", to, replyTo)
	}
	if !strings.Contains(got["text"].(string), "Is the Friday show all ages?") {
		t.Errorf("text = %v", got["text"])
	}
}

func TestContactIsUnavailableWithoutEmail(t *testing.T) {
	mux, fake := contactServer(t, false)
	if rec := postContact(mux, validContact); rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d", rec.Code)
	}
	if len(fake.sent) != 0 {
		t.Error("nothing may be sent")
	}
}

func TestContactValidation(t *testing.T) {
	mux, fake := contactServer(t, true)
	for name, body := range map[string]string{
		"no name":       `{"name":"  ","email":"sam@example.com","message":"A long enough message"}`,
		"long name":     `{"name":"` + strings.Repeat("a", 101) + `","email":"sam@example.com","message":"A long enough message"}`,
		"bad email":     `{"name":"Sam","email":"sam@","message":"A long enough message"}`,
		"email newline": `{"name":"Sam","email":"sam@example.com\nBcc: x@example.com","message":"A long enough message"}`,
		"short message": `{"name":"Sam","email":"sam@example.com","message":"hi"}`,
		"long message":  `{"name":"Sam","email":"sam@example.com","message":"` + strings.Repeat("a", 5001) + `"}`,
		"not json":      `nope`,
	} {
		if rec := postContact(mux, body); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d, want 400", name, rec.Code)
		}
	}
	if len(fake.sent) != 0 {
		t.Errorf("an invalid message must not be sent: %v", fake.sent)
	}
}

func TestContactHoneypotLooksLikeSuccessButSendsNothing(t *testing.T) {
	mux, fake := contactServer(t, true)
	rec := postContact(mux, `{"name":"Bot","email":"bot@example.com","message":"Buy things now please","website":"http://spam.example"}`)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d", rec.Code)
	}
	if len(fake.sent) != 0 {
		t.Error("a bot's message must not be sent")
	}
}

func TestContactRateLimit(t *testing.T) {
	mux, fake := contactServer(t, true)
	for i := 0; i < 5; i++ {
		if rec := postContact(mux, validContact); rec.Code != http.StatusOK {
			t.Fatalf("message %d: status = %d", i+1, rec.Code)
		}
	}
	if rec := postContact(mux, validContact); rec.Code != http.StatusTooManyRequests {
		t.Errorf("status = %d, want 429", rec.Code)
	}
	if len(fake.sent) != 5 {
		t.Errorf("sent = %d, want 5", len(fake.sent))
	}
}

func TestContactReportsAFailedSend(t *testing.T) {
	mux, fake := contactServer(t, true)
	fake.status = http.StatusForbidden
	rec := postContact(mux, validContact)
	if rec.Code != http.StatusBadGateway {
		t.Errorf("status = %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "re_key") || strings.Contains(rec.Body.String(), "403") {
		t.Errorf("the visitor must not see why: %s", rec.Body)
	}
}
