package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/borgehl/bruttice/packages/go-pkg/models"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/dbx"
	"github.com/borgehl/bruttice/packages/go-pkg/platform/logx"
)

// fakeConfig records the last call and answers with err.
type fakeConfig struct {
	calls []string
	err   error
	moved *int64
	del   bool
}

func (f *fakeConfig) rec(call string) error { f.calls = append(f.calls, call); return f.err }

func (f *fakeConfig) ListRegionRows() ([]dbx.RegionRow, error) {
	return []dbx.RegionRow{{ID: 1, Name: "Tulsa", Events: 3, Venues: 2}, {ID: 2, Name: "Old", RetiredAt: 9, MergedInto: "Tulsa"}}, f.err
}
func (f *fakeConfig) ListTagRows() ([]dbx.TagRow, error) {
	return []dbx.TagRow{{ID: 1, Name: "rock", Events: 4}}, nil
}
func (f *fakeConfig) ListAgeRows() ([]dbx.AgeRow, error) {
	return []dbx.AgeRow{{ID: 1, Name: "18+", Value: 18}}, nil
}
func (f *fakeConfig) ListVenueRows() ([]dbx.VenueRow, error) {
	return []dbx.VenueRow{{ID: 1, Name: "Hall", DefaultRegion: "Tulsa", IsScraped: true}}, nil
}
func (f *fakeConfig) CreateRegion(name string) error { return f.rec("create-region " + name) }
func (f *fakeConfig) RenameRegion(id int64, name string) error {
	return f.rec("rename-region " + name)
}
func (f *fakeConfig) RetireRegion(id int64, moveToID *int64, deleteEvents bool) (dbx.RetireResult, error) {
	f.moved, f.del = moveToID, deleteEvents
	return dbx.RetireResult{EventsMoved: 5, EventsDeleted: 0}, f.rec("retire-region")
}
func (f *fakeConfig) CreateTag(name string) error               { return f.rec("create-tag " + name) }
func (f *fakeConfig) RenameTag(id int64, n string) error        { return f.rec("rename-tag " + n) }
func (f *fakeConfig) DeleteTag(id int64) error                  { return f.rec("delete-tag") }
func (f *fakeConfig) CreateAge(n string, v int) error           { return f.rec("create-age " + n) }
func (f *fakeConfig) UpdateAge(id int64, n string, v int) error { return f.rec("update-age " + n) }
func (f *fakeConfig) DeleteAge(id int64, r *int64) (int64, error) {
	f.moved = r
	return 2, f.rec("delete-age")
}
func (f *fakeConfig) CreateVenue(n, l, r string) error { return f.rec("create-venue " + n) }
func (f *fakeConfig) UpdateVenue(id int64, n, l, r string) error {
	return f.rec("update-venue " + n)
}
func (f *fakeConfig) DeleteVenue(id int64) error { return f.rec("delete-venue") }

func configHarness(signedIn bool) (*http.ServeMux, *fakeConfig) {
	fake := &fakeConfig{}
	h := &APIHandler{
		logger: logx.NewLogger(nil, nil),
		config: fake,
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

func doConfig(mux *http.ServeMux, method, path, body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(method, path, strings.NewReader(body)))
	return rec
}

func TestAdminConfigNeedsSignIn(t *testing.T) {
	mux, fake := configHarness(false)
	for _, c := range [][2]string{
		{"GET", "/api/admin/config"},
		{"POST", "/api/admin/config/regions"},
		{"PATCH", "/api/admin/config/regions/1"},
		{"POST", "/api/admin/config/regions/1/retire"},
		{"POST", "/api/admin/config/tags"},
		{"PATCH", "/api/admin/config/tags/1"},
		{"DELETE", "/api/admin/config/tags/1"},
		{"POST", "/api/admin/config/ages"},
		{"PATCH", "/api/admin/config/ages/1"},
		{"DELETE", "/api/admin/config/ages/1"},
		{"POST", "/api/admin/config/venues"},
		{"PATCH", "/api/admin/config/venues/1"},
		{"DELETE", "/api/admin/config/venues/1"},
	} {
		if rec := doConfig(mux, c[0], c[1], `{"name":"x"}`); rec.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: status = %d", c[0], c[1], rec.Code)
		}
	}
	if len(fake.calls) != 0 {
		t.Errorf("nothing may be changed: %v", fake.calls)
	}
}

func TestAdminConfigListsEverything(t *testing.T) {
	mux, _ := configHarness(true)
	rec := doConfig(mux, "GET", "/api/admin/config", "")
	var got struct {
		Regions []struct {
			Name       string
			Retired    bool
			MergedInto string
			Events     int64
		}
		Tags   []struct{ Name string }
		Ages   []struct{ Value int }
		Venues []struct {
			Region  string
			Scraped bool
		}
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil || rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
	}
	if len(got.Regions) != 2 || got.Regions[0].Events != 3 || !got.Regions[1].Retired || got.Regions[1].MergedInto != "Tulsa" {
		t.Errorf("regions = %+v", got.Regions)
	}
	if len(got.Tags) != 1 || got.Ages[0].Value != 18 || !got.Venues[0].Scraped || got.Venues[0].Region != "Tulsa" {
		t.Errorf("response = %s", rec.Body)
	}
}

func TestAdminConfigChangesAreForwarded(t *testing.T) {
	mux, fake := configHarness(true)
	for _, c := range []struct{ method, path, body, call string }{
		{"POST", "/api/admin/config/regions", `{"name":"Norman"}`, "create-region Norman"},
		{"PATCH", "/api/admin/config/regions/3", `{"name":"OKC"}`, "rename-region OKC"},
		{"POST", "/api/admin/config/tags", `{"name":"punk"}`, "create-tag punk"},
		{"PATCH", "/api/admin/config/tags/3", `{"name":"hardcore"}`, "rename-tag hardcore"},
		{"DELETE", "/api/admin/config/tags/3", ``, "delete-tag"},
		{"POST", "/api/admin/config/ages", `{"name":"21+","value":21}`, "create-age 21+"},
		{"PATCH", "/api/admin/config/ages/3", `{"name":"All","value":0}`, "update-age All"},
		{"POST", "/api/admin/config/venues", `{"name":"Hall","link":"","region":"Tulsa"}`, "create-venue Hall"},
		{"PATCH", "/api/admin/config/venues/3", `{"name":"Hall 2","link":"","region":"Tulsa"}`, "update-venue Hall 2"},
		{"DELETE", "/api/admin/config/venues/3", ``, "delete-venue"},
	} {
		fake.calls = nil
		if rec := doConfig(mux, c.method, c.path, c.body); rec.Code != http.StatusOK || len(fake.calls) != 1 || fake.calls[0] != c.call {
			t.Errorf("%s %s: status = %d, calls = %v", c.method, c.path, rec.Code, fake.calls)
		}
	}
}

func TestAdminConfigRetireRegionPassesTheChoice(t *testing.T) {
	mux, fake := configHarness(true)
	rec := doConfig(mux, "POST", "/api/admin/config/regions/3/retire", `{"moveToId":7}`)
	var got struct{ EventsMoved int64 }
	_ = json.Unmarshal(rec.Body.Bytes(), &got)
	if rec.Code != http.StatusOK || fake.moved == nil || *fake.moved != 7 || fake.del || got.EventsMoved != 5 {
		t.Errorf("status = %d, moved = %v, delete = %v, body = %s", rec.Code, fake.moved, fake.del, rec.Body)
	}
	doConfig(mux, "POST", "/api/admin/config/regions/3/retire", `{"deleteEvents":true}`)
	if fake.moved != nil || !fake.del {
		t.Errorf("moved = %v, delete = %v", fake.moved, fake.del)
	}
}

func TestAdminConfigDeleteAgeTakesAReplacement(t *testing.T) {
	mux, fake := configHarness(true)
	doConfig(mux, "DELETE", "/api/admin/config/ages/3?replacementId=4", "")
	if fake.moved == nil || *fake.moved != 4 {
		t.Errorf("replacement = %v", fake.moved)
	}
	fake.moved = nil
	doConfig(mux, "DELETE", "/api/admin/config/ages/3", "")
	if fake.moved != nil {
		t.Errorf("replacement = %v", fake.moved)
	}
	if rec := doConfig(mux, "DELETE", "/api/admin/config/ages/3?replacementId=x", ""); rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d", rec.Code)
	}
}

func TestAdminConfigErrors(t *testing.T) {
	mux, fake := configHarness(true)
	cases := map[error]int{
		&dbx.ConfigError{Kind: dbx.ConfigInvalid, Message: "bad"}:   http.StatusBadRequest,
		&dbx.ConfigError{Kind: dbx.ConfigNotFound, Message: "gone"}: http.StatusNotFound,
		&dbx.ConfigError{Kind: dbx.ConfigConflict, Message: "used"}: http.StatusConflict,
	}
	for err, want := range cases {
		fake.err = err
		rec := doConfig(mux, "POST", "/api/admin/config/tags", `{"name":"x"}`)
		if rec.Code != want || !strings.Contains(rec.Body.String(), err.Error()) {
			t.Errorf("%v: status = %d, body = %s", err, rec.Code, rec.Body)
		}
	}
	fake.err = http.ErrAbortHandler
	rec := doConfig(mux, "POST", "/api/admin/config/tags", `{"name":"x"}`)
	if rec.Code != http.StatusInternalServerError || strings.Contains(rec.Body.String(), "abort") {
		t.Errorf("an unexpected error must not leak: status = %d, body = %s", rec.Code, rec.Body)
	}
	if rec := doConfig(mux, "PATCH", "/api/admin/config/tags/zero", `{"name":"x"}`); rec.Code != http.StatusBadRequest {
		t.Errorf("bad id: status = %d", rec.Code)
	}
}
