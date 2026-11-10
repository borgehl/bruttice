package dbx

import (
	"errors"
	"strings"
	"testing"

	"github.com/borgehl/bruttice/packages/go-pkg/models"
)

func validSponsor() SponsorInput {
	return SponsorInput{Name: "  Acme   Records ", Text: "Vinyl, tapes", Link: "https://example.com/shop", Enabled: true}
}

func TestCleanSponsorTidiesFields(t *testing.T) {
	got, err := cleanSponsor(validSponsor())
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Acme Records" {
		t.Errorf("name = %q", got.Name)
	}
}

func TestCleanSponsorRefuses(t *testing.T) {
	cases := map[string]func(*SponsorInput){
		"no name":           func(s *SponsorInput) { s.Name = "  " },
		"long name":         func(s *SponsorInput) { s.Name = strings.Repeat("a", maxSponsorName+1) },
		"long text":         func(s *SponsorInput) { s.Text = strings.Repeat("a", maxSponsorText+1) },
		"control in text":   func(s *SponsorInput) { s.Text = "a\x00b" },
		"no link":           func(s *SponsorInput) { s.Link = "" },
		"not a web link":    func(s *SponsorInput) { s.Link = "javascript:alert(1)" },
		"link without host": func(s *SponsorInput) { s.Link = "https://" },
		"relative link":     func(s *SponsorInput) { s.Link = "/shop" },
		"bad first day":     func(s *SponsorInput) { s.StartsOn = "11/10/2026" },
		"bad last day":      func(s *SponsorInput) { s.EndsOn = "soon" },
		"last before first": func(s *SponsorInput) { s.StartsOn, s.EndsOn = "2026-11-10", "2026-11-09" },
	}
	for name, change := range cases {
		in := validSponsor()
		change(&in)
		_, err := cleanSponsor(in)
		var configErr *ConfigError
		if !errors.As(err, &configErr) || configErr.Kind != ConfigInvalid {
			t.Errorf("%s: err = %v, want a ConfigInvalid", name, err)
		}
	}
}

func TestCleanSponsorAllowsOpenDatesAndNoText(t *testing.T) {
	in := validSponsor()
	in.Text = ""
	in.StartsOn, in.EndsOn = "", "2026-12-31"
	if _, err := cleanSponsor(in); err != nil {
		t.Errorf("err = %v", err)
	}
}

func TestCleanSponsorImage(t *testing.T) {
	png := []byte("\x89PNG\r\n\x1a\nrest")
	if kind, err := cleanSponsorImage(png); err != nil || kind != "image/png" {
		t.Errorf("png: kind = %q, err = %v", kind, err)
	}
	for name, image := range map[string][]byte{
		"svg":     []byte(`<svg xmlns="http://www.w3.org/2000/svg"></svg>`),
		"html":    []byte("<html><script>alert(1)</script></html>"),
		"text":    []byte("hello"),
		"too big": append([]byte("\x89PNG\r\n\x1a\n"), make([]byte, MaxSponsorImage)...),
	} {
		if _, err := cleanSponsorImage(image); err == nil {
			t.Errorf("%s: want an error", name)
		}
	}
}

func TestSponsorInputApplyPicture(t *testing.T) {
	png := []byte("\x89PNG\r\n\x1a\nrest")
	row := models.Sponsor{Image: []byte("old"), ImageType: "image/gif"}

	// no new picture keeps the old one
	in, _ := cleanSponsor(validSponsor())
	if err := in.apply(&row); err != nil || string(row.Image) != "old" {
		t.Errorf("keep: image = %q, err = %v", row.Image, err)
	}

	in.Image = png
	if err := in.apply(&row); err != nil || row.ImageType != "image/png" {
		t.Errorf("replace: type = %q, err = %v", row.ImageType, err)
	}

	in.Image, in.RemoveImage = nil, true
	if err := in.apply(&row); err != nil || row.Image != nil || row.ImageType != "" {
		t.Errorf("remove: image = %q, type = %q, err = %v", row.Image, row.ImageType, err)
	}
}
