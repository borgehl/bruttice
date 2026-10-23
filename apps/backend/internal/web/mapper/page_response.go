package mapper

import (
	"github.com/borgehl/bruttice/apps/backend/internal/web/dto"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/pagesx"
)

func PageSummaryResponse(page pagesx.Page) dto.PageSummaryResponse {
	return dto.PageSummaryResponse{Slug: page.Slug, Title: page.Title, Nav: page.Nav, Order: page.Order}
}

func PageSummaryResponses(pages []pagesx.Page) []dto.PageSummaryResponse {
	ret := make([]dto.PageSummaryResponse, 0, len(pages))
	for _, page := range pages {
		ret = append(ret, PageSummaryResponse(page))
	}
	return ret
}

func PageResponse(page pagesx.Page) dto.PageResponse {
	return dto.PageResponse{PageSummaryResponse: PageSummaryResponse(page), Body: page.Body, Embed: page.Embed, EmbedHTML: page.EmbedHTML}
}
