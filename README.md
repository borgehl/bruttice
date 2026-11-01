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

Nginx exposes `/api/` and `/` to the public. The `/internal/` path — used by scrapers — is blocked at nginx and only reachable from within the Docker network.

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
| `RESEND_API_KEY` | [Resend](https://resend.com) API key for outbound email: event verification, admin notifications, run reports and the contact form. Leave blank to disable all of them. Needs `site.contact-email` in `config.yaml`. See [Email](#email) |
| `RESEND_WEBHOOK_SECRET` | Signing secret of the Resend webhook that delivers received email to the admin inbox. Leave blank to turn receiving off |
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
| `ingest-key` | Secret shared with scrapers posting events (see [docs/writing-scrapers.md](docs/writing-scrapers.md)) |
| `ingest-trigger-url` | Optional. Endpoint of your scraper service that starts a run, used by the admin dashboard's "run ingestion" button (see [Triggering a run](#triggering-a-run-from-the-admin-dashboard)) |
| `admin-accounts` | `email;password` pairs. Order does not matter. The `ADMIN_ACCOUNTS` environment variable (`email1;pw1,email2;pw2`) replaces this list when set, so passwords need not be in the file; passwords can't contain `,` or `;` there. An existing admin's password is not changed by config |
| `site` | What shows on the pages of your deployment: `name`, `region`, `contact-email` and `description`. All optional, see `GET /api/site`. `contact-email` is the address all mail is sent from and the contact form's messages go to. It is never shown to visitors, and has to be on a domain verified in Resend |
| `pages-dir` | Optional. Directory of your own pages, as Markdown files (default `/app/pages`). See [docs/custom-pages.md](docs/custom-pages.md) |
| `tags` | Tags seeded into the database on first run |
| `age-groups` | Age restriction tiers (`name` + numeric `value`; 1 = all-ages, 255 = unknown) |
| `regions` | Geographic regions available for events |
| `venues` | Venues seeded on first run — include `"Unlisted (see link)"` to enable that feature |

The seed data (`tags`, `age-groups`, `regions`, `venues`) is saved into the database on the first start only. After that the admin dashboard owns it (see [Admin dashboard](#admin-dashboard)), and editing these lists in the config has no effect: the backend logs a reminder at startup. This is so a region or tag an admin retired is not brought back by a restart. A database that was set up before this change gets one last save of the config on its next start.

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

## Development

| Command | What it does |
|---|---|
| `make up-demo` | Builds and starts the stack with example venues and events at http://localhost:8080. Admin: `admin@example.com` / `demo-password`. Uses `demo/env` and `demo/config.yaml`, so your own `.env` and `config.yaml` are not read |
| `make up` | Builds and starts the stack from your own `.env` and `apps/backend/config.yaml` |
| `make down-demo` | Stops the demo and keeps its data |
| `make clean` | Removes the containers and database volumes of the demo and of `make up`. Nothing else in Docker is touched |
| `make test-backend` | Go tests for `apps/backend` and `packages/go-pkg`. No database needed |
| `make test-frontend` | Typecheck and tests for `apps/frontend` |
| `make test` | Both |

The example data is loaded through the ingest API by `packages/go-pkg/examples/demo`, so it also shows what a scraper's payload looks like. It includes a venue whose events wait in the admin moderation queue. Container names are fixed, so stop one stack before starting the other.

## Writing scrapers

Bruttice has no scrapers built in. Anything that can POST JSON to `/internal/ingest` can feed it, in any language, and the admin dashboard's "run ingestion" button can start yours through `ingest-trigger-url`.

See **[docs/writing-scrapers.md](docs/writing-scrapers.md)** for the API, the payload format and what the server does with it. For Go, [`packages/go-pkg/ingest`](packages/go-pkg/ingest) has the payload types and a reference client, and [`packages/go-pkg/examples/scraper`](packages/go-pkg/examples/scraper/main.go) is a complete working scraper to copy. For Python, [`packages/py-ingest`](packages/py-ingest) is a client with no dependencies, and [`examples/scraper.py`](packages/py-ingest/examples/scraper.py) is the same example scraper.

## Email

Email goes through [Resend](https://resend.com). `site.contact-email` is the address the site sends from, and `RESEND_API_KEY` lets it. The site also receives mail, which is how the contact form and replies work:

- A visitor writes through the contact form. The message lands in the **Inbox** tab of the admin dashboard, and the admins who turned on the inbox email in their settings get an email saying so. The visitor's address is never shown anywhere else.
- An admin replies from the inbox, or starts a conversation with **New email**. The message is emailed from `contact-email`, with a `Reply-To` of `reply-<conversation id>@<your domain>`.
- When the visitor answers, Resend receives it, posts it to `POST /api/webhooks/resend`, and it joins the same conversation. Mail sent to `contact-email` itself starts a new conversation.

To set it up:

1. Pick a subdomain for the site's mail, such as `mail.example.com`, and use an address on it as `contact-email` (`hello@mail.example.com`). Receiving needs an MX record for the domain, and a subdomain keeps your main domain's mail where it is.
2. In Resend, add that domain, add the DNS records it asks for (SPF and DKIM for sending), and turn on **Receiving** for it (the MX record).
3. Add a webhook in Resend for the events `email.received` and `email.delivered` with the URL `https://<your host>/api/webhooks/resend`, and put its signing secret in `RESEND_WEBHOOK_SECRET`. The second shows a "Delivered" time on the emails admins send, in the inbox.

The webhook is verified with that secret, requests older than five minutes are refused, and without the secret the endpoint is not served. Only mail to `contact-email` and to the conversation reply addresses is saved, other addresses on the domain are ignored. Messages are kept and shown as plain text.

## Admin dashboard

Log in at `/admin` with the credentials from `admin-accounts` in `config.yaml`. From there you can approve or reject submitted events, edit event details, and manage the moderation queue. The **Inbox** tab has the messages from the contact form and from email, with replies (see [Email](#email)). The **Site lists** page (`/admin/config`) manages regions, tags, age groups and venues: add, rename and delete them, and retire a region. A retired region's events either move to another region, together with the venues that default to it and the old name (a scraper that still sends it lands in the new region), or they are deleted; at least one region has to stay, and a region can't be retired by deleting its events while a venue still defaults to it. Naming a retired region again restores it. Deleting a tag removes it from its events; deleting an age group moves its events to another one. Scraped venues are kept up to date by their scraper and can't be edited, and a venue with events can't be deleted. Each admin can set their own name at `/admin/settings`; it is shown to the other admins, for example on replies in the inbox, and falls back to the email. There each admin also chooses which emails they get: new events awaiting approval (an hourly note), system errors (the report of an ingest run's errors, and database errors) and new inbox messages. Nothing is sent to an admin who has not turned it on, and a mail nobody chose is not sent (a log line says so).

## Event tags and flags

Event descriptions can embed structured flags. The helpers in `packages/go-pkg/platform/util` (`ParseTagsFromStr`, `ParseLocationFromStr`, `ParseAge`) read them, so a scraper turns them into tags, a region and an age group:

| Flag syntax | Effect |
|---|---|
| `bruttice-tags:rock:live-music` | Attaches tags `rock` and `live-music` to the event |
| `bruttice-loc:downtown` | Sets the event's region |
| `bruttice-ages:18+` | Sets the age restriction |

These are useful when scraping venues whose event pages embed this metadata in the description field.

## Feeds

Authenticated RSS, Atom, and JSON feeds are available at `/api/feeds`. Feed metadata is controlled by `FEED_TITLE`, `FEED_DESCRIPTION`, and `FEED_AUTHOR` in `.env`.
