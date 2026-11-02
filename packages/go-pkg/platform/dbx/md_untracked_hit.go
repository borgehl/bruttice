package dbx

import (
	"github.com/borgehl/bruttice/packages/go-pkg/models"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (d *DatabaseClient) IncrementUntrackedHit(path string) error {
	today := time.Now().UTC().Format("2006-01-02")
	hit := models.UntrackedHit{Date: today, Path: path, Count: 1}
	return d.DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "date"}, {Name: "path"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"count": gorm.Expr("untracked_hits.count + 1"),
		}),
	}).Create(&hit).Error
}
