package dbx

import (
	"github.com/borgehl/bruttice/packages/go-pkg/models"
	"time"

	"gorm.io/gorm"
)

func (d *DatabaseClient) CreateIngestRun() (models.IngestRun, error) {
	run := models.IngestRun{StartedAt: time.Now().Unix()}
	err := d.DB.Create(&run).Error
	return run, err
}

func (d *DatabaseClient) FinishIngestRun(runID int64, dedupeErr error) error {
	updates := map[string]interface{}{"finished_at": time.Now().Unix()}
	if dedupeErr != nil {
		updates["dedupe_error"] = dedupeErr.Error()
	}
	return d.DB.Model(&models.IngestRun{}).Where("id = ?", runID).Updates(updates).Error
}

func (d *DatabaseClient) SaveIngestRunSource(source models.IngestRunSource) error {
	return d.DB.Create(&source).Error
}

// the run with its sources, or an empty run if there is no such id
func (d *DatabaseClient) GetIngestRun(runID int64) (models.IngestRun, error) {
	run := models.IngestRun{}
	err := d.DB.Preload("Sources", func(db *gorm.DB) *gorm.DB { return db.Order("source") }).
		Where("id = ?", runID).
		First(&run).Error
	if isErrorNotFoundOK(err) {
		return models.IngestRun{}, err
	}
	return run, nil
}

// most recent runs first, with their sources
//
// beforeID > 0 returns only runs older than that id
func (d *DatabaseClient) GetRecentIngestRuns(limit int, beforeID int64) ([]models.IngestRun, error) {
	ret := make([]models.IngestRun, 0)
	tx := d.DB.
		Preload("Sources", func(db *gorm.DB) *gorm.DB { return db.Order("source") }).
		Order("id desc").
		Limit(limit)
	if beforeID > 0 {
		tx = tx.Where("id < ?", beforeID)
	}
	err := tx.Find(&ret).Error
	return ret, err
}

// records a deduplication failure on a run that is still in progress
func (d *DatabaseClient) SetIngestRunDedupeError(runID int64, dedupeErr error) error {
	return d.DB.Model(&models.IngestRun{}).Where("id = ?", runID).Update("dedupe_error", dedupeErr.Error()).Error
}
