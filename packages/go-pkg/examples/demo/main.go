// Command demo loads example venues and events through the ingest API, for local
// development. It is what `make up-demo` runs. Dates are relative to today, and
// the guids are stable, so running it again moves the same events forward.
//
//	INGEST_URL=http://localhost:8080 INGEST_KEY=... [DEMO_REGION=central] go run ./examples/demo
//
// It also loads example visitors for the admin Visitors page, and one example sponsor card, when
// the POSTGRES_* variables of the backend are set, see visitors.go and sponsor.go.
package main

import (
	"context"
	"fmt"
	"os"
	"time"
	_ "time/tzdata"

	"github.com/borgehl/bruttice/packages/go-pkg/ingest"
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
	// cancelled listings show with a "Cancelled" mark
	cancelled bool
}

type venue struct {
	name string
	link string
	// region of the venue's events, the default region when empty
	region string
	// pending venues are not scraped, and their events wait in the moderation queue
	pending  bool
	listings []listing
}

var venues = []venue{
	{
		name: "The Example Venue",
		link: "https://example.com/venue-one",
		listings: []listing{
			{"1", "Open Mic Night", 1, 19, "free", "all-ages", []string{"music", "community"}, false},
			{"2", "The Examples", 3, 20, "$10", "18+", []string{"music"}, false},
			{"3", "Stand-Up Showcase", 5, 20, "$8", "21+", []string{"comedy"}, false},
			{"4", "Sunday Jazz Brunch", 9, 11, "$5", "all-ages", []string{"music", "recurring"}, false},
			{"5", "Battle of the Bands", 14, 19, "$12", "16+", []string{"music"}, false},
			{"6", "Winter Songwriter Series", 30, 19, "$15", "21+", []string{"music"}, false},
			{"7", "Acoustic Night", 4, 19, "$8", "all-ages", []string{"music"}, true},
		},
	},
	{
		name: "Community Arts Space",
		link: "https://example.com/venue-two",
		listings: []listing{
			{"1", "Zine Making Workshop", 2, 18, "free", "all-ages", []string{"diy", "community"}, false},
			{"2", "Poetry Slam", 4, 19, "$5", "16+", []string{"literature"}, false},
			{"3", "Community Theater: Our Town", 7, 19, "$10", "all-ages", []string{"theater"}, false},
			{"4", "Bike Repair Night", 10, 18, "free", "all-ages", []string{"diy", "recurring"}, false},
			{"5", "Book Swap", 16, 14, "free", "all-ages", []string{"literature", "community"}, false},
		},
	},
	{
		name: "The Back Room",
		link: "https://example.com/venue-three",
		listings: []listing{
			{"1", "Late Night Improv", 2, 22, "$7", "21+", []string{"comedy"}, false},
			{"2", "Local Band Showcase", 6, 21, "$10", "21+", []string{"music"}, false},
			{"3", "Karaoke Tuesday", 8, 21, "free", "21+", []string{"music", "recurring"}, false},
		},
	},
	// the only venue outside the default region, for trying the region picker and region filters
	{
		name:   "Harbor Hall",
		link:   "https://example.com/venue-five",
		region: "eastern",
		listings: []listing{
			{"1", "Sea Shanty Sing-Along", 2, 19, "free", "all-ages", []string{"music", "community"}, false},
			{"2", "Dockside Stand-Up", 5, 20, "$10", "21+", []string{"comedy"}, false},
			{"3", "Lighthouse Poetry Reading", 8, 18, "$5", "16+", []string{"literature"}, false},
			{"4", "Harbor Jazz Night", 11, 20, "$12", "18+", []string{"music"}, false},
		},
	},
	{
		name:    "Pending Venue",
		link:    "https://example.com/venue-four",
		pending: true,
		listings: []listing{
			{"1", "Needs Approval: Pop-Up Market", 12, 12, "free", "all-ages", []string{"community"}, false},
			{"2", "Needs Approval: Open Studio", 13, 17, "$3", "all-ages", []string{"diy"}, false},
			{"3", "Needs Approval: Cancelled Craft Fair", 14, 12, "free", "all-ages", []string{"community"}, true},
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

	// the backend is up, so its tables exist; without a database to reach, only events are loaded
	if os.Getenv("POSTGRES_HOST") != "" {
		logs, err := seedVisitors("localhost:8080", loc, time.Now())
		if err != nil {
			fmt.Fprintln(os.Stderr, "visitors:", err)
			failed = true
		} else {
			fmt.Printf("visitors: saved %d visits\n", logs)
		}

		if err := seedSponsor(loc, time.Now()); err != nil {
			fmt.Fprintln(os.Stderr, "sponsor:", err)
			failed = true
		} else {
			fmt.Printf("sponsor: saved %q\n", demoSponsorName)
		}
	}

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
	region := region
	if v.region != "" {
		region = v.region
	}

	out := ingest.IngestionOutput{
		VenueInfo: ingest.VenueInfo{
			Name:            v.name,
			Link:            v.link,
			IsAutoIngest:    !v.pending,
			RequireApproval: v.pending,
			DefaultRegion:   region,
		},
	}

	for _, l := range v.listings {
		day := now.In(loc).AddDate(0, 0, l.daysAhead)
		event := ingest.Event{
			Title:     l.title,
			Link:      fmt.Sprintf("%s/events/%s", v.link, l.id),
			StartAt:   time.Date(day.Year(), day.Month(), day.Day(), l.hour, 0, 0, 0, loc),
			CoverCost: l.cost,
			Venue:     v.name,
			Region:    region,
			Age:       l.age,
			Tags:      l.tags,
			Cancelled: l.cancelled,
		}
		event.UpdateGuid(&l.id)
		out.Events = append(out.Events, event)
	}
	return out
}
