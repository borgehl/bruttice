package configx

import (
	"os"
	"path/filepath"
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
