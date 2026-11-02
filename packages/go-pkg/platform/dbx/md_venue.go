package dbx

import (
	"github.com/borgehl/bruttice/packages/go-pkg/models"

	"gorm.io/gorm"
)

func (d *DatabaseClient) GetSaveVenue(venue models.Venue, tx *gorm.DB) (models.Venue, error) {
	ret, err := d.GetVenue(venue, tx)
	if err != nil {
		return ret, err
	}
	if ret.ID != 0 {
		venue.ID = ret.ID
		venue.CreatedAt = ret.CreatedAt
		venue.DeletedAt = ret.DeletedAt
	}
	venue.InternalUpdate()
	return d.SaveVenue(venue, tx)
}

// Finds venue by name, and overwrites existing data with supplied data
func (d *DatabaseClient) GetSaveVenueByName(venue models.Venue, tx *gorm.DB) (ret models.Venue, err error) {
	getVenue := models.Venue{Name: venue.Name}
	getVenue, err = d.GetVenue(getVenue, tx)
	if err != nil {
		return ret, err
	}
	if getVenue.ID != 0 {
		venue.ID = getVenue.ID
	}
	return d.SaveVenue(venue, tx)
}

func (d *DatabaseClient) GetVenue(venue models.Venue, tx *gorm.DB) (models.Venue, error) {
	err := tx.Where(&venue).First(&venue).Error
	if isErrorNotFoundOK(err) {
		return models.Venue{}, err
	}
	return venue, nil
}

func (d *DatabaseClient) GetVenueByName(name string, tx *gorm.DB) (models.Venue, error) {
	venue := models.Venue{}
	err := tx.Where("name = ?", name).First(&venue).Error
	if isErrorNotFoundOK(err) {
		return models.Venue{}, err
	}
	return venue, nil
}

func (d *DatabaseClient) SaveVenue(venue models.Venue, tx *gorm.DB) (models.Venue, error) {
	err := tx.Save(&venue).Error
	if err != nil {
		return venue, err
	}
	return venue, nil
}

func (d *DatabaseClient) GetVenueNames(tx *gorm.DB) ([]string, error) {
	ret := make([]string, 0)
	err := tx.Model(&models.Venue{}).Order("slug").Pluck("name", &ret).Error
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (d *DatabaseClient) GetVenueIds(names []string, tx *gorm.DB) ([]int64, error) {
	ret := make([]int64, 0)
	err := tx.Model(&models.Venue{}).Where("name in ?", names).Pluck("id", &ret).Error
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (d *DatabaseClient) GetVenues(tx *gorm.DB) ([]models.Venue, error) {
	ret := make([]models.Venue, 0)
	err := tx.Order("slug").Find(&ret).Error
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (d *DatabaseClient) GetSubmissionVenues(tx *gorm.DB) ([]models.Venue, error) {
	ret := make([]models.Venue, 0)
	err := tx.Where("is_scraped = ?", false).Order("slug").Find(&ret).Error
	if err != nil {
		return nil, err
	}
	return ret, nil
}
