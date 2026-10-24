package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/borgehl/bruttice/packages/go-pkg/configs"
	"github.com/borgehl/bruttice/packages/go-pkg/models"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/emailx"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/logx"
)

const (
	testThread = "0b8c1f3e-4d6a-4e1b-9f3a-2c7d8e9f0a1b"
	testSecret = "whsec_YS1zZWNyZXQta2V5LW9mLXNvbWUtbGVuZ3Ro"
)

// fakeInbox keeps the inbox in memory.
type fakeInbox struct {
	mu         sync.Mutex
	messages   []models.InboxMessage
	failCreate bool
}

func (f *fakeInbox) CreateInboxMessage(m *models.InboxMessage) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.failCreate {
		return fmt.Errorf("database is down")
	}
	m.ID = int64(len(f.messages) + 1)
	m.CreatedAt = time.Now().Unix()
	f.messages = append(f.messages, *m)
	return nil
}

func (f *fakeInbox) InboxProviderIDSeen(id string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, m := range f.messages {
		if m.ProviderID != nil && *m.ProviderID == id {
			return true, nil
		}
	}
	return false, nil
}

func (f *fakeInbox) FindInboxMessageByMessageID(id string) (*models.InboxMessage, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, m := range f.messages {
		if m.MessageID == id {
			m := m
			return &m, nil
		}
	}
	return nil, nil
}

func (f *fakeInbox) mark(sentID string, at int64, field func(*models.InboxMessage) *int64) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.failCreate {
		return fmt.Errorf("database is down")
	}
	for i := range f.messages {
		m := &f.messages[i]
		if m.SentID != nil && *m.SentID == sentID && *field(m) == 0 {
			*field(m) = at
		}
	}
	return nil
}

func (f *fakeInbox) MarkInboxDelivered(sentID string, at int64) error {
	return f.mark(sentID, at, func(m *models.InboxMessage) *int64 { return &m.DeliveredAt })
}

func (f *fakeInbox) GetInboxThread(id string) ([]models.InboxMessage, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := []models.InboxMessage{}
	for _, m := range f.messages {
		if m.ThreadID == id {
			out = append(out, m)
		}
	}
	return out, nil
}

func (f *fakeInbox) ListInboxMessages(includeArchived bool) ([]models.InboxMessage, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := []models.InboxMessage{}
	for _, m := range f.messages {
		if includeArchived || m.ArchivedAt == 0 {
			out = append(out, m)
		}
	}
	return out, nil
}

func (f *fakeInbox) CountUnreadInboxThreads() (int64, error) { return 0, nil }

func (f *fakeInbox) MarkInboxThreadRead(id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for i := range f.messages {
		if f.messages[i].ThreadID == id && f.messages[i].ReadAt == 0 {
			f.messages[i].ReadAt = 1
		}
	}
	return nil
}

func (f *fakeInbox) SetInboxThreadArchived(id string, archived bool) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for i := range f.messages {
		if f.messages[i].ThreadID == id {
			f.messages[i].ArchivedAt = 0
			if archived {
				f.messages[i].ArchivedAt = 1
				f.messages[i].ReadAt = 1
			}
		}
	}
	return nil
}

func (f *fakeInbox) all() []models.InboxMessage {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]models.InboxMessage{}, f.messages...)
}

// fakeResend stands in for api.resend.com: it records what is sent and serves received emails.
type fakeResend struct {
	mu       sync.Mutex
	sent     []map[string]any
	received map[string]emailx.Received
	status   int
}

func (f *fakeResend) handler(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/emails/receiving/") {
		rec, ok := f.received[strings.TrimPrefix(r.URL.Path, "/emails/receiving/")]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(rec)
		return
	}
	var body map[string]any
	_ = json.NewDecoder(r.Body).Decode(&body)
	if f.status != 0 {
		w.WriteHeader(f.status)
		return
	}
	f.sent = append(f.sent, body)
	_, _ = fmt.Fprintf(w, `{"id":"sent_%d"}`, len(f.sent))
}

func (f *fakeResend) messages() []map[string]any {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]map[string]any{}, f.sent...)
}

// waitForSent waits for the notice the handlers send in the background.
func (f *fakeResend) waitForSent(t *testing.T, n int) []map[string]any {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if got := f.messages(); len(got) >= n {
			return got
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("expected %d sent emails, got %d", n, len(f.messages()))
	return nil
}

// The root admin's address is a global read by the notices handlers send in the background, so
// it is set once before any test runs.
func TestMain(m *testing.M) {
	configs.SetRootEmail("root@example.com")
	os.Exit(m.Run())
}

// fakeUsers remembers the names set.
type fakeUsers struct {
	mu    sync.Mutex
	names map[int64]string
	fail  bool
}

func (f *fakeUsers) SetUserName(id int64, name string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.fail {
		return fmt.Errorf("database is down")
	}
	if f.names == nil {
		f.names = map[int64]string{}
	}
	f.names[id] = name
	return nil
}

type harness struct {
	mux      *http.ServeMux
	resend   *fakeResend
	inbox    *fakeInbox
	users    *fakeUsers
	signedIn bool
	// the name the signed in admin has
	adminName string
}

// newHarness builds the inbox routes over fakes. enabled is whether email is set up, inbound
// whether the webhook has a signing secret.
func newHarness(t *testing.T, enabled, inbound bool) *harness {
	t.Helper()
	hr := &harness{resend: &fakeResend{received: map[string]emailx.Received{}}, inbox: &fakeInbox{}, users: &fakeUsers{}}
	server := httptest.NewServer(http.HandlerFunc(hr.resend.handler))
	t.Cleanup(server.Close)

	cfg := emailx.Config{Host: "https://events.example.com", SiteName: "Events", BaseURL: server.URL}
	if enabled {
		cfg.APIKey, cfg.From = "re_key", "hello@mail.example.com"
	}
	if inbound {
		cfg.WebhookSecret = testSecret
	}
	h := &APIHandler{
		logger:       logx.NewLogger(nil, nil),
		email:        emailx.NewEmailClient(logx.NewLogger(nil, nil), cfg),
		contactLimit: newContactLimiter(),
		inbox:        hr.inbox,
		users:        hr.users,
		authenticate: func(*http.Request) *models.User {
			if hr.signedIn {
				return &models.User{BaseModel: models.BaseModel{ID: 7}, Email: "admin@example.com", Name: hr.adminName}
			}
			return nil
		},
	}
	hr.mux = http.NewServeMux()
	hr.mux.HandleFunc("POST /api/contact", h.handleContact)
	hr.mux.HandleFunc("POST /api/webhooks/resend", h.handleResendWebhook)
	hr.mux.HandleFunc("PATCH /api/admin/settings", h.handleAdminSettings)
	hr.mux.HandleFunc("GET /api/admin/inbox", h.handleAdminInbox)
	hr.mux.HandleFunc("POST /api/admin/inbox", h.handleAdminInboxCompose)
	hr.mux.HandleFunc("POST /api/admin/inbox/{threadId}/reply", h.handleAdminInboxReply)
	hr.mux.HandleFunc("POST /api/admin/inbox/{threadId}/read", h.handleAdminInboxRead)
	hr.mux.HandleFunc("POST /api/admin/inbox/{threadId}/archive", h.handleAdminInboxArchive(true))
	hr.mux.HandleFunc("POST /api/admin/inbox/{threadId}/unarchive", h.handleAdminInboxArchive(false))
	return hr
}

func (hr *harness) do(method, path, body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.RemoteAddr = "203.0.113.7:4000"
	hr.mux.ServeHTTP(rec, req)
	return rec
}

// webhook posts an email.received event the way Resend does, signed.
func (hr *harness) webhook(emailID, from string, to ...string) *httptest.ResponseRecorder {
	toJSON, _ := json.Marshal(to)
	body := fmt.Sprintf(`{"type":"email.received","data":{"email_id":%q,"from":%q,"to":%s,"subject":"Re: Hi","message_id":"<in-1@mail.example>"}}`, emailID, from, toJSON)
	return hr.signedWebhook(body, time.Now())
}

func (hr *harness) signedWebhook(body string, at time.Time) *httptest.ResponseRecorder {
	ts := strconv.FormatInt(at.Unix(), 10)
	key, _ := base64.StdEncoding.DecodeString(strings.TrimPrefix(testSecret, "whsec_"))
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte("msg_1." + ts + "." + body))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/webhooks/resend", strings.NewReader(body))
	req.Header.Set("svix-id", "msg_1")
	req.Header.Set("svix-timestamp", ts)
	req.Header.Set("svix-signature", "v1,"+base64.StdEncoding.EncodeToString(mac.Sum(nil)))
	hr.mux.ServeHTTP(rec, req)
	return rec
}

// a conversation that exists already, started by a contact form message
func (hr *harness) seedThread(t *testing.T) {
	t.Helper()
	if err := hr.inbox.CreateInboxMessage(&models.InboxMessage{
		ThreadID: testThread, Source: models.InboxSourceContactForm, FromName: "Sam",
		FromEmail: "sam@example.com", Subject: "Contact form message", Body: "Is Friday all ages?",
	}); err != nil {
		t.Fatal(err)
	}
}
