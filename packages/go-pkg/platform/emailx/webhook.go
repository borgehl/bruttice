package emailx

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"html"
	"strconv"
	"strings"
	"time"
)

// webhookTolerance is how far a webhook's timestamp may be from now, so a recorded request
// can't be played back later.
const webhookTolerance = 5 * time.Minute

func htmlUnescape(s string) string { return html.UnescapeString(s) }

// VerifyWebhook checks that a request really came from Resend. Resend signs webhooks the way Svix
// does: an HMAC-SHA256 of "{svix-id}.{svix-timestamp}.{body}" with the secret, which is base64
// after the "whsec_" prefix. The header holds one or more "v1,<signature>" entries. Pass the body
// exactly as received.
func (e *EmailClient) VerifyWebhook(id, timestamp, signatures string, body []byte, now time.Time) error {
	return verifyWebhook(e.cfg.WebhookSecret, id, timestamp, signatures, body, now)
}

func verifyWebhook(secret, id, timestamp, signatures string, body []byte, now time.Time) error {
	if secret == "" {
		return errors.New("no webhook secret")
	}
	if id == "" || timestamp == "" || signatures == "" {
		return errors.New("missing signature headers")
	}
	seconds, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return errors.New("bad timestamp")
	}
	if age := now.Sub(time.Unix(seconds, 0)); age > webhookTolerance || age < -webhookTolerance {
		return errors.New("timestamp too old or too new")
	}
	key, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(secret, "whsec_"))
	if err != nil {
		return errors.New("bad webhook secret")
	}

	mac := hmac.New(sha256.New, key)
	_, _ = fmt.Fprintf(mac, "%s.%s.", id, timestamp)
	_, _ = mac.Write(body)
	expected := "v1," + base64.StdEncoding.EncodeToString(mac.Sum(nil))

	for _, signature := range strings.Fields(signatures) {
		if hmac.Equal([]byte(signature), []byte(expected)) {
			return nil
		}
	}
	return errors.New("signature mismatch")
}
