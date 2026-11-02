package dbx

import (
	"github.com/borgehl/bruttice/packages/go-pkg/models"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/configx"

	"gorm.io/gorm"
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

	if err := d.seedLists(cfg, tx); err != nil {
		panic(err)
	}

	err := d.MigrateViews(
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

// seedListsKey marks that the lists in the config were saved. They are only saved into a
// database that has not had them yet: after that the admin page owns them, and saving them again on
// every start would bring back whatever an admin retired or renamed.
const seedListsKey = "seed_initial_lists"

func (d *DatabaseClient) seedLists(cfg configx.Config, tx *gorm.DB) error {
	flag, err := d.GetMigrationFlag(seedListsKey, tx)
	if err != nil {
		return err
	}
	data := cfg.InitialData
	if flag.ID != 0 {
		if len(data.Regions)+len(data.Tags)+len(data.AgeGroups)+len(data.Venues) > 0 {
			d.Logger.Infof("regions, tags, age groups and venues in the config are only used on the first start, manage them in the admin settings")
		}
		return nil
	}

	for _, ag := range data.AgeGroups {
		_, err = d.GetSaveAge(models.Age{Name: ag.Name, Value: ag.Value}, tx)
		if err != nil {
			return err
		}
	}

	for _, region := range data.Regions {
		_, err = d.GetSaveRegion(models.Region{Name: region.Name}, tx)
		if err != nil {
			return err
		}
	}

	for _, v := range data.Venues {
		venue := models.Venue{
			Name:          v.Name,
			Link:          v.Link,
			DefaultRegion: v.Region,
		}
		venue.InternalUpdate()
		venue, err = d.GetSaveVenueByName(venue, tx)
		if err != nil {
			return err
		}
		d.Logger.Infof("saved initial venue: %+v", venue)
	}

	for _, tagStr := range data.Tags {
		tag, err := d.SaveTag(models.Tag{Name: tagStr}, tx)
		if err != nil {
			return err
		}
		d.Logger.Infof("saved initial tag: %s", tag.Name)
	}

	return d.CreateMigrationFlag(seedListsKey, tx)
}
