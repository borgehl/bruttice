// Package emailx sends the site's email through Resend (https://resend.com). Mail goes out
// from the site's contact address, which has to be on a domain verified in Resend.
package emailx

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"time"

	"github.com/borgehl/bruttice/packages/go-pkg/platform/logx"
)

const (
	resendURL   = "https://api.resend.com"
	sendTimeout = 10 * time.Second
	// the largest answer read, a received email with its inline images can be big
	maxResponse = 16 << 20
)

type Config struct {
	// APIKey is the Resend API key. Without it nothing is sent.
	APIKey string
	// From is the site's contact address, the sender of every message. Without it nothing is sent.
	From string
	// SiteName is shown as the sender's name.
	SiteName string
	// Host is the public base URL of the site, for links in messages.
	Host string
	// BaseURL is the API to send through, Resend's unless a test says otherwise.
	BaseURL string
	// WebhookSecret is the signing secret (whsec_...) of the Resend webhook that delivers
	// received mail. Without it incoming mail is off.
	WebhookSecret string
}

// Message is a plain text email. The sender is always the site's contact address.
type Message struct {
	To      []string
	Subject string
	Text    string
	// ReplyTo is where a reply goes, if not to the sender.
	ReplyTo string
	// Headers are extra email headers, such as In-Reply-To.
	Headers map[string]string
}

type EmailClient struct {
	cfg     Config
	baseURL string
	http    *http.Client
	logger  *logx.Logger
}

func NewEmailClient(logger *logx.Logger, cfg Config) *EmailClient {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = resendURL
	}
	return &EmailClient{
		cfg:     cfg,
		baseURL: baseURL,
		http:    &http.Client{Timeout: sendTimeout},
		logger:  logger,
	}
}

// Enabled is true if messages are really sent. Otherwise Send only logs.
func (e *EmailClient) Enabled() bool {
	return e.cfg.APIKey != "" && e.cfg.From != ""
}

// InboundEnabled is true if mail sent to the site's domain can be received: sending works and the
// webhook has a signing secret.
func (e *EmailClient) InboundEnabled() bool {
	return e.Enabled() && e.cfg.WebhookSecret != ""
}

// Contact is the site's contact address, empty if there is none.
func (e *EmailClient) Contact() string {
	return e.cfg.From
}

type sendRequest struct {
	From    string            `json:"from"`
	To      []string          `json:"to"`
	Subject string            `json:"subject"`
	Text    string            `json:"text"`
	ReplyTo []string          `json:"reply_to,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// Send delivers a message and returns Resend's id for it, and logs why if it could not. A site
// without a Resend key or a contact address sends nothing and it is not an error, a deployment
// may have no email. The id is empty then.
func (e *EmailClient) Send(msg Message) (string, error) {
	if !e.Enabled() {
		e.logger.Infof("email is not configured (RESEND_API_KEY and site.contact-email), not sending %q to %v", msg.Subject, msg.To)
		return "", nil
	}

	body := sendRequest{
		From:    (&mail.Address{Name: e.cfg.SiteName, Address: e.cfg.From}).String(),
		To:      msg.To,
		Subject: msg.Subject,
		Text:    msg.Text,
		Headers: msg.Headers,
	}
	if msg.ReplyTo != "" {
		body.ReplyTo = []string{msg.ReplyTo}
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return "", e.fail(msg, err)
	}

	var sent struct {
		ID string `json:"id"`
	}
	if err := e.call(http.MethodPost, "/emails", payload, &sent); err != nil {
		return "", e.fail(msg, err)
	}
	e.logger.Infof("sent %q to %v", msg.Subject, msg.To)
	return sent.ID, nil
}

// call makes one request to the API and decodes the answer into out.
func (e *EmailClient) call(method, path string, payload []byte, out any) error {
	ctx, cancel := context.WithTimeout(context.Background(), sendTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, method, e.baseURL+path, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+e.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("resend answered %d: %s", resp.StatusCode, resendMessage(resp.Body))
	}
	// an empty answer is fine, the caller has nothing to read from it
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxResponse)).Decode(out); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}

func (e *EmailClient) fail(msg Message, err error) error {
	e.logger.Errorf("failed to send %q to %v: %v", msg.Subject, msg.To, err)
	return err
}

// resendMessage is the reason in an error answer, which looks like {"message": "...", "name": "..."}.
func resendMessage(body io.Reader) string {
	data, _ := io.ReadAll(io.LimitReader(body, 4<<10))
	var parsed struct {
		Message string `json:"message"`
	}
	if json.Unmarshal(data, &parsed) == nil && parsed.Message != "" {
		return parsed.Message
	}
	return string(data)
}
