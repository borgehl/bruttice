package dbx

import (
	"github.com/borgehl/bruttice/packages/go-pkg/models"

	"gorm.io/gorm"
)

func (d *DatabaseClient) SaveVisitorLog(log models.VisitorLog, loc models.VisitorLocation, tx *gorm.DB) error {
	savedLocation := models.VisitorLocation{VisitorID: log.VisitorID}
	err := tx.Where(&savedLocation).First(&savedLocation).Error
	if isErrorNotFoundOK(err) {
		tx.Rollback()
		return err
	}

	if savedLocation.ID == 0 {
		if err = tx.Save(&loc).Error; err != nil {
			tx.Rollback()
			return err
		}
		savedLocation = loc
	}

	log.VisitorLocationID = savedLocation.ID
	return tx.Save(&log).Error
}
