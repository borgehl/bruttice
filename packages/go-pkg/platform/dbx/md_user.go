package dbx

import (
	"github.com/borgehl/bruttice/packages/go-pkg/models"

	"gorm.io/gorm"
)

func (d *DatabaseClient) GetSaveUser(user models.User, tx *gorm.DB) (models.User, error) {
	ret, err := d.GetUser(user, tx)
	if err != nil {
		return ret, err
	}
	if ret.ID != 0 {
		return ret, nil
	}
	return d.SaveUser(ret, tx)
}

func (d *DatabaseClient) GetUser(user models.User, tx *gorm.DB) (models.User, error) {
	err := tx.Where(&user).First(&user).Error
	if isErrorNotFoundOK(err) {
		return user, err
	}
	return user, nil
}

func (d *DatabaseClient) SaveUser(user models.User, tx *gorm.DB) (models.User, error) {
	// settings are only written by MergeUserSettings: saving the loaded row would put back what
	// it held when it was read
	err := tx.Omit("settings").Save(&user).Error
	if isErrorNotFoundOK(err) {
		return user, err
	}
	return user, nil
}

// sets the name an admin is shown with
func (d *DatabaseClient) SetUserName(userID int64, name string) error {
	return d.DB.Model(&models.User{}).Where("id = ?", userID).Update("name", name).Error
}

// MergeUserSettings merges patch (settings keyed by their json names) into the user's settings as
// one statement, so a setting absent from patch is never read or rewritten and a concurrent change
// of another one survives. It answers the settings as they are after the merge.
func (d *DatabaseClient) MergeUserSettings(userID int64, patch map[string]any) (models.UserSettings, error) {
	var settings models.UserSettings
	if len(patch) == 0 {
		err := d.DB.Model(&models.User{}).Select("settings").Where("id = ?", userID).Row().Scan(&settings)
		return settings, err
	}
	err := MergeJSONB(d.DB, "users", "settings", userID, patch, &settings)
	return settings, err
}

// NotificationRecipients are the emails of the admins who chose the notification (one of the
// models.SettingNotify keys). It is empty if nobody did: then the mail is not sent.
func (d *DatabaseClient) NotificationRecipients(kind string) ([]string, error) {
	emails := make([]string, 0)
	err := d.DB.Model(&models.User{}).
		Where("settings ->> ? = 'true' AND email <> ''", kind).
		Order("id").
		Pluck("email", &emails).Error
	return emails, err
}
