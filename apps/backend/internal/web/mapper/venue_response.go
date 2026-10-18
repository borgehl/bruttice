package mapper

import (
	"github.com/borgehl/bruttice/apps/backend/internal/web/dto"
	"github.com/borgehl/bruttice/packages/go-pkg/models"
)

func VenueResponses(venues []models.Venue) []dto.VenueResponse {
	ret := make([]dto.VenueResponse, 0, len(venues))
	for _, v := range venues {
		ret = append(ret, dto.VenueResponse{Name: v.Name, Slug: v.Slug, Link: v.Link})
	}
	return ret
}
