package mapper

import (
	"testing"

	"github.com/borgehl/bruttice/packages/go-pkg/platform/pagesx"
)

func TestPageResponses(t *testing.T) {
	page := pagesx.Page{Slug: "about", Title: "About", Nav: true, Order: 3, Embed: "https://x.example", Body: "Hi"}

	full := PageResponse(page)
	if full.Slug != "about" || full.Title != "About" || !full.Nav || full.Order != 3 || full.Body != "Hi" || full.Embed != "https://x.example" {
		t.Errorf("page = %+v", full)
	}

	list := PageSummaryResponses([]pagesx.Page{page})
	if len(list) != 1 || list[0].Slug != "about" {
		t.Errorf("list = %+v", list)
	}
	if got := PageSummaryResponses(nil); got == nil || len(got) != 0 {
		t.Errorf("no pages must be an empty list, not null: %#v", got)
	}
}
