"""A minimal bruttice scraper in Python.

It "scrapes" a hardcoded list so it runs anywhere. Replace scrape() with a real fetch and parse,
and keep the rest. See docs/writing-scrapers.md.

    INGEST_URL=http://localhost:8080 INGEST_KEY=... python examples/scraper.py
"""

import os
import sys

from bruttice_ingest import Client, Event, IngestError, IngestionOutput, VenueInfo, parse_local

VENUE_NAME = "Example Venue"
VENUE_LINK = "https://venue.example.com"

# must match a region name in the server's config
REGION = "tulsa"
# must match an age group name in the server's config
AGE = "all-ages"
# the timezone the venue publishes its times in
VENUE_TIMEZONE = "America/Chicago"


def scrape() -> IngestionOutput:
    """Returns one venue's output. Failures go in the output's errors rather than being raised, so the
    server records and emails them."""
    out = IngestionOutput(
        VenueInfo(name=VENUE_NAME, link=VENUE_LINK, is_auto_ingest=True, default_region=REGION)
    )

    # Stand-in for a fetch and parse. Times are written as the venue shows them.
    listings = [
        ("1001", "Open Mic Night", VENUE_LINK + "/events/1001", "2030-01-10 07:00PM", "free"),
        ("1002", "The Examples", VENUE_LINK + "/events/1002", "2030-01-11 08:30PM", "$10"),
    ]

    for source_id, title, link, start, cost in listings:
        try:
            # parse_local reads the wall-clock time in the venue's zone. A time without a zone is
            # refused, and UTC would shift every event.
            start_at = parse_local(start, "%Y-%m-%d %I:%M%p", VENUE_TIMEZONE)
        except ValueError as err:
            # leave the event out and report why; the others are still saved
            out.append_error(f"event {title!r}: {err}")
            continue

        event = Event(
            title=title,
            link=link,
            start_at=start_at,
            cover_cost=cost,
            venue=VENUE_NAME,
            region=REGION,
            age=AGE,
        )
        # A stable id from the venue's site makes the event keep its identity when its title or
        # time changes. Without one the server hashes title, time and venue.
        event.update_guid(source_id)
        out.events.append(event)
    return out


def main() -> int:
    url, key = os.environ.get("INGEST_URL"), os.environ.get("INGEST_KEY")
    if not url or not key:
        print("INGEST_URL and INGEST_KEY are required", file=sys.stderr)
        return 1

    client = Client(url, key)
    try:
        # The run is always finished, which is what triggers the error email, even after a failure.
        with client.run() as run:
            out = scrape()
            result = run.send(out)
            print(f"scraped {len(out.events)} events, server saved {result.saved}, {len(result.errors)} failed")
    except IngestError as err:
        print(err, file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
