package configx

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeConfig(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestParseConfigSite(t *testing.T) {
	path := writeConfig(t, `
timezone: America/Chicago
admin-accounts:
  - "a@example.com;pw"
site:
  name: "Example Events"
  region: "Tulsa"
  contact-email: "info@example.com"
  description: "A calendar."
pages-dir: /srv/pages
`)
	cfg := ParseConfig(&path)
	want := ConfigRawSite{
		Name: "Example Events", Region: "Tulsa", ContactEmail: "info@example.com",
		Description: "A calendar.",
	}
	if cfg.Site != want {
		t.Errorf("site = %+v, want %+v", cfg.Site, want)
	}
	if cfg.PagesDir != "/srv/pages" {
		t.Errorf("pages dir = %q", cfg.PagesDir)
	}
}

func TestParseConfigWithoutSite(t *testing.T) {
	path := writeConfig(t, "timezone: America/Chicago\nadmin-accounts:\n  - \"a@example.com;pw\"\n")
	cfg := ParseConfig(&path)
	if cfg.Site != (ConfigRawSite{}) {
		t.Errorf("site = %+v, want the zero value", cfg.Site)
	}
	if cfg.PagesDir != DefaultPagesDir {
		t.Errorf("pages dir = %q, want the default", cfg.PagesDir)
	}
}

func TestDemoConfigParses(t *testing.T) {
	path := filepath.Join("..", "..", "..", "..", "demo", "config.yaml")
	if _, err := os.Stat(path); err != nil {
		t.Skip("demo config not found")
	}
	cfg := ParseConfig(&path)
	if cfg.Site.Name == "" {
		t.Error("demo config has no site name")
	}
}

const adminFile = "timezone: America/Chicago\nadmin-accounts:\n  - \"file@example.com;filepw\"\n"

func TestAdminAccountsEnvReplacesFile(t *testing.T) {
	t.Setenv(AdminAccountsEnv, " root@example.com;pw1 , second@example.com;pw2,")
	path := writeConfig(t, adminFile)
	got := ParseConfig(&path).AdminAccounts
	want := []AdminAccount{{"root@example.com", "pw1"}, {"second@example.com", "pw2"}}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("accounts = %+v, want %+v", got, want)
	}
}

func TestAdminAccountsEnvUnsetKeepsFile(t *testing.T) {
	for _, value := range []string{"", "  "} {
		t.Setenv(AdminAccountsEnv, value)
		path := writeConfig(t, adminFile)
		got := ParseConfig(&path).AdminAccounts
		if len(got) != 1 || got[0].Email != "file@example.com" {
			t.Errorf("env %q: accounts = %+v, want the file's", value, got)
		}
	}
}

func TestAdminAccountsEnvMalformedHidesPassword(t *testing.T) {
	// a comma in a password leaves a piece without ";"
	t.Setenv(AdminAccountsEnv, "a@example.com;pw,secret-tail")
	path := writeConfig(t, adminFile)
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("want a panic for a malformed account")
		}
		msg := fmt.Sprint(r)
		if strings.Contains(msg, "secret-tail") || strings.Contains(msg, "pw") {
			t.Errorf("error leaks the account: %s", msg)
		}
	}()
	ParseConfig(&path)
}
