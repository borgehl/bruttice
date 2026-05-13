package dbx

import (
	"github.com/borgehl/bruttice/packages/go-pkg/configs"
	"github.com/borgehl/bruttice/packages/go-pkg/models"
	"slices"

	"gorm.io/gorm"
)

func (d *DatabaseClient) SaveTag(tag models.Tag, tx *gorm.DB) (models.Tag, error) {
	if !slices.Contains(configs.Tags, tag.Name) {
		return models.Tag{}, nil
	}
	existingTag, err := d.GetTagByName(tag.Name, tx)
	if err != nil {
		return tag, err
	}
	if existingTag.ID != 0 {
		return existingTag, nil
	}
	err = tx.Save(&tag).Error
	return tag, err
}

func (d *DatabaseClient) GetTagByName(name string, tx *gorm.DB) (ret models.Tag, err error) {
	err = tx.Where("name = ?", name).First(&ret).Error
	if isErrorNotFoundOK(err) {
		return
	}
	err = nil
	return
}

// ResolveTags returns only those tags already saved in the database; unknown tags are silently skipped.
func (d *DatabaseClient) ResolveTags(tags []models.Tag, tx *gorm.DB) ([]models.Tag, error) {
	resolved := make([]models.Tag, 0, len(tags))
	for _, tag := range tags {
		found, err := d.GetTagByName(tag.Name, tx)
		if err != nil {
			return nil, err
		}
		if found.ID == 0 {
			continue
		}
		resolved = append(resolved, found)
	}
	return resolved, nil
}

func (d *DatabaseClient) GetTagIds(tags []string, tx *gorm.DB) ([]int64, error) {
	ret := make([]int64, 0)
	err := tx.Model(&models.Tag{}).Where("name in ?", tags).Order("name").Pluck("id", &ret).Error
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// ReplaceEventTags resolves tags by name and replaces the event's join-table associations.
func (d *DatabaseClient) ReplaceEventTags(event models.Event, tags []models.Tag, tx *gorm.DB) error {
	resolved, err := d.ResolveTags(tags, tx)
	if err != nil {
		return err
	}
	return tx.Model(&event).Association("Tags").Replace(resolved)
}

func (d *DatabaseClient) GetTagNames(tx *gorm.DB) ([]string, error) {
	ret := make([]string, 0)
	err := tx.Model(&models.Tag{}).Order("name").Pluck("name", &ret).Error
	if err != nil {
		return nil, err
	}
	return ret, nil
}
