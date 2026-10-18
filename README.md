# Bruttice

An open-source event calendar with a moderation queue, RSS/Atom/JSON feeds, and an HTTP ingest API for posting events from custom scrapers or integrations.

## Overview

Bruttice runs four Docker services:

| Service | Role |
|---|---|
| `bruttice-backend` | Go HTTP API |
| `bruttice-frontend` | React SPA |
| `bruttice-db` | PostgreSQL |
| `bruttice-nginx` | Reverse proxy (public entry point) |

Nginx exposes `/api/` and `/` to the public. The `/internal/` path — used by ingestors — is blocked at nginx and only reachable from within the Docker network.

## Quick start

### 1. Pull images

Pre-built images are published to the GitHub Container Registry on every push to `main` and on each GitHub Release.

```
ghcr.io/borgehl/bruttice-backend:latest
ghcr.io/borgehl/bruttice-frontend:latest
```

Pinning to a release tag is recommended for production:

```
ghcr.io/borgehl/bruttice-backend:v1.2.3
ghcr.io/borgehl/bruttice-frontend:v1.2.3
```

### 2. Configure

Copy the env file and edit it:

```bash
cp .env.example .env
```

Key values to set:

| Variable | Description |
|---|---|
| `POSTGRES_PASSWORD` | Database password |
| `HOST` | Public base URL, e.g. `https://events.example.com` |
| `PORT` | Host port nginx binds on |
| `FEED_TITLE` / `FEED_DESCRIPTION` / `FEED_AUTHOR` | RSS/Atom/JSON feed metadata |
| `MAIL_SERVER` / `MAIL_EMAIL` / `MAIL_PASSWORD` / `MAIL_PORT` | SMTP for outbound email (leave blank to disable) |
| `IP_GEOLOCATION_IO_API_KEY` | Optional — enables geo lookup for analytics |

Create `apps/backend/config.yaml` from the example:

```bash
cp apps/backend/config.yaml.example apps/backend/config.yaml
```

Key fields:

| Key | Description |
|---|---|
| `timezone` | Server timezone (any tz database name, e.g. `America/Chicago`) |
| `cookie-salt` | Random secret used to sign session cookies |
| `ingest-key` | Secret shared with ingestors posting events |
| `admin-accounts` | `email;password` pairs — first entry is the root admin |
| `tags` | Tags seeded into the database on first run |
| `age-groups` | Age restriction tiers (`name` + numeric `value`; 1 = all-ages, 255 = unknown) |
| `regions` | Geographic regions available for events |
| `venues` | Venues seeded on first run — include `"Unlisted (see link)"` to enable that feature |

All seed data (`tags`, `age-groups`, `regions`, `venues`) is upserted on every startup, so adding new entries to the config takes effect on the next restart without manual database changes.

### 3. Start

Update `docker-compose.yaml` to pull from ghcr.io instead of building locally:

```yaml
services:
  bruttice-backend:
    image: ghcr.io/borgehl/bruttice-backend:latest
    pull_policy: always
    # remove the build: block
    ...

  bruttice-frontend:
    image: ghcr.io/borgehl/bruttice-frontend:latest
    pull_policy: always
    # remove the build: block
    ...
```

Then:

```bash
docker compose up -d
```

The site is available at `http://localhost:${PORT}`.

## Writing an ingestor

An ingestor is any process that scrapes or fetches event data and POSTs it to the backend. Because `/internal/ingest` is blocked at nginx, ingestors must run inside the same Docker network (`bruttice`) or on the same host.

### Endpoint

```
POST http://bruttice-backend:8080/internal/ingest
Authorization: Bearer <ingest-key>
Content-Type: application/json
```

The `ingest-key` must match the `ingest-key` value in `config.yaml`.

### Payload

```jsonc
{
  "venue_info": {
    "name": "The Venue Name",        // required — used as a unique key
    "link": "https://venue.example.com",
    "is_auto_ingest": true,          // if true, events skip the moderation queue
    "default_region": "downtown",    // optional region tag applied to all events
    "skip_venue_save": false,        // set true to omit creating a Venue record
    "timezone": "America/Chicago"    // timezone of the raw event times
  },
  "events": [
    {
      "Title": "Event Name",
      "Link": "https://tickets.example.com/event/123",
      "StartAt": "2025-09-15T20:00:00Z",   // RFC3339 UTC
      "DoorsAt": "2025-09-15T19:00:00Z",   // optional
      "CoverCost": "$10",                  // free-text, e.g. "$5–$15" or "free"
      "Venue": { "Name": "The Venue Name" },
      "Age":   { "Name": "18+" },
      "Tags":  [{ "Name": "rock" }, { "Name": "live-music" }]
    }
  ],
  "errors": []   // non-fatal scraper errors logged server-side
}
```

Events with `is_auto_ingest: true` are published immediately. Events from venues without this flag enter a moderation queue visible in the admin dashboard.

Duplicate detection runs after each ingest. An event is considered a duplicate if its GUID — derived from `<VenueName>-<title>-<start_time>` — already exists in the database.

### Example: minimal Go ingestor

The `packages/go-pkg/ingest` package provides the shared types so you don't have to write them by hand.

```go
import (
    "bytes"
    "encoding/json"
    "net/http"
    "time"

    "github.com/borgehl/bruttice/packages/go-pkg/ingest"
    "github.com/borgehl/bruttice/packages/go-pkg/models"
)

func submit(key string, output ingest.IngestionOutput) error {
    body, _ := json.Marshal(output)
    req, _ := http.NewRequest("POST", "http://bruttice-backend:8080/internal/ingest", bytes.NewReader(body))
    req.Header.Set("Authorization", "Bearer "+key)
    req.Header.Set("Content-Type", "application/json")
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    return nil
}

func main() {
    start, _ := time.Parse(time.RFC3339, "2025-09-15T20:00:00Z")

    output := ingest.IngestionOutput{
        VenueInfo: ingest.VenueInfo{
            Name:         "The Venue Name",
            Link:         "https://venue.example.com",
            IsAutoIngest: true,
        },
        Events: []models.Event{
            {
                Title:     "My Event",
                Link:      "https://tickets.example.com/123",
                StartAt:   start,
                CoverCost: "$10",
                Venue:     models.Venue{Name: "The Venue Name"},
                Age:       models.Age{Name: "18+"},
            },
        },
    }

    if err := submit("your-ingest-key", output); err != nil {
        panic(err)
    }
}
```

### Running an ingestor alongside the stack

Add your ingestor as a service in `docker-compose.yaml` on the same network:

```yaml
services:
  my-ingestor:
    image: my-ingestor:latest
    environment:
      INGEST_KEY: "your-ingest-key"
      BACKEND_URL: "http://bruttice-backend:8080"
    networks:
      - bruttice
    depends_on:
      - bruttice-backend

networks:
  bruttice:
    external: true   # if composing separately; omit if in the same file
```

## Admin dashboard

Log in at `/admin` with the credentials from `admin-accounts` in `config.yaml`. From there you can approve or reject submitted events, edit event details, and manage the moderation queue.

## Event tags and flags

Event descriptions can embed structured flags that the backend parses automatically:

| Flag syntax | Effect |
|---|---|
| `dit-tags:rock:live-music` | Attaches tags `rock` and `live-music` to the event |
| `dit-loc:downtown` | Sets the event's region |
| `dit-ages:18+` | Sets the age restriction |

These are useful when scraping venues whose event pages embed this metadata in the description field.

## Feeds

Authenticated RSS, Atom, and JSON feeds are available at `/api/feeds`. Feed metadata is controlled by `FEED_TITLE`, `FEED_DESCRIPTION`, and `FEED_AUTHOR` in `.env`.
