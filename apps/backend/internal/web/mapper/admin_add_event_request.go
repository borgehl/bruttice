package mapper

import (
	"time"

	"github.com/borgehl/bruttice/apps/backend/internal/web/dto"
	"github.com/borgehl/bruttice/packages/go-pkg/models"
)

// AdminAddEventRequest converts an event an admin created. It is verified and approved
// at now. It fails with ErrInvalidStartTime or ErrInvalidDoorsTime.
func AdminAddEventRequest(req dto.AdminAddEventRequest, now time.Time) (models.Event, error) {
	event, err := eventFromForm(formEvent{
		Title: req.Title, Link: req.Link, Venue: req.Venue, Region: req.Region, Age: req.Age,
		CoverCost: req.CoverCost, Date: req.Date, StartTime: req.StartTime, DoorsTime: req.DoorsTime,
		Tags: req.Tags,
	})
	if err != nil {
		return event, err
	}
	stamp := now.Unix()
	event.VerifiedAt = &stamp
	event.ApprovedAt = &stamp
	event.UpdateGuid(nil)
	return event, nil
}
