package dbx

import (
	"github.com/borgehl/bruttice/packages/go-pkg/configs"
	"github.com/borgehl/bruttice/packages/go-pkg/models"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/configx"
)

func (d *DatabaseClient) Migrate(cfg configx.Config) {
	d.DB.AutoMigrate(
		&models.Age{},
		&models.CalendarFilter{},
		&models.CookieModel{},
		&models.Event{},
		&models.EventAdminJudgement{},
		&models.InboxMessage{},
		&models.IngestRun{},
		&models.IngestRunSource{},
		&models.MigrationFlag{},
		&models.Region{},
		&models.Tag{},
		&models.User{},
		&models.Venue{},
		&models.VisitorLog{},
		&models.VisitorLocation{},
		&models.UntrackedHit{},
	)

	tx := d.DB.Begin()
	defer tx.Commit()
	var err error

	_, err = d.GetSaveUser(configs.RootUser, tx)
	if err != nil {
		panic(err)
	}

	for _, ag := range cfg.InitialData.AgeGroups {
		_, err = d.GetSaveAge(models.Age{Name: ag.Name, Value: ag.Value}, tx)
		if err != nil {
			panic(err)
		}
	}

	for _, region := range cfg.InitialData.Regions {
		_, err = d.GetSaveRegion(models.Region{Name: region.Name}, tx)
		if err != nil {
			panic(err)
		}
	}

	for _, v := range cfg.InitialData.Venues {
		venue := models.Venue{
			Name:          v.Name,
			Link:          v.Link,
			DefaultRegion: v.Region,
		}
		venue.InternalUpdate()
		venue, err = d.GetSaveVenueByName(venue, tx)
		if err != nil {
			panic(err)
		}
		d.Logger.Infof("saved initial venue: %+v", venue)
	}

	for _, tagStr := range cfg.InitialData.Tags {
		tag, err := d.SaveTag(models.Tag{Name: tagStr}, tx)
		if err != nil {
			panic(err)
		}
		d.Logger.Infof("saved initial tag: %s", tag.Name)
	}

	err = d.MigrateViews(
		[]ViewModel{
			&models.Visitors{},
			&models.VisitorsDaily{},
			&models.VisitorsMonthly{},
		},
		tx,
	)
	if err != nil {
		panic(err)
	}

	if err = d.MigrateTimestamps(tx); err != nil {
		panic(err)
	}
}
