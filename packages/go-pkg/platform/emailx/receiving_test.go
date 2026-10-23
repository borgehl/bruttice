package emailx

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/borgehl/bruttice/packages/go-pkg/platform/logx"
)

const threadID = "0b8c1f3e-4d6a-4e1b-9f3a-2c7d8e9f0a1b"

func client(cfg Config) *EmailClient {
	return NewEmailClient(logx.NewLogger(nil, nil), cfg)
}

func TestReplyAddressRoundTrip(t *testing.T) {
	c := client(Config{From: "Hello@Mail.Example.com"})
	addr := c.ReplyAddress(threadID)
	if addr != "reply-"+threadID+"@Mail.Example.com" {
		t.Fatalf("address = %q", addr)
	}
	for _, form := range []string{addr, "Site <" + strings.ToLower(addr) + ">"} {
		if id, ok := c.ThreadFromAddress(form); !ok || id != threadID {
			t.Errorf("%q: id = %q, ok = %v", form, id, ok)
		}
	}
}

func TestThreadFromAddressRefusesOtherAddresses(t *testing.T) {
	c := client(Config{From: "hello@mail.example.com"})
	for _, addr := range []string{
		"hello@mail.example.com",
		"reply-" + threadID + "@other.example.com",
		"reply-not-a-uuid@mail.example.com",
		"reply-" + threadID + "x@mail.example.com",
		"nonsense",
	} {
		if id, ok := c.ThreadFromAddress(addr); ok {
			t.Errorf("%q must not be a thread address, got %q", addr, id)
		}
	}
}

func TestForUs(t *testing.T) {
	c := client(Config{From: "hello@mail.example.com"})
	if !c.ForUs("Hello <HELLO@mail.example.com>") || !c.ForUs("reply-"+threadID+"@mail.example.com") {
		t.Error("the contact address and reply addresses are ours")
	}
	if c.ForUs("someone@mail.example.com") || c.ForUs("hello@example.com") {
		t.Error("other addresses are not")
	}
}

func TestSender(t *testing.T) {
	if name, addr := Sender(`"Sam Jones" <sam@example.com>`); name != "Sam Jones" || addr != "sam@example.com" {
		t.Errorf("got %q, %q", name, addr)
	}
	if name, addr := Sender("sam@example.com"); name != "" || addr != "sam@example.com" {
		t.Errorf("got %q, %q", name, addr)
	}
}

func TestPlainText(t *testing.T) {
	if got := (Received{Text: "  Hi\r\nthere  "}).PlainText(); got != "Hi\nthere" {
		t.Errorf("text = %q", got)
	}
	html := `<div>Hello <b>Sam</b><br>Line two</div><script>alert(1)</script><p>Bye &amp; thanks</p>`
	got := (Received{HTML: html}).PlainText()
	if strings.ContainsAny(got, "<>") || strings.Contains(got, "alert") || !strings.Contains(got, "Bye & thanks") {
		t.Errorf("text = %q", got)
	}
}

func TestHeaderIgnoresCaseAndTakesTheFirstOfAList(t *testing.T) {
	r := Received{Headers: map[string]any{"In-Reply-To": "<a@b>", "references": []any{"<x@y>", "<z@y>"}}}
	if r.Header("in-reply-to") != "<a@b>" || r.Header("References") != "<x@y>" || r.Header("Missing") != "" {
		t.Errorf("headers = %v", r.Headers)
	}
}

func TestFetchReceived(t *testing.T) {
	var path string
	c := testClient(t, Config{APIKey: "k", From: "hello@example.com"}, func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.RequestURI()
		_, _ = w.Write([]byte(`{"id":"e1","from":"Sam <sam@example.com>","to":["hello@example.com"],"subject":"Hi","message_id":"<m@x>","text":"Body","html":"<p>Body</p>","headers":{"In-Reply-To":"<a@b>"}}`))
	})
	got, err := c.FetchReceived("e1")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(path, "/emails/receiving/e1") {
		t.Errorf("path = %q", path)
	}
	if got.Subject != "Hi" || got.Text != "Body" || got.Header("in-reply-to") != "<a@b>" || got.MessageID != "<m@x>" {
		t.Errorf("received = %+v", got)
	}
}

func TestFetchReceivedFailure(t *testing.T) {
	c := testClient(t, Config{APIKey: "k", From: "hello@example.com"}, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	if _, err := c.FetchReceived("nope"); err == nil {
		t.Error("a missing email must be an error")
	}
}

// sign makes the header Resend would send for a body.
func sign(secret, id, timestamp string, body []byte) string {
	key, _ := base64.StdEncoding.DecodeString(strings.TrimPrefix(secret, "whsec_"))
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(id + "." + timestamp + "."))
	mac.Write(body)
	return "v1," + base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func TestVerifyWebhook(t *testing.T) {
	secret := "whsec_" + base64.StdEncoding.EncodeToString([]byte("a-secret-key-of-some-length"))
	body := []byte(`{"type":"email.received"}`)
	now := time.Unix(1_800_000_000, 0)
	ts := "1800000000"
	good := sign(secret, "msg_1", ts, body)

	if err := verifyWebhook(secret, "msg_1", ts, good, body, now); err != nil {
		t.Errorf("a signed request must pass: %v", err)
	}
	if err := verifyWebhook(secret, "msg_1", ts, "v1,bogus "+good, body, now); err != nil {
		t.Errorf("one good signature among several must pass: %v", err)
	}

	for name, err := range map[string]error{
		"changed body":  verifyWebhook(secret, "msg_1", ts, good, []byte(`{"type":"other"}`), now),
		"changed id":    verifyWebhook(secret, "msg_2", ts, good, body, now),
		"bad signature": verifyWebhook(secret, "msg_1", ts, "v1,AAAA", body, now),
		"no signature":  verifyWebhook(secret, "msg_1", ts, "", body, now),
		"no secret":     verifyWebhook("", "msg_1", ts, good, body, now),
		"old":           verifyWebhook(secret, "msg_1", ts, good, body, now.Add(10*time.Minute)),
		"future":        verifyWebhook(secret, "msg_1", ts, good, body, now.Add(-10*time.Minute)),
		"bad timestamp": verifyWebhook(secret, "msg_1", "soon", good, body, now),
	} {
		if err == nil {
			t.Errorf("%s must be refused", name)
		}
	}
}

func TestInboundEnabled(t *testing.T) {
	if client(Config{APIKey: "k", From: "a@b.c"}).InboundEnabled() {
		t.Error("needs a webhook secret")
	}
	if !client(Config{APIKey: "k", From: "a@b.c", WebhookSecret: "whsec_x"}).InboundEnabled() {
		t.Error("key, address and secret enable it")
	}
	if client(Config{WebhookSecret: "whsec_x"}).InboundEnabled() {
		t.Error("needs sending to work too")
	}
}
