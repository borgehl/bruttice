package emailx

import (
	"fmt"
	"strings"
)

const (
	secureProtocol   = "https://"
	unsecureProtocol = "http://"
)

// GenericMessage is a note to the admins. It has no recipient: the caller sets To to the admins who
// chose to get it.
func (e *EmailClient) GenericMessage(subject string, text string) Message {
	return Message{Subject: subject, Text: text}
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

// InboxNoticeMessage tells the admins that a message is waiting in the admin inbox. It has no
// recipient: the caller sets To to the admins who chose to get it. It points to the inbox and has
// no Reply-To: the answer is written there, so it is kept with the thread.
func (e *EmailClient) InboxNoticeMessage(name, email, subject, text string) Message {
	name = oneLine(name)
	if name == "" {
		name = email
	}
	return Message{
		Subject: fmt.Sprintf("New message from %s", name),
		Text: fmt.Sprintf("%s <%s>\nSubject: %s\n\n%s\n\nRead it and reply in the admin inbox: %s/admin\n",
			name, email, oneLine(subject), preview(text), e.baseHost()),
	}
}

// ReplyMessage is the admin's answer to a thread, to the visitor. The visitor's reply to it goes
// to the thread's reply address, so it lands in the same conversation. inReplyTo is the Message-ID
// of the message answered, if it had one. fromName is the admin's name, the site's if empty.
func (e *EmailClient) ReplyMessage(to, fromName, subject, text, threadID, inReplyTo string) Message {
	msg := Message{To: []string{to}, FromName: fromName, Subject: subject, Text: text, ReplyTo: e.ReplyAddress(threadID)}
	if inReplyTo != "" {
		msg.Headers = map[string]string{"In-Reply-To": inReplyTo, "References": inReplyTo}
	}
	return msg
}

// baseHost is the site's public URL with a scheme, http if the configured one has none.
func (e *EmailClient) baseHost() string {
	host := strings.TrimPrefix(strings.TrimPrefix(e.cfg.Host, secureProtocol), unsecureProtocol)
	if strings.HasPrefix(e.cfg.Host, secureProtocol) {
		return secureProtocol + host
	}
	return unsecureProtocol + host
}

// preview is the start of a message, enough to know what it is about.
func preview(text string) string {
	const max = 500
	text = strings.TrimSpace(text)
	if r := []rune(text); len(r) > max {
		return string(r[:max]) + "..."
	}
	return text
}

// oneLine keeps a visitor's text from breaking out of the line it is shown on.
func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
