// Package pagesx loads a deployment's own pages: Markdown files in a directory, each with an
// optional YAML header. A page's slug is its file name, so about.md is the page "about".
//
//	---
//	title: Pop! Booker
//	nav: true
//	order: 60
//	embed: https://app.example.com/calendar
//	---
//	The body, in Markdown.
//
// A page can also have an HTML file with the same name, such as pop-booker.html next to
// pop-booker.md. The frontend shows it in a sandboxed frame under the body, see Page.EmbedHTML.
package pagesx

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// maxFileSize keeps one oversized file from being read into memory and served.
const maxFileSize = 256 << 10

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

type Page struct {
	Slug  string
	Title string
	// Nav lists the page in the site's navigation. Pages that replace a built-in page
	// (about, privacy-policy) are never listed twice, the frontend knows which those are.
	Nav   bool
	Order int
	// Embed is an http(s) page shown under the body in a frame, empty if there is none.
	Embed string
	// EmbedHTML is a complete HTML document from the page's .html file, empty if there is none.
	// It can hold scripts, so the frontend only ever shows it in a frame with an opaque origin.
	EmbedHTML string
	Body      string
}

type header struct {
	Title string `yaml:"title"`
	Nav   bool   `yaml:"nav"`
	Order int    `yaml:"order"`
	Embed string `yaml:"embed"`
}

// Load reads every .md file in dir, and the .html file next to each, if any. A missing directory is not an error, a deployment may
// have no pages. Files that can't be used are skipped and described in problems, so one bad
// file never takes the others down. Pages come back ordered by Order, then Title.
func Load(dir string) (pages []Page, problems []string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if !os.IsNotExist(err) {
			problems = append(problems, fmt.Sprintf("read %s: %v", dir, err))
		}
		return nil, problems
	}

	for _, entry := range entries {
		name := entry.Name()
		// a ConfigMap mounted as a directory holds symlinks, and hidden entries like ..data
		if strings.HasPrefix(name, ".") {
			continue
		}
		switch filepath.Ext(name) {
		case ".html":
			// belongs to the page of the same name, read below
			if _, err := os.Stat(filepath.Join(dir, strings.TrimSuffix(name, ".html")+".md")); err != nil {
				problems = append(problems, fmt.Sprintf("%s: no %s.md next to it, skipped", name, strings.TrimSuffix(name, ".html")))
			}
			continue
		case ".md":
		default:
			continue
		}
		data, ok := readFile(dir, name, &problems)
		if !ok {
			continue
		}
		slug := strings.TrimSuffix(name, ".md")
		page, err := Parse(slug, string(data))
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s: %v", name, err))
			continue
		}
		if _, err := os.Stat(filepath.Join(dir, slug+".html")); err == nil {
			if html, ok := readFile(dir, slug+".html", &problems); ok {
				page.EmbedHTML = strings.TrimSpace(string(html))
			}
		}
		pages = append(pages, page)
	}

	sort.SliceStable(pages, func(i, j int) bool {
		if pages[i].Order != pages[j].Order {
			return pages[i].Order < pages[j].Order
		}
		return pages[i].Title < pages[j].Title
	})
	return pages, problems
}

// readFile reads one file, unless it is a directory or too large. Why it could not be used goes
// to problems, except for a directory, which is ignored.
func readFile(dir, name string, problems *[]string) ([]byte, bool) {
	path := filepath.Join(dir, name)
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return nil, false
	}
	if info.Size() > maxFileSize {
		*problems = append(*problems, fmt.Sprintf("%s: larger than %d bytes, skipped", name, maxFileSize))
		return nil, false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		*problems = append(*problems, fmt.Sprintf("%s: %v", name, err))
		return nil, false
	}
	return data, true
}

// Parse builds a page from a slug and the text of its file.
func Parse(slug, text string) (Page, error) {
	if !slugPattern.MatchString(slug) {
		return Page{}, fmt.Errorf("file name %q must be lower case letters, digits and single dashes", slug)
	}

	var h header
	body := text
	if rest, ok := strings.CutPrefix(strings.ReplaceAll(text, "\r\n", "\n"), "---\n"); ok {
		raw, after, found := strings.Cut("\n"+rest, "\n---")
		if !found {
			return Page{}, fmt.Errorf("the header has no closing ---")
		}
		// the closing line is "---" alone, the rest of its line must be empty
		after, ok := strings.CutPrefix(after, "\n")
		if !ok && after != "" {
			return Page{}, fmt.Errorf("the header has no closing ---")
		}
		if err := yaml.Unmarshal([]byte(raw), &h); err != nil {
			return Page{}, fmt.Errorf("header: %v", err)
		}
		body = after
	}

	page := Page{
		Slug:  slug,
		Title: strings.TrimSpace(h.Title),
		Nav:   h.Nav,
		Order: h.Order,
		Embed: httpURL(h.Embed),
		Body:  strings.TrimSpace(body),
	}
	if page.Title == "" {
		page.Title = titleFromSlug(slug)
	}
	return page, nil
}

// httpURL keeps a link only if it is http(s), a frame is never pointed at another scheme.
func httpURL(url string) string {
	url = strings.TrimSpace(url)
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return url
	}
	return ""
}

func titleFromSlug(slug string) string {
	words := strings.Split(slug, "-")
	for i, w := range words {
		words[i] = strings.ToUpper(w[:1]) + w[1:]
	}
	return strings.Join(words, " ")
}
