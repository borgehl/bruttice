package emailx

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/borgehl/bruttice/packages/go-pkg/platform/logx"
)

func sampleReport() ErrorsReport {
	return ErrorsReport{
		RunID: 12,
		Sources: []SourceReport{
			{Name: "Cains", Scraped: 10, Saved: 10},
			{
				Name: "Bricktown Comedy Club", Scraped: 40, Saved: 38, Failed: 2,
				Errors: []string{
					"event 28 (\"Closed for Thanksgiving\"): missing start_at\n  link: https://venue.example/events/28\n  venue: Bricktown Comedy Club (the payload's)\n  start_at: (none)\n  age: all-ages",
					"Bricktown Comedy Club: fetching the calendar: timeout",
				},
			},
			{
				Name: "The Vanguard", Scraped: 20, Saved: 19, Failed: 1,
				Errors: []string{"event 13 (\"NUNS\"): missing age\n  age: (none)"},
			},
		},
	}
}

func TestSplitErrorDetails(t *testing.T) {
	headline, details := SplitErrorDetails("event 1 (\"x\"): bad\n  link: https://a.example/?q=1: 2\n  age: (none)\nnot a detail\n  odd line")
	if headline != "event 1 (\"x\"): bad\nnot a detail\n  odd line" {
		t.Errorf("headline = %q", headline)
	}
	want := []ErrorDetail{{"link", "https://a.example/?q=1: 2"}, {"age", "(none)"}}
	if !reflect.DeepEqual(details, want) {
		t.Errorf("details = %+v", details)
	}
	if h, d := SplitErrorDetails("just a failure"); h != "just a failure" || len(d) != 0 {
		t.Errorf("plain message = %q, %v", h, d)
	}
}

func TestErrorsMessageIsGroupedBySource(t *testing.T) {
	c := NewEmailClient(logx.NewLogger(nil, nil), Config{Host: "https://events.example.com"})
	m := c.ErrorsMessage(sampleReport())

	if len(m.To) != 0 || !strings.HasSuffix(m.Subject, "Scraping Errors: 3") {
		t.Errorf("to = %v, subject = %q", m.To, m.Subject)
	}
	for _, want := range []string{
		"3 errors in 2 sources (run 12). 70 events scraped, 67 saved.",
		"== Bricktown Comedy Club: 2 errors (40 scraped, 38 saved, 2 failed)",
		"1. event 28 (\"Closed for Thanksgiving\"): missing start_at",
		"     link: https://venue.example/events/28",
		"     start_at: (none)",
		"2. fetching the calendar: timeout", // the source's name is shown once, above
		"== The Vanguard: 1 error",
		"https://events.example.com/admin",
	} {
		if !strings.Contains(m.Text, want) {
			t.Errorf("text is missing %q:\n%s", want, m.Text)
		}
	}
	if strings.Contains(m.Text, "== Cains") {
		t.Errorf("a source without errors is not listed:\n%s", m.Text)
	}
}

func TestErrorsMessageHTML(t *testing.T) {
	c := NewEmailClient(logx.NewLogger(nil, nil), Config{Host: "https://events.example.com"})
	m := c.ErrorsMessage(sampleReport())
	for _, want := range []string{
		"Bricktown Comedy Club",
		"2 errors</span>",
		"3 errors in 2 sources (run 12)",
		"event 28 (&#34;Closed for Thanksgiving&#34;): missing start_at",
		`<a href="https://venue.example/events/28"`,
		">start_at</td>",
		"(none)",
		`<a href="https://events.example.com/admin"`,
	} {
		if !strings.Contains(m.HTML, want) {
			t.Errorf("html is missing %q:\n%s", want, m.HTML)
		}
	}
	if strings.Contains(m.HTML, "Cains") {
		t.Error("a source without errors is not listed")
	}
}

func TestErrorsMessageHTMLEscapesWhatScrapersSend(t *testing.T) {
	c := NewEmailClient(logx.NewLogger(nil, nil), Config{})
	m := c.ErrorsMessage(ErrorsReport{Sources: []SourceReport{{
		Name:   "<b>Venue</b>",
		Errors: []string{"event 1 (\"<script>alert(1)</script>\"): bad\n  link: https://x.example/\"onmouseover=\"alert(1)\n  note: <img src=x onerror=alert(1)>\n  other: javascript:alert(1)"},
	}}})
	for _, bad := range []string{"<script>", "<b>Venue", "<img", `href="javascript`, `"onmouseover="`} {
		if strings.Contains(m.HTML, bad) {
			t.Errorf("html contains %q:\n%s", bad, m.HTML)
		}
	}
	if !strings.Contains(m.HTML, "&lt;script&gt;") {
		t.Errorf("scraper text must be shown as text:\n%s", m.HTML)
	}
}

func TestErrorsMessageWithoutErrors(t *testing.T) {
	c := NewEmailClient(logx.NewLogger(nil, nil), Config{})
	m := c.ErrorsMessage(ErrorsReport{RunID: 3, Sources: []SourceReport{{Name: "Cains", Scraped: 5, Saved: 5}}})
	if !strings.HasSuffix(m.Subject, "Scraping Errors: 0") || !strings.Contains(m.Text, "No errors in this run (run 3). 1 source checked") || !strings.Contains(m.HTML, "Scraping check") {
		t.Errorf("message = %+v", m)
	}
}

func TestErrorsMessageShowsAFailedDedupe(t *testing.T) {
	c := NewEmailClient(logx.NewLogger(nil, nil), Config{})
	m := c.ErrorsMessage(ErrorsReport{DedupeError: "deadlock detected"})
	if !strings.Contains(m.Text, "Removing duplicates failed: deadlock detected") || !strings.Contains(m.HTML, "Removing duplicates failed.") {
		t.Errorf("message = %+v", m)
	}
}

func TestSendIncludesTheHTMLPart(t *testing.T) {
	var got sendRequest
	c := testClient(t, Config{APIKey: "k", From: "hello@example.com"}, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&got)
		_, _ = w.Write([]byte(`{"id":"abc"}`))
	})
	if _, err := c.Send(Message{To: []string{"a@example.com"}, Subject: "s", Text: "plain", HTML: "<p>plain</p>"}); err != nil {
		t.Fatal(err)
	}
	if got.Text != "plain" || got.HTML != "<p>plain</p>" {
		t.Errorf("request = %+v", got)
	}
}
