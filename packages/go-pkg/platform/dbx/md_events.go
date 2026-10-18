package dbx

import (
	"fmt"
	"github.com/borgehl/bruttice/packages/go-pkg/configs"
	"github.com/borgehl/bruttice/packages/go-pkg/models"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var eventPreloads = []string{
	"Age",
	"Region",
	"Venue",
	"EventAdminJudgements",
	"Tags",
}

// checks for guid and overwrites if found
// if age/region/venue are not found,ids will be zero and
// error will be returned
func (d *DatabaseClient) SaveEvent(event models.Event, isScraper bool, tx *gorm.DB) (models.Event, error) {
	existingEvent, err := d.GetEventByGuid(event.Guid, tx.Unscoped())
	if err != nil {
		return event, err
	}
	if existingEvent.ID != 0 {
		event.ID = existingEvent.ID
		event.CreatedAt = existingEvent.CreatedAt
		event.DeletedAt = existingEvent.DeletedAt
	}

	event.Age, err = d.GetAge(event.Age, tx)
	if err != nil {
		return event, err
	}
	event.Region, err = d.GetRegion(event.Region, tx)
	if err != nil {
		return event, err
	}
	event.Venue, err = d.GetVenue(event.Venue, tx)
	if err != nil {
		return event, err
	}

	if event.Venue.IsScraped != isScraper {
		return event, fmt.Errorf("venue auto flag is %t, but calling function auto flag is %t", event.Venue.IsScraped, isScraper)
	}

	// overwrite region with venue region if the default is known (non-unlisted venue)
	if event.Venue.Name != configs.UnlistedName && event.Venue.DefaultRegion != event.Region.Name {
		event.Region, err = d.GetRegion(models.Region{Name: event.Venue.DefaultRegion}, tx)
		if err != nil {
			return event, err
		}
	}

	event.Tags, err = d.ResolveTags(event.Tags, tx)
	if err != nil {
		return event, err
	}

	if event.Age.ID == 0 || event.Region.ID == 0 || event.Venue.ID == 0 {
		return event, fmt.Errorf("invalid age, region or venue supplied for event: %s, %s, %s", event.Age.Name, event.Region.Name, event.Venue.Name)
	}

	if err = tx.Save(&event).Error; err != nil {
		d.Logger.Errorf("failed to save %s: %s", event.Title, err)
		return event, err
	}
	return event, nil
}

// UpdateEvent saves an already-loaded event with modified fields.
// Replaces the tag association via ReplaceEventTags.
// Clears judgements as well
func (d *DatabaseClient) UpdateEvent(event models.Event, tx *gorm.DB) (models.Event, error) {
	if err := d.ReplaceEventJudgements(event, nil, tx); err != nil {
		d.Logger.Errorf("failed to clear judgements for event: %v", err)
		return event, err
	}
	if err := tx.Omit(clause.Associations).Save(&event).Error; err != nil {
		d.Logger.Errorf("failed to update %s: %s", event.Title, err)
		return event, err
	}
	if err := d.ReplaceEventTags(event, event.Tags, tx); err != nil {
		d.Logger.Errorf("failed to update tags for %s: %s", event.Title, err)
		return event, err
	}
	return event, nil
}

func (d *DatabaseClient) GetEventByGuid(guid string, tx *gorm.DB) (models.Event, error) {
	ret := models.Event{Guid: guid}
	for _, preload := range eventPreloads {
		tx = tx.Preload(preload)
	}
	err := tx.Where(&ret).First(&ret).Error
	if isErrorNotFoundOK(err) {
		return ret, err
	}
	return ret, nil
}

func (d *DatabaseClient) ApproveEvent(event models.Event, tx *gorm.DB) error {
	return tx.Model(&event).Update("approved_at", time.Now().Unix()).Error
}

// Used for checking if a record was previously deleted
func (d *DatabaseClient) GetEventByGuidUnscoped(guid string, tx *gorm.DB) (models.Event, error) {
	ret := models.Event{Guid: guid}
	err := tx.Unscoped().Where(&ret).First(&ret).Error
	if isErrorNotFoundOK(err) {
		return ret, err
	}
	return ret, nil
}

func (d *DatabaseClient) GetEventByVerifyKey(verifyKey string, tx *gorm.DB) (models.Event, error) {
	ret := models.Event{VerificationKey: &verifyKey}
	err := tx.Where(&ret).First(&ret).Error
	return ret, err
}

// gets events including today and 1 month forward
func (d *DatabaseClient) GetIndexEvents(filter models.CalendarFilterBackend, tx *gorm.DB) ([]models.Event, error) {
	ret := make([]models.Event, 0)
	var err error
	tx, err = d.GetTxWithFilter(filter, tx)
	if err != nil {
		return nil, err
	}
	for _, preload := range eventPreloads {
		tx = tx.Preload(preload)
	}
	// There shouldn't be a case when this isn't set properly,
	// but leaving this note here in case bugs show up later.
	// filter.UpdateDates()
	err = tx.Where(
		"? <= start_at and start_at <= ? and approved_at is not null",
		filter.FromDate, filter.ToDate,
	).Order("start_at").Find(&ret).Error
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// gets events including today and 1 month forward
func (d *DatabaseClient) GetAwaitingApprovalEvents(tx *gorm.DB) ([]models.Event, error) {
	ret := make([]models.Event, 0)
	for _, preload := range eventPreloads {
		tx = tx.Preload(preload)
	}
	err := tx.Where(
		"verified_at is not null and approved_at is null",
	).Order("start_at").Find(&ret).Error
	if err != nil {
		return nil, err
	}
	return ret, nil
}
