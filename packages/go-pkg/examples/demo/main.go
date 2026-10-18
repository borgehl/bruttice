// Command demo loads example venues and events through the ingest API, for local
// development. It is what `make up-demo` runs. Dates are relative to today, and
// the guids are stable, so running it again moves the same events forward.
//
//	INGEST_URL=http://localhost:8080 INGEST_KEY=... [DEMO_REGION=central] go run ./examples/demo
package main

import (
	"context"
	"fmt"
	"os"
	"time"
	_ "time/tzdata"

	"github.com/borgehl/bruttice/packages/go-pkg/ingest"
	"github.com/borgehl/bruttice/packages/go-pkg/models"
)

const zone = "America/Chicago"

// region is the default region of every demo venue. It must be in the server's
// config, set DEMO_REGION if it has no "central".
var region = "central"

type listing struct {
	id        string
	title     string
	daysAhead int
	hour      int
	cost      string
	age       string
	tags      []string
}

type venue struct {
	name string
	link string
	// pending venues are not scraped, their events wait in the moderation queue
	pending  bool
	listings []listing
}

var venues = []venue{
	{
		name: "The Example Venue",
		link: "https://example.com/venue-one",
		listings: []listing{
			{"1", "Open Mic Night", 1, 19, "free", "all-ages", []string{"music", "community"}},
			{"2", "The Examples", 3, 20, "$10", "18+", []string{"music"}},
			{"3", "Stand-Up Showcase", 5, 20, "$8", "21+", []string{"comedy"}},
			{"4", "Sunday Jazz Brunch", 9, 11, "$5", "all-ages", []string{"music", "recurring"}},
			{"5", "Battle of the Bands", 14, 19, "$12", "16+", []string{"music"}},
			{"6", "Winter Songwriter Series", 30, 19, "$15", "21+", []string{"music"}},
		},
	},
	{
		name: "Community Arts Space",
		link: "https://example.com/venue-two",
		listings: []listing{
			{"1", "Zine Making Workshop", 2, 18, "free", "all-ages", []string{"diy", "community"}},
			{"2", "Poetry Slam", 4, 19, "$5", "16+", []string{"literature"}},
			{"3", "Community Theater: Our Town", 7, 19, "$10", "all-ages", []string{"theater"}},
			{"4", "Bike Repair Night", 10, 18, "free", "all-ages", []string{"diy", "recurring"}},
			{"5", "Book Swap", 16, 14, "free", "all-ages", []string{"literature", "community"}},
		},
	},
	{
		name: "The Back Room",
		link: "https://example.com/venue-three",
		listings: []listing{
			{"1", "Late Night Improv", 2, 22, "$7", "21+", []string{"comedy"}},
			{"2", "Local Band Showcase", 6, 21, "$10", "21+", []string{"music"}},
			{"3", "Karaoke Tuesday", 8, 21, "free", "21+", []string{"music", "recurring"}},
		},
	},
	{
		name:    "Pending Venue",
		link:    "https://example.com/venue-four",
		pending: true,
		listings: []listing{
			{"1", "Needs Approval: Pop-Up Market", 12, 12, "free", "all-ages", []string{"community"}},
			{"2", "Needs Approval: Open Studio", 13, 17, "$3", "all-ages", []string{"diy"}},
		},
	},
}

func main() {
	url, key := os.Getenv("INGEST_URL"), os.Getenv("INGEST_KEY")
	if url == "" || key == "" {
		fmt.Fprintln(os.Stderr, "INGEST_URL and INGEST_KEY are required")
		os.Exit(1)
	}
	if r := os.Getenv("DEMO_REGION"); r != "" {
		region = r
	}
	loc, err := time.LoadLocation(zone)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ctx := context.Background()
	client := ingest.NewClient(url, key)

	runID, err := startRun(ctx, client)
	if err != nil {
		fmt.Fprintln(os.Stderr, "start run:", err)
		os.Exit(1)
	}

	failed := false
	for _, v := range venues {
		out := payload(v, loc, time.Now())
		result, err := client.Send(ctx, runID, out)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %s\n", v.name, err)
			failed = true
			continue
		}
		fmt.Printf("%s: saved %d of %d events\n", v.name, result.Saved, len(out.Events))
		if len(result.Errors) > 0 {
			fmt.Fprintf(os.Stderr, "%s: %v\n", v.name, result.Errors)
			failed = true
		}
	}

	if err := client.FinishRun(ctx, runID); err != nil {
		fmt.Fprintln(os.Stderr, "finish run:", err)
		failed = true
	}
	if failed {
		os.Exit(1)
	}
}

// startRun waits for the backend to come up, which can take a while on a first start.
func startRun(ctx context.Context, client *ingest.Client) (int64, error) {
	var err error
	for range 60 {
		var id int64
		if id, err = client.StartRun(ctx); err == nil {
			return id, nil
		}
		time.Sleep(2 * time.Second)
	}
	return 0, err
}

func payload(v venue, loc *time.Location, now time.Time) ingest.IngestionOutput {
	out := ingest.IngestionOutput{
		VenueInfo: ingest.VenueInfo{
			Name:          v.name,
			Link:          v.link,
			IsAutoIngest:  !v.pending,
			DefaultRegion: region,
		},
	}

	verified := now.Unix()
	for _, l := range v.listings {
		day := now.In(loc).AddDate(0, 0, l.daysAhead)
		event := models.Event{
			Title:      l.title,
			Link:       fmt.Sprintf("%s/events/%s", v.link, l.id),
			StartAt:    time.Date(day.Year(), day.Month(), day.Day(), l.hour, 0, 0, 0, loc),
			CoverCost:  l.cost,
			Venue:      models.Venue{Name: v.name},
			Region:     models.Region{Name: region},
			Age:        models.Age{Name: l.age},
			VerifiedAt: &verified,
		}
		if !v.pending {
			event.ApprovedAt = &verified
		}
		for _, tag := range l.tags {
			event.AddTagStr(tag)
		}
		event.UpdateGuid(&l.id)
		out.Events = append(out.Events, event)
	}
	return out
}
