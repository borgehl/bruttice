package dbx

import (
	"github.com/borgehl/bruttice/packages/go-pkg/models"

	"gorm.io/gorm"
)

func (d *DatabaseClient) GetMigrationFlag(name string, tx *gorm.DB) (ret models.MigrationFlag, err error) {
	err = tx.Where("name = ?", name).First(&ret).Error
	if isErrorNotFoundOK(err) {
		return
	}
	err = nil
	return
}

func (d *DatabaseClient) CreateMigrationFlag(name string, tx *gorm.DB) error {
	return tx.Create(&models.MigrationFlag{Name: name}).Error
}
