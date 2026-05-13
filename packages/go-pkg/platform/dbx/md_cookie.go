package dbx

import (
	"github.com/borgehl/bruttice/packages/go-pkg/models"

	"gorm.io/gorm"
)

func (d *DatabaseClient) SaveCookie(cookie models.SessionCookie, tx *gorm.DB) (models.CookieModel, error) {
	ret, err := d.GetCookieByValue(cookie.Value, tx)
	if err != nil {
		return ret, err
	}
	ret.UpdateCookie(cookie)
	err = tx.Save(&ret).Error
	return ret, err
}

func (d *DatabaseClient) GetCookieByValue(value string, tx *gorm.DB) (models.CookieModel, error) {
	ret := models.CookieModel{Value: value}
	err := tx.Where(&ret).Find(&ret).Error
	if isErrorNotFoundOK(err) {
		return ret, err
	}
	return ret, nil
}
