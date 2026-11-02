package dbx

import (
	"time"

	"github.com/borgehl/bruttice/packages/go-pkg/models"

	"gorm.io/gorm"
)

// how many conversations the inbox lists
const inboxThreadLimit = 100

func (d *DatabaseClient) CreateInboxMessage(message *models.InboxMessage) error {
	return d.DB.Create(message).Error
}

// whether a received email was already saved
func (d *DatabaseClient) InboxProviderIDSeen(providerID string) (bool, error) {
	var count int64
	err := d.DB.Model(&models.InboxMessage{}).Where("provider_id = ?", providerID).Count(&count).Error
	return count > 0, err
}

// MarkInboxDelivered records when an email an admin sent reached the recipient's server. A
// repeated or late event changes nothing, the first time stays. An id that is not a message of the
// inbox (the site sends other emails too) matches nothing and is not an error.
func (d *DatabaseClient) MarkInboxDelivered(sentID string, at int64) error {
	return d.DB.Model(&models.InboxMessage{}).
		Where("sent_id = ? AND delivered_at = 0", sentID).
		Update("delivered_at", at).Error
}

// the message with this Message-ID, or nil if there is none
func (d *DatabaseClient) FindInboxMessageByMessageID(messageID string) (*models.InboxMessage, error) {
	var message models.InboxMessage
	err := d.DB.Where("message_id = ?", messageID).Order("id").Limit(1).Find(&message).Error
	if err != nil || message.ID == 0 {
		return nil, err
	}
	return &message, nil
}

// the messages of one conversation, oldest first
func (d *DatabaseClient) GetInboxThread(threadID string) ([]models.InboxMessage, error) {
	messages := make([]models.InboxMessage, 0)
	err := d.DB.Where("thread_id = ?", threadID).Order("id").Find(&messages).Error
	return messages, err
}

// the messages of the conversations with the latest activity, oldest first. Archived
// conversations are left out unless includeArchived.
func (d *DatabaseClient) ListInboxMessages(includeArchived bool) ([]models.InboxMessage, error) {
	threads := d.DB.Model(&models.InboxMessage{}).Select("thread_id").Group("thread_id").
		Order("MAX(id) DESC").Limit(inboxThreadLimit)
	if !includeArchived {
		threads = threads.Where("archived_at = 0")
	}

	messages := make([]models.InboxMessage, 0)
	err := d.DB.Where("thread_id IN (?)", threads).Order("id").Find(&messages).Error
	return messages, err
}

// how many conversations have a message an admin has not read
func (d *DatabaseClient) CountUnreadInboxThreads() (int64, error) {
	var count int64
	err := d.DB.Model(&models.InboxMessage{}).
		Where("read_at = 0 AND archived_at = 0 AND source <> ?", models.InboxSourceAdminReply).
		Distinct("thread_id").Count(&count).Error
	return count, err
}

func (d *DatabaseClient) MarkInboxThreadRead(threadID string) error {
	return d.DB.Model(&models.InboxMessage{}).
		Where("thread_id = ? AND read_at = 0", threadID).
		Update("read_at", time.Now().Unix()).Error
}

// archives a conversation, or brings it back. Archiving reads it too.
func (d *DatabaseClient) SetInboxThreadArchived(threadID string, archived bool) error {
	if !archived {
		return d.DB.Model(&models.InboxMessage{}).Where("thread_id = ?", threadID).Update("archived_at", 0).Error
	}
	now := time.Now().Unix()
	return d.DB.Model(&models.InboxMessage{}).Where("thread_id = ?", threadID).
		Updates(map[string]any{"archived_at": now, "read_at": gorm.Expr("CASE WHEN read_at = 0 THEN ? ELSE read_at END", now)}).Error
}

// DeleteInboxThread removes every message of a conversation for good, in any state. The rows are
// not soft-deleted: the point is to be rid of them, and a soft-deleted row would still hold its
// provider_id in the unique index. A reply to a deleted conversation starts a new one, because
// neither its reply address nor its In-Reply-To finds a message any more.
func (d *DatabaseClient) DeleteInboxThread(threadID string) error {
	return d.DB.Unscoped().Where("thread_id = ?", threadID).Delete(&models.InboxMessage{}).Error
}

// DeleteArchivedInboxThreads removes every archived conversation for good and says how many
// messages went. A conversation is archived when all of its messages are, as in the inbox list, so
// one that got a new message since is kept.
func (d *DatabaseClient) DeleteArchivedInboxThreads() (int64, error) {
	result := d.DB.Unscoped().
		Where("thread_id NOT IN (?)", d.DB.Model(&models.InboxMessage{}).Select("thread_id").Where("archived_at = 0")).
		Where("archived_at > 0").
		Delete(&models.InboxMessage{})
	return result.RowsAffected, result.Error
}
