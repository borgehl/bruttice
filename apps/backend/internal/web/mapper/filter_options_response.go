package mapper

import (
	"github.com/borgehl/bruttice/apps/backend/internal/web/dto"
	"github.com/borgehl/bruttice/packages/go-pkg/models"
)

// maxFilterAgeValue is the highest age value offered. Above it is "unknown age limit".
const maxFilterAgeValue = 100

// FilterOptionsResponse lists what the calendar filter and the add event form can choose from.
// Tags and regions are named, so their ids are their position in the list.
func FilterOptionsResponse(tagNames []string, venues []models.Venue, ages []models.Age, regionNames []string) dto.FilterOptionsResponse {
	tags := make([]dto.TagResponse, 0, len(tagNames))
	for i, name := range tagNames {
		tags = append(tags, dto.TagResponse{ID: int64(i + 1), Name: name})
	}

	choosable := make([]models.Age, 0, len(ages))
	for _, age := range ages {
		if age.Value > maxFilterAgeValue {
			continue
		}
		choosable = append(choosable, age)
	}

	regions := make([]dto.RegionResponse, 0, len(regionNames))
	for i, name := range regionNames {
		regions = append(regions, dto.RegionResponse{ID: int64(i + 1), Name: name})
	}

	return dto.FilterOptionsResponse{
		Tags:    tags,
		Venues:  VenueResponses(venues),
		Ages:    AgeResponses(choosable),
		Regions: regions,
	}
}
