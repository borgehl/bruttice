package dbx

import (
	"github.com/borgehl/bruttice/packages/go-pkg/models"

	"gorm.io/gorm"
)

func (d *DatabaseClient) GetFilterByKey(key string, tx *gorm.DB) (ret models.CalendarFilter, err error) {
	if key == "" {
		return
	}
	ret.HashKey = key
	err = tx.Where(&ret).First(&ret).Error
	if isErrorNotFoundOK(err) {
		return
	}
	err = nil
	return
}

func (d *DatabaseClient) GetFilterByID(id int64, tx *gorm.DB) (ret models.CalendarFilter, err error) {
	ret.ID = id
	err = tx.Where(&ret).First(&ret).Error
	if isErrorNotFoundOK(err) {
		return
	}
	err = nil
	return
}

func (d *DatabaseClient) SaveFilter(filter models.CalendarFilter, tx *gorm.DB) (models.CalendarFilter, error) {
	ret, err := d.GetFilterByKey(filter.HashKey, tx)
	if err != nil {
		return models.CalendarFilter{}, err
	}
	filter.ID = ret.ID
	err = tx.Save(&filter).Error
	return ret, err
}
