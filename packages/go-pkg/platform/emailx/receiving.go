package emailx

import (
	"fmt"
	"net/http"
	"net/mail"
	"net/url"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

// Received is an email that arrived at the site's domain, as Resend describes it.
type Received struct {
	ID        string   `json:"id"`
	From      string   `json:"from"`
	To        []string `json:"to"`
	Cc        []string `json:"cc"`
	Subject   string   `json:"subject"`
	MessageID string   `json:"message_id"`
	Text      string   `json:"text"`
	HTML      string   `json:"html"`
	// Headers is the email's headers. A header with several values is a list.
	Headers map[string]any `json:"headers"`
}

// FetchReceived gets a received email, with its body and headers, by the id from the webhook.
// The webhook itself carries neither.
func (e *EmailClient) FetchReceived(id string) (Received, error) {
	var rec Received
	// html_format=cid keeps inline images out of the answer, they are not shown anyway
	path := "/emails/receiving/" + url.PathEscape(id) + "?html_format=cid"
	if err := e.callGet(path, &rec); err != nil {
		return Received{}, fmt.Errorf("get received email %q: %w", id, err)
	}
	return rec, nil
}

func (e *EmailClient) callGet(path string, out any) error {
	return e.call(http.MethodGet, path, nil, out)
}

// Header is the first value of a header, whatever its case, or empty if the email has none.
func (r Received) Header(name string) string {
	for key, value := range r.Headers {
		if !strings.EqualFold(key, name) {
			continue
		}
		switch v := value.(type) {
		case string:
			return v
		case []any:
			if len(v) > 0 {
				if s, ok := v[0].(string); ok {
					return s
				}
			}
		}
	}
	return ""
}

// Recipients is every address the email was sent to, To and Cc.
func (r Received) Recipients() []string {
	return append(append([]string{}, r.To...), r.Cc...)
}

// Sender splits a From header into the name and the address.
func Sender(from string) (name, address string) {
	parsed, err := mail.ParseAddress(from)
	if err != nil {
		return "", strings.TrimSpace(from)
	}
	return parsed.Name, parsed.Address
}

const replyPrefix = "reply-"

// ReplyAddress is where a visitor's answer to a thread goes. Replies to the site's messages are
// sent with it as Reply-To, so an answer says which conversation it belongs to without relying on
// headers. It is on the site's own domain, the one Resend receives mail for.
func (e *EmailClient) ReplyAddress(threadID string) string {
	_, domain, ok := strings.Cut(e.cfg.From, "@")
	if !ok {
		return ""
	}
	return replyPrefix + threadID + "@" + domain
}

// ThreadFromAddress is the thread an address from ReplyAddress belongs to.
func (e *EmailClient) ThreadFromAddress(address string) (string, bool) {
	_, addr := Sender(address)
	local, domain, ok := strings.Cut(strings.ToLower(addr), "@")
	_, ourDomain, _ := strings.Cut(strings.ToLower(e.cfg.From), "@")
	if !ok || domain != ourDomain || !strings.HasPrefix(local, replyPrefix) {
		return "", false
	}
	id := strings.TrimPrefix(local, replyPrefix)
	if _, err := uuid.Parse(id); err != nil {
		return "", false
	}
	return id, true
}

// ForUs is true if an address is one of the site's: the contact address or a reply address.
func (e *EmailClient) ForUs(address string) bool {
	if _, ok := e.ThreadFromAddress(address); ok {
		return true
	}
	_, addr := Sender(address)
	return strings.EqualFold(addr, e.cfg.From)
}

var (
	tagPattern   = regexp.MustCompile(`(?s)<(script|style)[^>]*>.*?</(script|style)>|<[^>]+>`)
	blankPattern = regexp.MustCompile(`\n{3,}`)
)

// PlainText is the text of a received email. Mail often has only an HTML part, which is reduced
// to its text, nothing from the sender is ever shown as markup.
func (r Received) PlainText() string {
	text := r.Text
	if strings.TrimSpace(text) == "" {
		text = tagPattern.ReplaceAllString(strings.NewReplacer("<br>", "\n", "<br/>", "\n", "</p>", "\n\n", "</div>", "\n").Replace(r.HTML), "")
		text = htmlUnescape(text)
	}
	text = strings.ReplaceAll(text, "\r\n", "\n")
	return strings.TrimSpace(blankPattern.ReplaceAllString(text, "\n\n"))
}
