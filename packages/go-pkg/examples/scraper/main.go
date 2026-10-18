// Command scraper is a minimal example of a bruttice scraper.
//
// It "scrapes" a hardcoded list so it runs anywhere. Replace scrape() with a
// real fetch and parse, and keep the rest. See docs/writing-scrapers.md.
//
//	INGEST_URL=http://localhost:8080 INGEST_KEY=... go run ./examples/scraper
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/borgehl/bruttice/packages/go-pkg/ingest"
)

const (
	venueName = "Example Venue"
	venueLink = "https://venue.example.com"

	// must match a region name in the server's config.yaml
	region = "tulsa"
	// must match an age group name in the server's config.yaml
	age = "all-ages"
	// the timezone the venue publishes its times in
	venueTimezone = "America/Chicago"
)

func main() {
	url, key := os.Getenv("INGEST_URL"), os.Getenv("INGEST_KEY")
	if url == "" || key == "" {
		fmt.Fprintln(os.Stderr, "INGEST_URL and INGEST_KEY are required")
		os.Exit(1)
	}

	ctx := context.Background()
	client := ingest.NewClient(url, key)

	// A run groups the payloads of one scrape in the admin ingestion log.
	// A scraper with a single payload can skip it, and pass 0 to Send.
	runID, err := client.StartRun(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "start run:", err)
		os.Exit(1)
	}

	out := scrape()
	result, err := client.Send(ctx, runID, out)
	if err != nil {
		fmt.Fprintln(os.Stderr, "send:", err)
	} else {
		fmt.Printf("scraped %d events, server saved %d, %d failed\n", len(out.Events), result.Saved, len(result.Errors))
	}

	// Finishing the run triggers the error email, so do it even after a failure.
	if err := client.FinishRun(ctx, runID); err != nil {
		fmt.Fprintln(os.Stderr, "finish run:", err)
		os.Exit(1)
	}
}

// scrape returns one venue's output. Failures go in the output's Errors rather
// than being returned or panicking, so the server records and emails them.
func scrape() ingest.IngestionOutput {
	out := ingest.IngestionOutput{
		VenueInfo: ingest.VenueInfo{
			Name:          venueName,
			Link:          venueLink,
			IsAutoIngest:  true,
			DefaultRegion: region,
		},
	}

	loc, err := time.LoadLocation(venueTimezone)
	if err != nil {
		out.AppendError(err)
		return out
	}

	// Stand-in for a fetch and parse. Times are written as the venue shows them.
	listings := []struct{ id, title, link, start, cost string }{
		{"1001", "Open Mic Night", venueLink + "/events/1001", "2030-01-10 7:00pm", "free"},
		{"1002", "The Examples", venueLink + "/events/1002", "2030-01-11 8:30pm", "$10"},
	}

	for _, l := range listings {
		// ParseInLocation reads the wall-clock time in the venue's zone.
		// time.Parse would read it as UTC and shift every event.
		start, err := time.ParseInLocation(ingest.DateFormat, l.start, loc)
		if err != nil {
			// leave the event out and report why; the others are still saved
			out.AppendError(fmt.Errorf("event %q: %w", l.title, err))
			continue
		}

		event := ingest.Event{
			Title:     l.title,
			Link:      l.link,
			StartAt:   start,
			CoverCost: l.cost,
			Venue:     venueName,
			Region:    region,
			Age:       age,
		}
		// A stable id from the venue's site makes the event keep its identity when
		// its title or time changes. Without one the guid hashes title, time and venue.
		event.UpdateGuid(&l.id)

		out.Events = append(out.Events, event)
	}
	return out
}
