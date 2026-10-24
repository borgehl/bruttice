# bruttice-ingest

A small client for the bruttice ingest API, so scrapers can be written in Python. It has no dependencies (Python 3.9 or later, standard library only) and mirrors the Go client in [`packages/go-pkg/ingest`](../go-pkg/ingest). The API itself is described in [docs/writing-scrapers.md](../../docs/writing-scrapers.md).

## Install

From a checkout of the repo:

```sh
pip install ./packages/py-ingest
```

or copy the `bruttice_ingest` folder next to your scraper.

## A scraper

```python
from bruttice_ingest import Client, Event, IngestionOutput, VenueInfo, parse_local

client = Client("http://bruttice-backend:8080", "<ingest-key>")

with client.run() as run:
    out = IngestionOutput(VenueInfo(name="The Venue", link="https://venue.example", default_region="tulsa"))

    out.events.append(Event(
        title="Open Mic Night",
        link="https://venue.example/events/1001",
        start_at=parse_local("2030-01-10 07:00PM", "%Y-%m-%d %I:%M%p", "America/Chicago"),
        cover_cost="free",
        age="all-ages",
        venue="The Venue",
        tags=["music"],
    ))
    out.events[-1].update_guid("1001")      # the venue's own id for the event

    result = run.send(out)
    print(result.saved, result.errors)
```

[`examples/scraper.py`](examples/scraper.py) is a complete scraper that runs against a local stack:

```sh
INGEST_URL=http://localhost:8080 INGEST_KEY=... python examples/scraper.py
```

The region and age group names must exist on the server (see "Before you start" in the API docs). The example uses `tulsa` and `all-ages`; change them to names from your `config.yaml`.

## What it gives you

| | |
|---|---|
| `Client(base_url, key, timeout=120)` | `base_url` is the backend itself, not the public site: `/internal` is blocked at nginx |
| `client.run()` | A context manager that starts a run and always finishes it, even if the scrape raises. Finishing sends the error email |
| `client.start_run()`, `client.send(output, run_id=0)`, `client.finish_run(run_id)` | The same steps by hand. A `run_id` of 0 sends outside any run |
| `VenueInfo`, `Event`, `IngestionOutput` | The payload. `to_dict()` is the JSON the server expects, with empty optional fields left out |
| `parse_local(text, format, tz)` | A wall-clock time in the venue's zone, as a timezone-aware `datetime` |
| `Event.update_guid(source_id)` | Identity from the source's own id, `"<venue>-<id>"` like the Go client. Without one the server derives it |
| `Event.validate(payload_venue)` | Raises `ValueError` for what the server would refuse an event for |
| `IngestionOutput.append_error(e)` | Records a failure, prefixed with the venue's name, to be sent with the payload |
| `IngestError` | A refused request (`.status` is the HTTP status) or one that never reached the server (`.status` is `None`) |

Times without a timezone are refused when the payload is built, because the server can't know which zone a venue meant. This is the most common mistake in a scraper.

## Tests

```sh
python3 -m unittest discover -s tests -t .
```

or `make test-python` from the repo root. The tests run the client against a stand-in server in the same process.
