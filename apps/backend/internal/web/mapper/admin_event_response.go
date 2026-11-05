package mapper

import (
	"time"

	"github.com/borgehl/bruttice/apps/backend/internal/web/dto"
	"github.com/borgehl/bruttice/packages/go-pkg/models"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/util"
)

// AdminEventDayResponses groups events by the day they start, in the order the days first appear.
func AdminEventDayResponses(events []models.Event) []dto.AdminEventDayResponse {
	if len(events) == 0 {
		return []dto.AdminEventDayResponse{}
	}
	dayMap := make(map[string]*dto.AdminEventDayResponse)
	var days []string

	for _, event := range events {
		dayStr := event.StartAt.In(util.TZLocation).Format("Monday, January 2, 2006")
		if _, exists := dayMap[dayStr]; !exists {
			dayMap[dayStr] = &dto.AdminEventDayResponse{Day: dayStr, Events: []dto.AdminEventResponse{}}
			days = append(days, dayStr)
		}
		dayMap[dayStr].Events = append(dayMap[dayStr].Events, adminEventResponse(event))
	}

	result := make([]dto.AdminEventDayResponse, 0, len(days))
	for _, day := range days {
		result = append(result, *dayMap[day])
	}
	return result
}

func adminEventResponse(event models.Event) dto.AdminEventResponse {
	resp := dto.AdminEventResponse{
		Guid:      event.Guid,
		Title:     event.Title,
		Link:      event.Link,
		Venue:     event.Venue.Name,
		VenueLink: event.Venue.Link,
		Age:       event.Age.Name,
		Region:    event.Region.Name,
		Date:      event.StartAt.In(util.TZLocation).Format("2006-01-02"),
		Time:      formatEventTime(event),
		StartTime: event.StartAt.In(util.TZLocation).Format("15:04"),
		CoverCost: event.CoverCost,
		FlyerURL:  event.FlyerURL,
		Tags:      tagNames(event),
		Score:     event.Score(),
	}
	if event.DoorsAt != nil {
		resp.DoorsTime = event.DoorsAt.In(util.TZLocation).Format("15:04")
	}
	if event.CancelledAt != nil {
		resp.CancelledAt = time.Unix(*event.CancelledAt, 0).UTC().Format(time.RFC3339)
	}
	if event.VerificationEmail != nil {
		resp.Email = *event.VerificationEmail
	}
	return resp
}

func formatEventTime(event models.Event) string {
	if event.DoorsAt != nil && *event.DoorsAt != event.StartAt {
		return "Doors " + event.DoorsAt.In(util.TZLocation).Format("3:04pm") + ", Start " + event.StartAt.In(util.TZLocation).Format("3:04pm")
	}
	return event.StartAt.In(util.TZLocation).Format("3:04pm")
}
