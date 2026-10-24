"""The payload of the ingest API as Python objects.

These mirror the Go types in packages/go-pkg/ingest. ``to_dict`` gives the JSON the server
expects, with empty optional fields left out the way the Go client leaves them out.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from datetime import datetime, timezone
from typing import Any, Dict, List, Optional
from zoneinfo import ZoneInfo


def _rfc3339(value: datetime, what: str) -> str:
    """A time as RFC 3339 with an offset. A time without a zone is refused: the server cannot know
    which zone the venue meant, and reading it as UTC shifts every event."""
    if value.tzinfo is None or value.utcoffset() is None:
        raise ValueError(
            f"{what} has no timezone. Venues publish wall-clock times, so attach the venue's zone, "
            "for example with parse_local(text, format, 'America/Chicago')"
        )
    return value.isoformat()


@dataclass
class VenueInfo:
    """The venue (or source) a payload is for."""

    name: str
    link: str = ""
    #: Marks the venue as fed by scrapers. The server refuses events for a venue whose flag
    #: differs, so one venue is never fed by both scrapers and people.
    is_auto_ingest: bool = True
    #: Sends the payload's events to the admin moderation queue. By default they are published.
    require_approval: bool = False
    #: A region name from the server's regions.
    default_region: str = ""
    #: Label the payload without creating a venue.
    skip_venue_save: bool = False
    #: Not used by the server.
    timezone: str = ""

    def to_dict(self) -> Dict[str, Any]:
        out: Dict[str, Any] = {
            "name": self.name,
            "link": self.link,
            "is_auto_ingest": self.is_auto_ingest,
            "default_region": self.default_region,
            "skip_venue_save": self.skip_venue_save,
            "timezone": self.timezone,
        }
        if self.require_approval:
            out["require_approval"] = True
        return out


@dataclass
class Event:
    """One event as a scraper reports it."""

    title: str
    #: A timezone-aware time, see ``parse_local``.
    start_at: datetime
    #: An age group name from the server's age groups.
    age: str
    link: str = ""
    doors_at: Optional[datetime] = None
    #: Free text: "$10", "$5-$15", "free".
    cover_cost: str = ""
    #: An http(s) link to an image of the flyer. An empty value leaves a stored flyer alone.
    flyer_url: str = ""
    #: A venue name, the payload's venue if empty.
    venue: str = ""
    #: A region name. For a listed venue the venue's default region wins.
    region: str = ""
    #: Tag names. Names the server does not have are dropped.
    tags: List[str] = field(default_factory=list)
    cancelled: bool = False
    #: The event's identity: a re-sent event with the same guid updates in place. Leave it empty to
    #: let the server derive one from title, start time and venue, or use ``update_guid``.
    guid: str = ""

    def add_tag(self, tag: str) -> None:
        self.tags.append(tag)

    def update_guid(self, source_id: Optional[str] = None) -> None:
        """Sets the guid from the source's own id for the event, "<venue>-<id>" like the Go
        client. Does nothing if the guid is set or there is no id, then the server derives one.
        Set ``venue`` first, or the payload's venue is not part of it."""
        if self.guid or not source_id:
            return
        self.guid = f"{self.venue}-{source_id}"

    def validate(self, payload_venue: str = "") -> None:
        """Raises ValueError for what the server can't guess and would refuse the event for."""
        missing = []
        if not self.title.strip():
            missing.append("title")
        if not self.age:
            missing.append("age")
        if not self.venue and not payload_venue:
            missing.append("venue")
        if missing:
            raise ValueError("missing " + ", ".join(missing))

    def to_dict(self) -> Dict[str, Any]:
        out: Dict[str, Any] = {
            "title": self.title,
            "link": self.link,
            "start_at": _rfc3339(self.start_at, "start_at"),
            "cover_cost": self.cover_cost,
            "age": self.age,
        }
        if self.guid:
            out["guid"] = self.guid
        if self.doors_at is not None:
            out["doors_at"] = _rfc3339(self.doors_at, "doors_at")
        if self.flyer_url:
            out["flyer_url"] = self.flyer_url
        if self.venue:
            out["venue"] = self.venue
        if self.region:
            out["region"] = self.region
        if self.cancelled:
            out["cancelled"] = True
        if self.tags:
            out["tags"] = list(self.tags)
        return out


@dataclass
class IngestionOutput:
    """What a scraper produced for one venue: the venue, its events, and what went wrong."""

    venue_info: VenueInfo
    events: List[Event] = field(default_factory=list)
    errors: List[str] = field(default_factory=list)

    def append_error(self, error: object) -> None:
        """Records a failure to be sent with the payload, prefixed with the venue's name. Don't
        crash and don't exit quietly: report it here and send the payload anyway."""
        self.errors.append(f"{self.venue_info.name}: {error}")

    def to_dict(self, run_id: int = 0) -> Dict[str, Any]:
        out: Dict[str, Any] = {}
        if run_id:
            out["run_id"] = run_id
        out["venue_info"] = self.venue_info.to_dict()
        out["events"] = [e.to_dict() for e in self.events]
        out["errors"] = list(self.errors)
        return out


@dataclass
class Result:
    """The server's answer to a payload."""

    #: How many events the server saved.
    saved: int = 0
    #: The payload's own errors, then one for each event that failed to save. The rest were saved.
    errors: List[str] = field(default_factory=list)


def parse_local(text: str, fmt: str, tz: str) -> datetime:
    """Reads a wall-clock time in the zone the venue publishes it in, such as 'America/Chicago'.

    ``datetime.strptime`` alone gives a time without a zone, which the client refuses. Attaching UTC
    instead would shift every event, which is the most common mistake in a scraper."""
    return datetime.strptime(text, fmt).replace(tzinfo=ZoneInfo(tz))


def utc(value: datetime) -> datetime:
    """The same moment in UTC."""
    return value.astimezone(timezone.utc)
