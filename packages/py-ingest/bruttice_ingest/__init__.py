"""Write bruttice scrapers in Python.

    from bruttice_ingest import Client, Event, IngestionOutput, VenueInfo, parse_local

See docs/writing-scrapers.md for the API, and examples/scraper.py for a complete scraper.
"""

from .client import Client, IngestError, Run
from .types import (
    Event,
    IngestionOutput,
    Result,
    VenueInfo,
    parse_local,
    utc,
)

__all__ = [
    "Client",
    "Event",
    "IngestError",
    "IngestionOutput",
    "Result",
    "Run",
    "VenueInfo",
    "parse_local",
    "utc",
]
