package models

const (
	// a visitor wrote through the contact form
	InboxSourceContactForm = "contact_form"
	// someone emailed the site's domain, usually answering one of its messages
	InboxSourceEmail = "email"
	// an admin answered from the dashboard
	InboxSourceAdminReply = "admin_reply"
)

// A message in the admin inbox. Messages with the same ThreadID are one conversation: what a
// visitor wrote, the answers, and what they wrote back.
type InboxMessage struct {
	BaseModel
	ThreadID  string `gorm:"index;not null"`
	Source    string `gorm:"not null"`
	FromName  string
	FromEmail string
	Subject   string
	// plain text, never markup: it comes from outside and is shown as it is
	Body string `gorm:"type:text"`
	// the RFC 5322 Message-ID of a received email, the In-Reply-To of its answer
	MessageID string `gorm:"index"`
	InReplyTo string
	// Resend's id of a received email. A webhook can be delivered more than once, this is
	// how the second delivery is recognised. Empty (NULL) for messages that were not received.
	ProviderID *string `gorm:"uniqueIndex"`
	// unix time the message was read or its thread archived, 0 if it was not
	ReadAt     int64
	ArchivedAt int64
}

// whether this message is one an admin still has to read
func (m InboxMessage) NeedsReading() bool {
	return m.Source != InboxSourceAdminReply && m.ReadAt == 0
}
