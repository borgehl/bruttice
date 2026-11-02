package dbx

import (
	"github.com/borgehl/bruttice/packages/go-pkg/models"

	"gorm.io/gorm"
)

func (d *DatabaseClient) GetSaveJudgement(judgement models.EventAdminJudgement, tx *gorm.DB) (models.EventAdminJudgement, error) {
	ret, err := d.GetJudgement(judgement, tx)
	if err != nil {
		return ret, err
	}
	if ret.ID != 0 {
		return ret, nil
	}
	return d.SaveJudgement(ret, tx)
}

func (d *DatabaseClient) GetJudgement(judgement models.EventAdminJudgement, tx *gorm.DB) (models.EventAdminJudgement, error) {
	err := tx.Where(&judgement).First(&judgement).Error
	if isErrorNotFoundOK(err) {
		return judgement, err
	}
	return judgement, nil
}

func (d *DatabaseClient) SaveJudgement(judgement models.EventAdminJudgement, tx *gorm.DB) (models.EventAdminJudgement, error) {
	err := tx.Save(&judgement).Error
	if isErrorNotFoundOK(err) {
		return judgement, err
	}
	return judgement, nil
}

func (d *DatabaseClient) DeleteJudgementsForEvent(eventID int64, tx *gorm.DB) error {
	return tx.Where("event_id = ?", eventID).Delete(&models.EventAdminJudgement{}).Error
}

// ReplaceEventJudgements soft-deletes all existing judgements for the event and
// creates the new set. GORM's has-many Replace nullifies the FK rather than
// soft-deleting, so we manage this explicitly.
func (d *DatabaseClient) ReplaceEventJudgements(event models.Event, judgements []models.EventAdminJudgement, tx *gorm.DB) error {
	if err := d.DeleteJudgementsForEvent(event.ID, tx); err != nil {
		return err
	}
	if len(judgements) == 0 {
		return nil
	}
	for i := range judgements {
		judgements[i].EventID = event.ID
	}
	return tx.Create(&judgements).Error
}
