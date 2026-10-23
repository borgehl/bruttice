package pagesx

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseHeader(t *testing.T) {
	p, err := Parse("our-faq", "---\ntitle: FAQ\nnav: true\norder: 60\nembed: https://x.example/cal\n---\nHello *there*.\n")
	if err != nil {
		t.Fatal(err)
	}
	want := Page{Slug: "our-faq", Title: "FAQ", Nav: true, Order: 60, Embed: "https://x.example/cal", Body: "Hello *there*."}
	if p != want {
		t.Errorf("page = %+v, want %+v", p, want)
	}
}

func TestParseEmptyHeader(t *testing.T) {
	p, err := Parse("about", "---\n---\nBody")
	if err != nil {
		t.Fatal(err)
	}
	if p.Title != "About" || p.Body != "Body" {
		t.Errorf("page = %+v", p)
	}
}

func TestParseWithoutHeader(t *testing.T) {
	p, err := Parse("our-story", "Just text.")
	if err != nil {
		t.Fatal(err)
	}
	if p.Title != "Our Story" || p.Body != "Just text." || p.Nav || p.Order != 0 {
		t.Errorf("page = %+v", p)
	}
}

func TestParseWindowsLineEndings(t *testing.T) {
	p, err := Parse("about", "---\r\ntitle: About\r\n---\r\nBody\r\n")
	if err != nil {
		t.Fatal(err)
	}
	if p.Title != "About" || p.Body != "Body" {
		t.Errorf("page = %+v", p)
	}
}

func TestParseKeepsDashesInTheBody(t *testing.T) {
	p, err := Parse("about", "---\ntitle: About\n---\nAbove\n\n---\n\nBelow\n")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(p.Body, "Above") || !strings.Contains(p.Body, "---") || !strings.Contains(p.Body, "Below") {
		t.Errorf("body = %q", p.Body)
	}
}

func TestParseDropsANonHTTPEmbed(t *testing.T) {
	p, err := Parse("a", "---\nembed: javascript:alert(1)\n---\nx")
	if err != nil {
		t.Fatal(err)
	}
	if p.Embed != "" {
		t.Errorf("embed = %q", p.Embed)
	}
}

func TestParseErrors(t *testing.T) {
	for name, tc := range map[string]struct{ slug, text string }{
		"bad slug":        {"About Us", "x"},
		"upper case slug": {"About", "x"},
		"unclosed header": {"a", "---\ntitle: x\nbody"},
		"bad yaml":        {"a", "---\ntitle: [\n---\nbody"},
	} {
		if _, err := Parse(tc.slug, tc.text); err == nil {
			t.Errorf("%s: no error", name)
		}
	}
}

func write(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "b.md", "---\ntitle: B\norder: 2\n---\nb")
	write(t, dir, "a.md", "---\ntitle: A\norder: 2\n---\na")
	write(t, dir, "first.md", "---\ntitle: First\norder: 1\n---\nf")
	write(t, dir, "notes.txt", "ignored")
	write(t, dir, ".hidden.md", "ignored")
	write(t, dir, "Bad Name.md", "skipped")
	write(t, dir, "broken.md", "---\ntitle: [\n---\n")

	pages, problems := Load(dir)
	var slugs []string
	for _, p := range pages {
		slugs = append(slugs, p.Slug)
	}
	if got := strings.Join(slugs, ","); got != "first,a,b" {
		t.Errorf("pages = %s, want first,a,b", got)
	}
	if len(problems) != 2 {
		t.Errorf("problems = %v, want the two unusable files", problems)
	}
}

func TestLoadMissingDirectory(t *testing.T) {
	pages, problems := Load(filepath.Join(t.TempDir(), "nope"))
	if len(pages) != 0 || len(problems) != 0 {
		t.Errorf("pages = %v, problems = %v", pages, problems)
	}
}

func TestLoadSkipsAnOversizedFile(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "big.md", strings.Repeat("x", maxFileSize+1))
	pages, problems := Load(dir)
	if len(pages) != 0 || len(problems) != 1 {
		t.Errorf("pages = %v, problems = %v", pages, problems)
	}
}

func TestLoadFollowsSymlinks(t *testing.T) {
	// a mounted ConfigMap has this shape: name -> ..data/name
	dir := t.TempDir()
	data := filepath.Join(dir, "..data")
	if err := os.Mkdir(data, 0o700); err != nil {
		t.Fatal(err)
	}
	write(t, data, "about.md", "---\ntitle: About\n---\nhi")
	if err := os.Symlink(filepath.Join("..data", "about.md"), filepath.Join(dir, "about.md")); err != nil {
		t.Skip("symlinks not supported")
	}
	pages, problems := Load(dir)
	if len(pages) != 1 || len(problems) != 0 || pages[0].Title != "About" {
		t.Errorf("pages = %v, problems = %v", pages, problems)
	}
}

func TestLoadReadsTheHTMLNextToAPage(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "cal.md", "---\ntitle: Cal\n---\ntext")
	write(t, dir, "cal.html", "\n<!doctype html><p>widget</p>\n")
	write(t, dir, "plain.md", "text")
	write(t, dir, "stray.html", "<p>no page</p>")

	pages, problems := Load(dir)
	got := map[string]string{}
	for _, p := range pages {
		got[p.Slug] = p.EmbedHTML
	}
	if len(pages) != 2 || got["cal"] != "<!doctype html><p>widget</p>" || got["plain"] != "" {
		t.Errorf("pages = %+v", pages)
	}
	if len(problems) != 1 || !strings.Contains(problems[0], "stray.html") {
		t.Errorf("problems = %v, want the html file without a page", problems)
	}
}

func TestLoadSkipsAnOversizedHTMLFile(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "cal.md", "text")
	write(t, dir, "cal.html", strings.Repeat("x", maxFileSize+1))

	pages, problems := Load(dir)
	if len(pages) != 1 || pages[0].EmbedHTML != "" || len(problems) != 1 {
		t.Errorf("pages = %+v, problems = %v", pages, problems)
	}
}
