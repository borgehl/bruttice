package mapper

import (
	"github.com/borgehl/bruttice/apps/backend/internal/web/dto"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/configx"
)

// httpURL keeps a link only if it is http(s): the embedded calendar is never loaded from another scheme
func httpURL(url string) string { return FlyerURL(url) }

func SiteResponse(site configx.ConfigRawSite) dto.SiteResponse {
	return dto.SiteResponse{
		Name:         site.Name,
		Region:       site.Region,
		ContactEmail: site.ContactEmail,
		Description:  site.Description,
		Features:     dto.SiteFeatures{PopBooker: site.PopBooker},
		PopBookerURL: httpURL(site.PopBookerURL),
	}
}
