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

// GetRegion finds a region by what is set on the one given, normally its name. A retired region is
// answered by the region it was merged into, so a scraper that still uses an old name lands in the
// new one. A region retired without a replacement is not found (an empty region).
func (d *DatabaseClient) GetRegion(region models.Region, tx *gorm.DB) (models.Region, error) {
	err := tx.Where(&region).First(&region).Error
	if isErrorNotFoundOK(err) {
		return region, err
	}
	if !region.Retired() {
		return region, nil
	}
	if region.MergedIntoID == nil {
		return models.Region{}, nil
	}
	var target models.Region
	err = tx.Where("id = ?", *region.MergedIntoID).First(&target).Error
	if isErrorNotFoundOK(err) {
		return target, err
	}
	if target.Retired() {
		return models.Region{}, nil
	}
	return target, nil
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
	err := tx.Model(&models.Region{}).Where("retired_at = 0").Order("name").Pluck("name", &ret).Error
	if err != nil {
		return nil, err
	}
	return ret, nil
}
