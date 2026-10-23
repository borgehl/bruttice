package mapper

import (
	"github.com/borgehl/bruttice/apps/backend/internal/web/dto"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/configx"
)

func SiteResponse(site configx.ConfigRawSite) dto.SiteResponse {
	return dto.SiteResponse{
		Name:        site.Name,
		Region:      site.Region,
		Description: site.Description,
	}
}
