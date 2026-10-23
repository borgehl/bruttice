package emailx

import (
	"fmt"
	"strings"
	"time"

	"github.com/borgehl/bruttice/packages/go-pkg/configs"
)

const (
	errDateFmt       = "Mon Jan _2"
	secureProtocol   = "https://"
	unsecureProtocol = "http://"
)

// ErrorsMessage tells the root admin about the errors of an ingest run.
func (e *EmailClient) ErrorsMessage(errs []error) Message {
	text := "Please see recorded errors below\n"
	for i, err := range errs {
		text += fmt.Sprintf("%d: %s\n", i, err)
	}
	return Message{
		To:      []string{configs.RootUser.Email},
		Subject: fmt.Sprintf("%s Scraping Errors: %d", time.Now().Format(errDateFmt), len(errs)),
		Text:    text,
	}
}

// GenericMessage is a note to the root admin.
func (e *EmailClient) GenericMessage(subject string, text string) Message {
	return Message{To: []string{configs.RootUser.Email}, Subject: subject, Text: text}
}

// VerifyEventMessage asks the author of a submitted event to confirm their address.
func (e *EmailClient) VerifyEventMessage(title string, to string, verificationCode string) Message {
	return Message{
		To:      []string{to},
		Subject: "Please verify your event",
		Text: fmt.Sprintf("Hello Author of %s!\n", title) +
			"Thanks for posting your event.\n" +
			"Please click on the following link to verify your email address:\n" +
			fmt.Sprintf("%s/verify-event?verifyKey=%s\n", e.baseHost(), verificationCode),
	}
}

// ContactMessage is a visitor's message from the contact form, to the site's contact address.
// A reply goes to the visitor.
func (e *EmailClient) ContactMessage(name, email, text string) Message {
	name = oneLine(name)
	return Message{
		To:      []string{e.cfg.From},
		Subject: fmt.Sprintf("Message from %s", name),
		ReplyTo: email,
		Text:    fmt.Sprintf("From: %s <%s>\nSent from the contact form on %s\n\n%s\n", name, email, e.baseHost(), text),
	}
}

// baseHost is the site's public URL with a scheme, http if the configured one has none.
func (e *EmailClient) baseHost() string {
	host := strings.TrimPrefix(strings.TrimPrefix(e.cfg.Host, secureProtocol), unsecureProtocol)
	if strings.HasPrefix(e.cfg.Host, secureProtocol) {
		return secureProtocol + host
	}
	return unsecureProtocol + host
}

// oneLine keeps a visitor's text from breaking out of the line it is shown on.
func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
