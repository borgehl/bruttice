package mapper

import (
	"fmt"

	"github.com/borgehl/bruttice/apps/backend/internal/web/dto"
	"github.com/borgehl/bruttice/packages/go-pkg/models"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/dbx"
)

// sponsorImageURL is where a sponsor's picture is served. updatedAt is in the address so a new
// picture is fetched again, and the old one can be cached for a long time.
func sponsorImageURL(id, updatedAt int64) string {
	return fmt.Sprintf("/api/sponsors/%d/image?v=%d", id, updatedAt)
}

// SponsorResponses are the sponsors that are active on day. Sponsors come without their picture.
func SponsorResponses(sponsors []models.Sponsor, day string) []dto.SponsorResponse {
	ret := make([]dto.SponsorResponse, 0, len(sponsors))
	for _, s := range sponsors {
		if !s.ActiveOn(day) {
			continue
		}
		imageURL := ""
		if s.ImageType != "" {
			imageURL = sponsorImageURL(s.ID, s.UpdatedAt)
		}
		ret = append(ret, dto.SponsorResponse{
			ID:       s.ID,
			Name:     s.Name,
			Text:     s.Text,
			ImageURL: imageURL,
			ClickURL: fmt.Sprintf("/api/sponsors/%d/click", s.ID),
		})
	}
	return ret
}

// AdminSponsorResponses are all sponsors for the admin. day is today, for whether each is shown.
func AdminSponsorResponses(rows []dbx.SponsorRow, day string) []dto.AdminSponsorResponse {
	ret := make([]dto.AdminSponsorResponse, 0, len(rows))
	for _, r := range rows {
		imageURL := ""
		if r.HasImage {
			imageURL = sponsorImageURL(r.ID, r.UpdatedAt)
		}
		active := models.Sponsor{Enabled: r.Enabled, StartsOn: r.StartsOn, EndsOn: r.EndsOn}.ActiveOn(day)
		ret = append(ret, dto.AdminSponsorResponse{
			ID: r.ID, Name: r.Name, Text: r.Text, Link: r.Link,
			StartsOn: r.StartsOn, EndsOn: r.EndsOn, Enabled: r.Enabled,
			ImageURL: imageURL, Active: active,
			Impressions: r.Impressions, Clicks: r.Clicks,
			Impressions30: r.Impressions30, Clicks30: r.Clicks30,
		})
	}
	return ret
}
