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
	Body  string
}

type header struct {
	Title string `yaml:"title"`
	Nav   bool   `yaml:"nav"`
	Order int    `yaml:"order"`
	Embed string `yaml:"embed"`
}

// Load reads every .md file in dir. A missing directory is not an error, a deployment may
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
		if strings.HasPrefix(name, ".") || filepath.Ext(name) != ".md" {
			continue
		}
		path := filepath.Join(dir, name)
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			continue
		}
		if info.Size() > maxFileSize {
			problems = append(problems, fmt.Sprintf("%s: larger than %d bytes, skipped", name, maxFileSize))
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s: %v", name, err))
			continue
		}
		page, err := Parse(strings.TrimSuffix(name, ".md"), string(data))
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s: %v", name, err))
			continue
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
