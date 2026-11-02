package dbx

import "gorm.io/gorm"

type ViewModel interface {
	CreateView() string
}

func (d *DatabaseClient) MigrateViews(views []ViewModel, tx *gorm.DB) error {
	for _, view := range views {
		d.Logger.Infof("Migrating view %s", view.CreateView())
		if err := tx.Exec(view.CreateView()).Error; err != nil {
			return err
		}
	}
	return nil
}
