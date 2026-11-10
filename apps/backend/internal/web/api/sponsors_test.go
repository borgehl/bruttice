package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/borgehl/bruttice/apps/backend/internal/web/dto"
	"github.com/borgehl/bruttice/packages/go-pkg/models"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/dbx"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/logx"
)

var testPNG = []byte("\x89PNG\r\n\x1a\nrest")

type fakeSponsors struct {
	sponsors []models.Sponsor
	calls    []string
	counts   []dbx.SponsorCount
	input    dbx.SponsorInput
	err      error
}

func (f *fakeSponsors) ListSponsorRows(today string) ([]dbx.SponsorRow, error) {
	return []dbx.SponsorRow{{ID: 1, Name: "Acme", Link: "https://example.com", Enabled: true, HasImage: true, UpdatedAt: 7, Impressions: 10, Clicks: 2}}, f.err
}
func (f *fakeSponsors) ListSponsors() ([]models.Sponsor, error) { return f.sponsors, f.err }
func (f *fakeSponsors) GetSponsor(id int64) (models.Sponsor, error) {
	for _, s := range f.sponsors {
		if s.ID == id {
			return s, nil
		}
	}
	return models.Sponsor{}, &dbx.ConfigError{Kind: dbx.ConfigNotFound, Message: "That sponsor does not exist"}
}
func (f *fakeSponsors) SponsorImage(id int64) ([]byte, string, error) {
	for _, s := range f.sponsors {
		if s.ID == id {
			return s.Image, s.ImageType, nil
		}
	}
	return nil, "", &dbx.ConfigError{Kind: dbx.ConfigNotFound, Message: "That sponsor does not exist"}
}
func (f *fakeSponsors) CreateSponsor(in dbx.SponsorInput) error {
	f.input = in
	f.calls = append(f.calls, "create")
	return f.err
}
func (f *fakeSponsors) UpdateSponsor(id int64, in dbx.SponsorInput) error {
	f.input = in
	f.calls = append(f.calls, "update")
	return f.err
}
func (f *fakeSponsors) DeleteSponsor(id int64) error {
	f.calls = append(f.calls, "delete")
	return f.err
}
func (f *fakeSponsors) CountSponsor(id int64, kind dbx.SponsorCount, today string) error {
	f.counts = append(f.counts, kind)
	return nil
}

func sponsorHarness(signedIn bool, sponsors ...models.Sponsor) (*http.ServeMux, *fakeSponsors) {
	fake := &fakeSponsors{sponsors: sponsors}
	h := &APIHandler{
		logger:       logx.NewLogger(nil, nil),
		sponsors:     fake,
		sponsorLimit: newRateLimiter(time.Hour, 2, 100),
		authenticate: func(*http.Request) *models.User {
			if signedIn {
				return &models.User{Email: "admin@example.com"}
			}
			return nil
		},
	}
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	return mux, fake
}

func activeSponsor() models.Sponsor {
	return models.Sponsor{BaseModel: models.BaseModel{ID: 1, UpdatedAt: 7}, Name: "Acme", Text: "Vinyl", Link: "https://example.com/shop", Enabled: true, Image: testPNG, ImageType: "image/png"}
}

func TestSponsorsListsOnlyActiveWithoutTheLink(t *testing.T) {
	hidden := activeSponsor()
	hidden.ID, hidden.Enabled = 2, false
	mux, _ := sponsorHarness(false, activeSponsor(), hidden)

	rec := doConfig(mux, "GET", "/api/sponsors", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "example.com") {
		t.Errorf("the sponsor's own link must not be sent: %s", rec.Body.String())
	}
	var got []dto.SponsorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != 1 || got[0].ClickURL != "/api/sponsors/1/click" || got[0].ImageURL != "/api/sponsors/1/image?v=7" {
		t.Errorf("got %+v", got)
	}
}

func TestSponsorsListIsEmptyJSONArray(t *testing.T) {
	mux, _ := sponsorHarness(false)
	if body := strings.TrimSpace(doConfig(mux, "GET", "/api/sponsors", "").Body.String()); body != "[]" {
		t.Errorf("body = %s", body)
	}
}

func TestSponsorImage(t *testing.T) {
	noImage := activeSponsor()
	noImage.ID, noImage.Image, noImage.ImageType = 2, nil, ""
	mux, _ := sponsorHarness(false, activeSponsor(), noImage)

	rec := doConfig(mux, "GET", "/api/sponsors/1/image", "")
	if rec.Code != http.StatusOK || rec.Header().Get("Content-Type") != "image/png" {
		t.Errorf("status = %d, type = %q", rec.Code, rec.Header().Get("Content-Type"))
	}
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("the picture needs nosniff")
	}
	if rec := doConfig(mux, "GET", "/api/sponsors/2/image", ""); rec.Code != http.StatusNotFound {
		t.Errorf("no picture: status = %d", rec.Code)
	}
	if rec := doConfig(mux, "GET", "/api/sponsors/9/image", ""); rec.Code != http.StatusNotFound {
		t.Errorf("no sponsor: status = %d", rec.Code)
	}
}

func TestSponsorClickCountsAndRedirects(t *testing.T) {
	mux, fake := sponsorHarness(false, activeSponsor())
	rec := doConfig(mux, "GET", "/api/sponsors/1/click", "")
	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "https://example.com/shop" {
		t.Errorf("status = %d, location = %q", rec.Code, rec.Header().Get("Location"))
	}
	if len(fake.counts) != 1 || fake.counts[0] != dbx.SponsorClick {
		t.Errorf("counts = %v", fake.counts)
	}
}

func TestSponsorImpressionCounts(t *testing.T) {
	mux, fake := sponsorHarness(false, activeSponsor())
	if rec := doConfig(mux, "POST", "/api/sponsors/1/impression", ""); rec.Code != http.StatusNoContent {
		t.Errorf("status = %d", rec.Code)
	}
	if len(fake.counts) != 1 || fake.counts[0] != dbx.SponsorImpression {
		t.Errorf("counts = %v", fake.counts)
	}
}

func TestInactiveSponsorIsNeitherCountedNorFollowed(t *testing.T) {
	off := activeSponsor()
	off.Enabled = false
	mux, fake := sponsorHarness(false, off)
	for _, c := range [][2]string{{"POST", "/api/sponsors/1/impression"}, {"GET", "/api/sponsors/1/click"}, {"GET", "/api/sponsors/9/click"}} {
		if rec := doConfig(mux, c[0], c[1], ""); rec.Code != http.StatusNotFound {
			t.Errorf("%s %s: status = %d", c[0], c[1], rec.Code)
		}
	}
	if len(fake.counts) != 0 {
		t.Errorf("counts = %v", fake.counts)
	}
}

func TestSponsorCountsAreRateLimitedButClicksStillRedirect(t *testing.T) {
	mux, fake := sponsorHarness(false, activeSponsor())
	for i := 0; i < 4; i++ {
		if rec := doConfig(mux, "GET", "/api/sponsors/1/click", ""); rec.Code != http.StatusFound {
			t.Fatalf("click %d: status = %d", i, rec.Code)
		}
	}
	if len(fake.counts) != 2 {
		t.Errorf("counted %d, the limit is 2", len(fake.counts))
	}
}

func TestSponsorIDMustBeANumber(t *testing.T) {
	mux, _ := sponsorHarness(false, activeSponsor())
	if rec := doConfig(mux, "GET", "/api/sponsors/abc/click", ""); rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d", rec.Code)
	}
}

func TestAdminSponsorsNeedSignIn(t *testing.T) {
	mux, fake := sponsorHarness(false, activeSponsor())
	for _, c := range [][2]string{
		{"GET", "/api/admin/sponsors"},
		{"POST", "/api/admin/sponsors"},
		{"PATCH", "/api/admin/sponsors/1"},
		{"DELETE", "/api/admin/sponsors/1"},
	} {
		if rec := doConfig(mux, c[0], c[1], `{"name":"x"}`); rec.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: status = %d", c[0], c[1], rec.Code)
		}
	}
	if len(fake.calls) != 0 {
		t.Errorf("nothing may be changed: %v", fake.calls)
	}
}

func TestAdminSponsorsList(t *testing.T) {
	mux, _ := sponsorHarness(true)
	rec := doConfig(mux, "GET", "/api/admin/sponsors", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var got []dto.AdminSponsorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Link != "https://example.com" || !got[0].Active || got[0].Clicks != 2 || got[0].ImageURL != "/api/sponsors/1/image?v=7" {
		t.Errorf("got %+v", got)
	}
}

func TestAdminSponsorCreateDecodesThePicture(t *testing.T) {
	mux, fake := sponsorHarness(true)
	body := `{"name":"Acme","link":"https://example.com","enabled":true,"image":"` + base64.StdEncoding.EncodeToString(testPNG) + `"}`
	if rec := doConfig(mux, "POST", "/api/admin/sponsors", body); rec.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body.String())
	}
	if len(fake.calls) != 1 || fake.calls[0] != "create" || string(fake.input.Image) != string(testPNG) || !fake.input.Enabled {
		t.Errorf("calls = %v, input = %+v", fake.calls, fake.input)
	}
}

func TestAdminSponsorUpdateWithoutPictureKeepsIt(t *testing.T) {
	mux, fake := sponsorHarness(true)
	if rec := doConfig(mux, "PATCH", "/api/admin/sponsors/1", `{"name":"Acme","link":"https://example.com"}`); rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if fake.input.Image != nil || fake.input.RemoveImage {
		t.Errorf("input = %+v", fake.input)
	}
}

func TestAdminSponsorBadPictureEncoding(t *testing.T) {
	mux, fake := sponsorHarness(true)
	if rec := doConfig(mux, "POST", "/api/admin/sponsors", `{"name":"Acme","link":"https://example.com","image":"!!!"}`); rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d", rec.Code)
	}
	if len(fake.calls) != 0 {
		t.Errorf("calls = %v", fake.calls)
	}
}

func TestAdminSponsorShowsWhyAChangeFailed(t *testing.T) {
	mux, fake := sponsorHarness(true)
	fake.err = &dbx.ConfigError{Kind: dbx.ConfigInvalid, Message: "Enter the sponsor's name"}
	rec := doConfig(mux, "POST", "/api/admin/sponsors", `{"link":"https://example.com"}`)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "Enter the sponsor's name") {
		t.Errorf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestAdminSponsorDelete(t *testing.T) {
	mux, fake := sponsorHarness(true)
	if rec := doConfig(mux, "DELETE", "/api/admin/sponsors/1", ""); rec.Code != http.StatusOK {
		t.Errorf("status = %d", rec.Code)
	}
	if len(fake.calls) != 1 || fake.calls[0] != "delete" {
		t.Errorf("calls = %v", fake.calls)
	}
}
