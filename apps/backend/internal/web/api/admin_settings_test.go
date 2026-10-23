package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/borgehl/bruttice/apps/backend/internal/web/dto"
	"github.com/borgehl/bruttice/packages/go-pkg/models"
)

func TestSettingsNeedSignIn(t *testing.T) {
	hr := newHarness(t, true, true)
	if rec := hr.do(http.MethodPatch, "/api/admin/settings", `{"name":"Sam"}`); rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d", rec.Code)
	}
	if len(hr.users.names) != 0 {
		t.Error("nothing may be saved")
	}
}

func TestSettingsSaveTheNameOfTheSignedInAdmin(t *testing.T) {
	hr := newHarness(t, true, true)
	hr.signedIn = true
	rec := hr.do(http.MethodPatch, "/api/admin/settings", `{"name":"  Sam \t Jones\n"}`)
	var got dto.LoginResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK || got.Name != "Sam Jones" || got.Email != "admin@example.com" {
		t.Errorf("status = %d, response = %+v", rec.Code, got)
	}
	if hr.users.names[7] != "Sam Jones" {
		t.Errorf("saved names = %v", hr.users.names)
	}
}

func TestSettingsCleanTheName(t *testing.T) {
	hr := newHarness(t, true, true)
	hr.signedIn = true
	hr.do(http.MethodPatch, "/api/admin/settings", "{\"name\":\"Sam\\u0000\\u0007 Jones\"}")
	if hr.users.names[7] != "Sam Jones" {
		t.Errorf("control characters must go: %q", hr.users.names[7])
	}
}

func TestSettingsCanClearTheName(t *testing.T) {
	hr := newHarness(t, true, true)
	hr.signedIn = true
	if rec := hr.do(http.MethodPatch, "/api/admin/settings", `{"name":"   "}`); rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if name, ok := hr.users.names[7]; !ok || name != "" {
		t.Errorf("an empty name goes back to the email: %v", hr.users.names)
	}
}

func TestSettingsRefuseALongName(t *testing.T) {
	hr := newHarness(t, true, true)
	hr.signedIn = true
	rec := hr.do(http.MethodPatch, "/api/admin/settings", `{"name":"`+strings.Repeat("a", maxAdminName+1)+`"}`)
	if rec.Code != http.StatusBadRequest || len(hr.users.names) != 0 {
		t.Errorf("status = %d, saved = %v", rec.Code, hr.users.names)
	}
	if rec := hr.do(http.MethodPatch, "/api/admin/settings", `{"name":"`+strings.Repeat("é", maxAdminName)+`"}`); rec.Code != http.StatusOK {
		t.Errorf("the limit counts characters, not bytes: status = %d", rec.Code)
	}
}

func TestSettingsReportAFailedSave(t *testing.T) {
	hr := newHarness(t, true, true)
	hr.signedIn = true
	hr.users.fail = true
	rec := hr.do(http.MethodPatch, "/api/admin/settings", `{"name":"Sam"}`)
	if rec.Code != http.StatusInternalServerError || strings.Contains(rec.Body.String(), "database") {
		t.Errorf("status = %d, body = %s", rec.Code, rec.Body)
	}
}

func TestAdminReplyIsSignedWithTheAdminsName(t *testing.T) {
	hr := newHarness(t, true, true)
	hr.signedIn = true
	hr.adminName = "Sam"
	hr.seedThread(t)
	hr.do(http.MethodPost, "/api/admin/inbox/"+testThread+"/reply", `{"body":"Yes"}`)

	saved := hr.inbox.all()
	if len(saved) != 2 || saved[1].Source != models.InboxSourceAdminReply || saved[1].FromName != "Sam" {
		t.Errorf("saved = %+v", saved)
	}
}
