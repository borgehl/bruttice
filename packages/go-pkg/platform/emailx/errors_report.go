package emailx

import (
	"fmt"
	"html"
	"strings"
	"time"
)

const errDateFmt = "Mon Jan _2"

// SourceReport is what one scraper source did in a run.
type SourceReport struct {
	Name    string
	Scraped int
	Saved   int
	Failed  int
	// Errors are the messages of the source. The first line of a message says what went wrong, the
	// lines after it that start with spaces are "key: value" details about the event, see
	// SplitErrorDetails.
	Errors []string
}

// ErrorsReport is an ingest run, for the email to the admins about its errors.
type ErrorsReport struct {
	RunID   int64
	Sources []SourceReport
	// DedupeError is why removing duplicates failed, empty if it did not.
	DedupeError string
}

// Total is the number of errors in all sources.
func (r ErrorsReport) Total() int {
	n := 0
	for _, s := range r.Sources {
		n += len(s.Errors)
	}
	return n
}

// ErrorDetail is one line of "key: value" under an error.
type ErrorDetail struct{ Key, Value string }

// SplitErrorDetails separates an error message into its first line and the detail lines under it.
// A detail line starts with two spaces and has a ": ". Any other line stays with the headline.
func SplitErrorDetails(message string) (headline string, details []ErrorDetail) {
	lines := strings.Split(strings.TrimRight(message, "\n"), "\n")
	headlines := []string{lines[0]}
	for _, line := range lines[1:] {
		key, value, ok := strings.Cut(strings.TrimPrefix(line, "  "), ": ")
		if strings.HasPrefix(line, "  ") && ok && key != "" {
			details = append(details, ErrorDetail{Key: strings.TrimSpace(key), Value: value})
			continue
		}
		headlines = append(headlines, line)
	}
	return strings.Join(headlines, "\n"), details
}

// headline of an error without the source's name in front of it, which the email shows once above.
func (s SourceReport) cleanHeadline(message string) (string, []ErrorDetail) {
	headline, details := SplitErrorDetails(message)
	return strings.TrimPrefix(headline, s.Name+": "), details
}

func plural(n int, noun string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, noun)
	}
	return fmt.Sprintf("%d %ss", n, noun)
}

func (r ErrorsReport) summary() string {
	withErrors, scraped, saved := 0, 0, 0
	for _, s := range r.Sources {
		if len(s.Errors) > 0 {
			withErrors++
		}
		scraped += s.Scraped
		saved += s.Saved
	}
	run := ""
	if r.RunID != 0 {
		run = fmt.Sprintf(" (run %d)", r.RunID)
	}
	if r.Total() == 0 && r.DedupeError == "" {
		return fmt.Sprintf("No errors in this run%s. %s checked, %d events scraped, %d saved.", run, plural(len(r.Sources), "source"), scraped, saved)
	}
	return fmt.Sprintf("%s in %s%s. %d events scraped, %d saved.", plural(r.Total(), "error"), plural(withErrors, "source"), run, scraped, saved)
}

// ErrorsMessage tells the admins about the errors of an ingest run (the caller sets To), grouped by source, with the
// details of each event that failed.
func (e *EmailClient) ErrorsMessage(report ErrorsReport) Message {
	return Message{
		Subject: fmt.Sprintf("%s Scraping Errors: %d", time.Now().Format(errDateFmt), report.Total()),
		Text:    report.text(e.baseHost() + "/admin"),
		HTML:    report.html(e.baseHost() + "/admin"),
	}
}

func (r ErrorsReport) text(adminURL string) string {
	var b strings.Builder
	b.WriteString(r.summary() + "\n")
	if r.DedupeError != "" {
		b.WriteString("\nRemoving duplicates failed: " + r.DedupeError + "\n")
	}
	for _, s := range r.Sources {
		if len(s.Errors) == 0 {
			continue
		}
		fmt.Fprintf(&b, "\n== %s: %s (%d scraped, %d saved, %d failed)\n", s.Name, plural(len(s.Errors), "error"), s.Scraped, s.Saved, s.Failed)
		for i, message := range s.Errors {
			headline, details := s.cleanHeadline(message)
			fmt.Fprintf(&b, "%d. %s\n", i+1, strings.ReplaceAll(headline, "\n", "\n   "))
			for _, d := range details {
				fmt.Fprintf(&b, "     %s: %s\n", d.Key, d.Value)
			}
		}
	}
	b.WriteString("\nThe ingestion log: " + adminURL + "\n")
	return b.String()
}

// The colors are written out because mail clients ignore style sheets and many ignore custom
// properties, and each is readable on a white card, which some clients put on dark.
const (
	htmlFont  = `font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Helvetica,Arial,sans-serif;`
	htmlMono  = `font-family:SFMono-Regular,Consolas,'Liberation Mono',Menlo,monospace;`
	colorText = "#1f2430"
	colorSoft = "#5b6472"
	colorLine = "#e3e6ec"
	colorBad  = "#b42318"
	colorOk   = "#067647"
)

func esc(s string) string { return html.EscapeString(s) }

// detailValue shows a value, as a link if it is a web address.
func detailValue(value string) string {
	if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") {
		return fmt.Sprintf(`<a href="%s" style="color:#175cd3;word-break:break-all">%s</a>`, esc(value), esc(value))
	}
	return esc(value)
}

func (r ErrorsReport) html(adminURL string) string {
	var b strings.Builder
	b.WriteString(`<div style="background:#f4f5f7;padding:16px 8px">`)
	fmt.Fprintf(&b, `<div style="%smax-width:640px;margin:0 auto;background:#ffffff;border:1px solid %s;border-radius:8px;padding:20px;color:%s;line-height:1.45">`, htmlFont, colorLine, colorText)

	title, titleColor := "Scraping errors", colorBad
	if r.Total() == 0 && r.DedupeError == "" {
		title, titleColor = "Scraping check", colorOk
	}
	fmt.Fprintf(&b, `<h1 style="margin:0 0 4px;font-size:20px;color:%s">%s</h1>`, titleColor, title)
	fmt.Fprintf(&b, `<p style="margin:0 0 16px;color:%s;font-size:14px">%s</p>`, colorSoft, esc(r.summary()))

	if r.DedupeError != "" {
		fmt.Fprintf(&b, `<p style="margin:0 0 16px;padding:10px 12px;border-left:4px solid %s;background:#fef3f2;font-size:14px"><strong>Removing duplicates failed.</strong> %s</p>`, colorBad, esc(r.DedupeError))
	}

	for _, s := range r.Sources {
		if len(s.Errors) == 0 {
			continue
		}
		fmt.Fprintf(&b, `<h2 style="margin:20px 0 2px;font-size:16px">%s <span style="font-size:12px;font-weight:600;color:%s;background:#fef3f2;border-radius:10px;padding:2px 8px;white-space:nowrap">%s</span></h2>`, esc(s.Name), colorBad, plural(len(s.Errors), "error"))
		fmt.Fprintf(&b, `<p style="margin:0 0 8px;color:%s;font-size:13px">%d scraped &middot; %d saved &middot; %d failed</p>`, colorSoft, s.Scraped, s.Saved, s.Failed)
		for _, message := range s.Errors {
			headline, details := s.cleanHeadline(message)
			fmt.Fprintf(&b, `<div style="margin:0 0 10px;padding:10px 12px;border-left:4px solid %s;background:#fafafb;border-radius:0 6px 6px 0">`, colorBad)
			fmt.Fprintf(&b, `<div style="font-size:14px;font-weight:600;white-space:pre-wrap">%s</div>`, esc(headline))
			if len(details) > 0 {
				b.WriteString(`<table role="presentation" cellspacing="0" cellpadding="0" style="margin-top:6px;font-size:13px;border-collapse:collapse">`)
				for _, d := range details {
					fmt.Fprintf(&b, `<tr><td style="padding:1px 12px 1px 0;color:%s;vertical-align:top;white-space:nowrap">%s</td><td style="padding:1px 0;%sword-break:break-word">%s</td></tr>`, colorSoft, esc(d.Key), htmlMono, detailValue(d.Value))
				}
				b.WriteString(`</table>`)
			}
			b.WriteString(`</div>`)
		}
	}

	fmt.Fprintf(&b, `<p style="margin:20px 0 0;font-size:13px"><a href="%s" style="color:#175cd3">Open the ingestion log</a></p>`, esc(adminURL))
	b.WriteString(`</div></div>`)
	return b.String()
}
