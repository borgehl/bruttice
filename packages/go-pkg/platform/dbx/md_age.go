package dbx

import (
	"github.com/borgehl/bruttice/packages/go-pkg/models"

	"gorm.io/gorm"
)

func (d *DatabaseClient) GetSaveAge(age models.Age, tx *gorm.DB) (models.Age, error) {
	ret, err := d.GetAge(models.Age{Name: age.Name}, tx)
	if err != nil {
		return ret, err
	}
	ret.Value = age.Value
	return d.SaveAge(ret, tx)
}

// only gets by the name
func (d *DatabaseClient) GetAge(age models.Age, tx *gorm.DB) (models.Age, error) {
	err := tx.Where(&age).First(&age).Error
	if isErrorNotFoundOK(err) {
		return age, err
	}
	return age, nil
}

func (d *DatabaseClient) SaveAge(age models.Age, tx *gorm.DB) (models.Age, error) {
	err := tx.Save(&age).Error
	if isErrorNotFoundOK(err) {
		return age, err
	}
	return age, nil
}

func (d *DatabaseClient) GetAges(tx *gorm.DB) ([]models.Age, error) {
	ret := make([]models.Age, 0)
	err := tx.Model(&models.Age{}).Order("name desc").Find(&ret).Error
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (d *DatabaseClient) GetAgeNames(tx *gorm.DB) ([]string, error) {
	ret := make([]string, 0)
	err := tx.Model(&models.Age{}).Order("name desc").Pluck("name", &ret).Error
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (d *DatabaseClient) GetAgeValues(tx *gorm.DB) ([]uint8, error) {
	ret := make([]uint8, 0)
	err := tx.Model(&models.Age{}).Order("name desc").Pluck("value", &ret).Error
	if err != nil {
		return nil, err
	}
	return ret, nil
}
