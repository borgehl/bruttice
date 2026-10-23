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
