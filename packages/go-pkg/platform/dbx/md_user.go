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
	err := tx.Save(&user).Error
	if isErrorNotFoundOK(err) {
		return user, err
	}
	return user, nil
}
