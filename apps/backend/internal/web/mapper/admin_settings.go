package mapper

import (
	"github.com/borgehl/bruttice/apps/backend/internal/web/dto"
	"github.com/borgehl/bruttice/packages/go-pkg/models"
)

// NotificationsPatch is the notifications a request changes, keyed by their names in the
// users.settings jsonb: the patch merged into it. Absent fields are not in it.
func NotificationsPatch(p dto.AdminNotificationsPatch) map[string]any {
	keys := map[string]any{}
	if p.NewEvents.Present {
		keys[models.SettingNotifyNewEvents] = p.NewEvents.Value
	}
	if p.SystemErrors.Present {
		keys[models.SettingNotifySystemErrors] = p.SystemErrors.Value
	}
	if p.Inbox.Present {
		keys[models.SettingNotifyInbox] = p.Inbox.Value
	}
	return keys
}

// Notifications is the notifications an admin chose.
func Notifications(s models.UserSettings) *dto.AdminNotifications {
	return &dto.AdminNotifications{
		NewEvents:    s.NotifyNewEvents,
		SystemErrors: s.NotifySystemErrors,
		Inbox:        s.NotifyInbox,
	}
}
