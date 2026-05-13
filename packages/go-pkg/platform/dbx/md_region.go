package dbx

import (
	"github.com/borgehl/bruttice/packages/go-pkg/models"

	"gorm.io/gorm"
)

func (d *DatabaseClient) GetSaveRegion(region models.Region, tx *gorm.DB) (models.Region, error) {
	ret, err := d.GetRegion(region, tx)
	if err != nil {
		return ret, err
	}
	if ret.ID != 0 {
		return ret, nil
	}
	return d.SaveRegion(ret, tx)
}

func (d *DatabaseClient) GetRegion(region models.Region, tx *gorm.DB) (models.Region, error) {
	err := tx.Where(&region).First(&region).Error
	if isErrorNotFoundOK(err) {
		return region, err
	}
	return region, nil
}

func (d *DatabaseClient) SaveRegion(region models.Region, tx *gorm.DB) (models.Region, error) {
	err := tx.Save(&region).Error
	if isErrorNotFoundOK(err) {
		return region, err
	}
	return region, nil
}

func (d *DatabaseClient) GetRegionNames(tx *gorm.DB) ([]string, error) {
	ret := make([]string, 0)
	err := tx.Model(&models.Region{}).Order("name").Pluck("name", &ret).Error
	if err != nil {
		return nil, err
	}
	return ret, nil
}
