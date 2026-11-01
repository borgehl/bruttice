package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type User struct {
	BaseModel

	Email        string
	PasswordHash []byte
	// how the admin is shown to other admins, such as on their replies in the inbox. Empty uses
	// the email.
	Name string
	// Settings is written only by a jsonb merge (dbx MergeUserSettings), never with the rest of the
	// row, so two requests changing different settings keep both. An absent key is the zero value.
	Settings UserSettings `gorm:"type:jsonb;not null;default:'{}'"`
}

// the keys of UserSettings in the jsonb column, which are also the notification kinds
const (
	SettingNotifyNewEvents    = "notify_new_events"
	SettingNotifySystemErrors = "notify_system_errors"
	SettingNotifyInbox        = "notify_inbox"
)

// UserSettings is what an admin chose for themself. The notify settings say which emails they get.
type UserSettings struct {
	// NotifyNewEvents is the hourly note of events waiting for approval.
	NotifyNewEvents bool `json:"notify_new_events"`
	// NotifySystemErrors is the report of the errors of an ingest run, and database errors the
	// site hits on its own.
	NotifySystemErrors bool `json:"notify_system_errors"`
	// NotifyInbox is the note that a message is waiting in the admin inbox.
	NotifyInbox bool `json:"notify_inbox"`
}

func (s *UserSettings) Scan(value any) error {
	switch v := value.(type) {
	case nil:
		*s = UserSettings{}
		return nil
	case []byte:
		return json.Unmarshal(v, s)
	case string:
		return json.Unmarshal([]byte(v), s)
	}
	return fmt.Errorf("cannot scan type %T into UserSettings", value)
}

func (s UserSettings) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// DisplayName is the name to show for an admin.
func (u User) DisplayName() string {
	if u.Name != "" {
		return u.Name
	}
	return u.Email
}
