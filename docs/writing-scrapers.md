# Writing scrapers

A scraper is any program that collects event listings and POSTs them to the bruttice backend. Bruttice ships no scrapers. It defines an HTTP API, and a scraper can be written in any language.

For Go there is a reference client and the payload types in [`packages/go-pkg/ingest`](../packages/go-pkg/ingest), and a complete runnable example in [`packages/go-pkg/examples/scraper`](../packages/go-pkg/examples/scraper/main.go). Read the example first, it is about 100 lines.

## How it fits together

```
scraper ──POST /internal/ingest──▶ backend ──▶ database
   ▲                                  │
   └───── POST <ingest-trigger-url> ──┘   (optional, admin "run ingestion" button)
```

- The backend never fetches anything itself. Scrapers push.
- `/internal/` is blocked at nginx. A scraper must reach the backend directly, from the same Docker network or host (`http://bruttice-backend:8080`).
- Every request carries `Authorization: Bearer <ingest-key>`, where the key is `ingest-key` in the backend's `config.yaml`. A wrong or missing key gets `401`.

## Before you start: server configuration

The server rejects events that name things it doesn't know. These come from `config.yaml`:

| Payload field | Must match |
|---|---|
| `Age.Name` | an entry in `age-groups` |
| `Region.Name` or `venue_info.default_region` | an entry in `regions` |
| `Tags[].Name` | an entry in `tags`. Unknown tags are **silently dropped** |

## The API

### Run lifecycle

A run groups the payloads of one scrape in the admin dashboard's ingestion log, and makes the server email errors once, at the end.

| Request | Response |
|---|---|
| `POST /internal/ingest/runs` | `201 {"id": 12}` |
| `POST /internal/ingest` with `"run_id": 12` | one per venue, see below |
| `POST /internal/ingest/runs/12/finish` | `204` |

Rules:

- Always finish a run, even after a failure. Finishing sends the error email: the root admin is emailed any errors the run recorded, and every Monday even if there were none, so a silent inbox means ingestion is broken. Email is skipped when the server has no `MAIL_SERVER`.
- Runs are optional. A payload without `run_id` is recorded, and emailed on, as a run of its own.

### `POST /internal/ingest`

One payload per venue (or per source, if a scraper covers several venues):

```jsonc
{
  "run_id": 12,                         // optional
  "venue_info": {
    "name": "The Venue Name",           // required, the venue's unique key
    "link": "https://venue.example.com",
    "is_auto_ingest": true,             // see "Auto ingest" below
    "require_approval": false,          // true: events wait in the moderation queue
    "default_region": "tulsa",          // region name, applied to the venue
    "skip_venue_save": false            // true: label the payload but don't create a venue
  },
  "events": [ /* see below */ ],
  "errors": []                          // optional, see "Reporting failures"
}
```

An event:

```jsonc
{
  "guid": "The Venue Name-1001",              // optional, see "Identity"
  "title": "Open Mic Night",
  "link": "https://venue.example.com/events/1001",
  "start_at": "2030-01-10T19:00:00-06:00",    // RFC 3339, with an offset
  "doors_at": null,                           // optional
  "cover_cost": "free",                       // free text: "$10", "$5-$15", "free"
  "venue": "The Venue Name",                  // optional, defaults to venue_info.name
  "region": "tulsa",                          // optional, defaults to venue_info.default_region
  "age": "all-ages",
  "tags": ["music"],
  "cancelled": false                          // optional, a new cancelled event is stored as deleted
}
```

`title`, `start_at` and `age` are required, and so is a venue (the event's own or the payload's). An event missing any of them is refused and counted as failed, with the reason in the response's `errors`, and the other events are still saved. The server can't guess these: an empty age or a zero time would otherwise resolve to whichever row the database finds first.

The server refuses a payload with a field it doesn't know, so a typo or an out-of-date format is a `400` that names the field instead of an event that quietly loses data. Go scrapers don't write any of this by hand: marshal an `ingest.IngestionOutput` and it comes out right.

Responses:

| Status | Meaning |
|---|---|
| `200 {"saved": N, "errors": [...]}` | `errors` is the payload's own `errors` followed by one string for each event that failed to save. The rest were still saved |
| `400` | body is not valid JSON, or has a field the server doesn't know. The message says which |
| `401` | missing or wrong key |
| `500 {"saved": 0, "errors": [...]}` | the venue could not be saved. Nothing was ingested |

## What the server does with a payload

1. Saves the venue (unless `skip_venue_save`).
2. Records the payload's `errors` against the run, marking the venue's entry as having errors. They do not stop the events from being saved.
3. Saves each event in its own savepoint, so one bad event doesn't lose the rest. A partial payload is safe to save: the guid makes a re-sent event an update, and nothing is removed because it is missing from a payload.
4. Deduplicates the venues it touched. Two events at the same venue on the same day are duplicates if one title contains the other or they start within an hour of each other, and the older one is removed. Dedup is scoped to the venues named in the payload, so a scraper can't remove events from venues it didn't post.

Details that matter:

- **Identity.** Posting the same event again updates it. The identity is the `guid`. If you leave it empty, the server sets it to a hash of title, start time and venue, so a changed title or time makes a *new* event and orphans the old one. If the source has a stable id, use it: in Go, `event.UpdateGuid(&id)` produces `"<Venue>-<id>"`.
- **Deleted events stay deleted.** If an admin deleted an event, re-ingesting its guid is ignored.
- **Skipped events.** Titles containing "private party" or "private event" are dropped.
- **Free events** (a zero or "free" cover cost) get the `free` tag automatically.
- **Region.** For a normal venue, the venue's `default_region` wins over the event's `region`. Only for the reserved venue `Unlisted (see link)` is the event's own region used.
- **Titles** are HTML-unescaped.

### Auto ingest and approval

These are independent:

- `is_auto_ingest: true` marks the venue as scraped. The server **rejects** an event whose venue's scraped flag differs from the payload's `is_auto_ingest`, so one venue can't be fed by both scrapers and humans. A scraper that posts events for venues people manage, such as a shared calendar, sends `false`.
- Events are published on arrival. Set `require_approval: true` on the payload to send its events to the admin moderation queue instead.

### Reporting failures

Don't crash, and don't exit quietly. Put the failure in `errors` and send the payload anyway:

```go
out.AppendError(fmt.Errorf("fetching calendar: %w", err))
```

The server logs it, records it under the run, and emails it when the run finishes. In Go, `AppendError` prefixes the venue name. A scraper that dies without sending anything shows up as nothing, which is the one failure nobody is told about.

`errors` and `events` are independent. A scraper that parsed 40 of 41 rows sends the 40 events and an error for the 41st, and the 40 are saved. A scraper that failed outright sends no events and one error.

Partial results are safe to send because ingestion is idempotent. Events are matched by guid, so a re-sent event updates in place, and deduplication is scoped to the venue. A partial run never removes anything. The next complete run fills in what was missing.

## Scraper checklist

- **Timezones.** Venues publish wall-clock times. Parse them in the venue's zone (`time.ParseInLocation`), not UTC, and send an RFC 3339 time with an offset. This is the most common mistake. `venue_info.timezone` is not used by the server.
- **Stable ids** for the `guid`, see above.
- **Keep past events out of your output** unless you want them kept. The server doesn't delete events that vanish from a feed.
- **Timeouts.** Give every fetch a timeout. One hung venue shouldn't stall the run.
- **Be polite.** Set a User-Agent, don't hammer sites, and prefer a public API or calendar feed over scraping HTML.
- **Don't send secrets** in event fields.

## The Go client

```go
client := ingest.NewClient("http://bruttice-backend:8080", key)

runID, err := client.StartRun(ctx)
result, err := client.Send(ctx, runID, output) // result.Saved, result.Errors
err = client.FinishRun(ctx, runID)
```

The client uses only the standard library. Its `HTTP` field can be replaced to change the timeout (default two minutes) or transport.

## Other languages

The same flow with curl:

```sh
H='Authorization: Bearer '$INGEST_KEY
RUN=$(curl -s -X POST -H "$H" $URL/internal/ingest/runs | jq .id)

curl -s -X POST -H "$H" -H 'Content-Type: application/json' $URL/internal/ingest -d '{
  "run_id": '$RUN',
  "venue_info": {"name": "Example Venue", "is_auto_ingest": true, "default_region": "tulsa"},
  "events": [{
    "title": "Open Mic Night", "link": "https://venue.example.com/events/1001",
    "start_at": "2030-01-10T19:00:00-06:00", "cover_cost": "free",
    "age": "all-ages", "guid": "Example Venue-1001"
  }],
  "errors": []
}'

curl -s -X POST -H "$H" $URL/internal/ingest/runs/$RUN/finish
```

Any language that can POST JSON works the same way.

## Triggering a scraper from the dashboard

If `ingest-trigger-url` is set in `config.yaml`, the admin dashboard's "run ingestion" button sends `POST <ingest-trigger-url>` with `Authorization: Bearer <ingest-key>`. Your service answers:

- `202` once it has started a run (don't wait for it to finish),
- `409` if a run is already in progress.

Anything else is shown to the admin as a failure. Without the setting, the button reports that no scraper service is configured. This is optional: a scraper run from cron never needs it.

## Running a scraper

Run it however suits you, as long as it can reach the backend and has the key. In a compose stack, put it on the backend's network:

```yaml
services:
  my-scraper:
    image: my-scraper:latest
    environment:
      INGEST_URL: "http://bruttice-backend:8080"
      INGEST_KEY: "${INGEST_KEY}"
    networks: [bruttice]
    depends_on: [bruttice-backend]
```

A real-world layout (registry of scrapers, concurrent runner with panic recovery, daily schedule, trigger endpoint) is the private `dit-tulsa-scrapers` service described in [`open-source-split.md`](../open-source-split.md).

## Shared helpers

Google Calendar, Google Sheets and similar sources come up often, but bruttice does not provide helpers for them. They would pull third-party SDKs into the core and tie its releases to those SDKs' breaking changes. Write them in your scraper repo. If the community wants a shared library, it can live in its own module that depends on `go-pkg/ingest`.
