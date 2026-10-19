package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/borgehl/bruttice/apps/backend/internal/web/dto"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/logx"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/pagesx"
)

func pagesServer(pages []pagesx.Page) *http.ServeMux {
	h := &APIHandler{logger: logx.NewLogger(nil, nil), pages: pages}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/pages", h.handleGetPages)
	mux.HandleFunc("GET /api/pages/{slug}", h.handleGetPage)
	return mux
}

func get(mux *http.ServeMux, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func TestGetPages(t *testing.T) {
	mux := pagesServer([]pagesx.Page{{Slug: "about", Title: "About", Nav: true, Body: "Hi"}})

	rec := get(mux, "/api/pages")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var list []dto.PageSummaryResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].Slug != "about" || !list[0].Nav {
		t.Errorf("list = %+v", list)
	}
	if body := rec.Body.String(); strings.Contains(body, "Hi") {
		t.Errorf("the list must not carry page bodies: %s", body)
	}
}

func TestGetPagesWithNone(t *testing.T) {
	if body := get(pagesServer(nil), "/api/pages").Body.String(); body != "[]\n" {
		t.Errorf("body = %q, want an empty list", body)
	}
}

func TestGetPage(t *testing.T) {
	mux := pagesServer([]pagesx.Page{{Slug: "faq", Title: "FAQ", Body: "Answers", Embed: "https://x.example"}})

	rec := get(mux, "/api/pages/faq")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var page dto.PageResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	if page.Title != "FAQ" || page.Body != "Answers" || page.Embed != "https://x.example" {
		t.Errorf("page = %+v", page)
	}

	if rec := get(mux, "/api/pages/nope"); rec.Code != http.StatusNotFound {
		t.Errorf("missing page status = %d, want 404", rec.Code)
	}
}
