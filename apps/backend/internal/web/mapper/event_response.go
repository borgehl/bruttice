package mapper

import (
	"time"

	"github.com/borgehl/bruttice/apps/backend/internal/web/dto"
	"github.com/borgehl/bruttice/packages/go-pkg/models"
)

func EventResponse(event models.Event) dto.EventResponse {
	resp := dto.EventResponse{
		Guid:      event.Guid,
		Title:     event.Title,
		Link:      event.Link,
		Venue:     event.Venue.Name,
		VenueLink: event.Venue.Link,
		Age:       event.Age.Name,
		Region:    event.Region.Name,
		StartAt:   event.StartAt.Format(time.RFC3339),
		CoverCost: event.CoverCost,
		FlyerURL:  event.FlyerURL,
		Tags:      tagNames(event),
	}
	if event.DoorsAt != nil && *event.DoorsAt != event.StartAt {
		resp.DoorsAt = event.DoorsAt.UTC().Format(time.RFC3339)
	}
	return resp
}

func EventResponses(events []models.Event) []dto.EventResponse {
	ret := make([]dto.EventResponse, 0, len(events))
	for _, event := range events {
		ret = append(ret, EventResponse(event))
	}
	return ret
}

func tagNames(event models.Event) []string {
	tags := make([]string, 0, len(event.Tags))
	for _, tag := range event.Tags {
		tags = append(tags, tag.Name)
	}
	return tags
}
