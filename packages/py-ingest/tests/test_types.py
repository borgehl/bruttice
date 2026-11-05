import unittest
from datetime import datetime, timedelta, timezone

from bruttice_ingest import Event, IngestionOutput, VenueInfo, parse_local


class ParseLocalTest(unittest.TestCase):
    def test_reads_the_time_in_the_venues_zone(self):
        t = parse_local("2030-01-10 07:00PM", "%Y-%m-%d %I:%M%p", "America/Chicago")
        self.assertEqual(t.utcoffset(), timedelta(hours=-6))
        self.assertEqual(t.isoformat(), "2030-01-10T19:00:00-06:00")

    def test_follows_daylight_saving(self):
        t = parse_local("2030-07-10 07:00PM", "%Y-%m-%d %I:%M%p", "America/Chicago")
        self.assertEqual(t.utcoffset(), timedelta(hours=-5))


class EventTest(unittest.TestCase):
    def setUp(self):
        self.start = datetime(2030, 1, 10, 19, 0, tzinfo=timezone(timedelta(hours=-6)))

    def test_wire_format_leaves_out_empty_optionals(self):
        e = Event(title="Open Mic", start_at=self.start, age="all-ages")
        self.assertEqual(
            e.to_dict(),
            {"title": "Open Mic", "link": "", "start_at": "2030-01-10T19:00:00-06:00", "cover_cost": "", "age": "all-ages"},
        )

    def test_wire_format_with_everything(self):
        e = Event(
            title="Open Mic", start_at=self.start, age="all-ages", link="https://x.example/1",
            doors_at=self.start - timedelta(hours=1), cover_cost="$10", flyer_url="https://x.example/f.jpg",
            venue="V", region="tulsa", tags=["music"], cancelled=True, guid="V-1",
        )
        self.assertEqual(
            e.to_dict(),
            {
                "title": "Open Mic", "link": "https://x.example/1", "start_at": "2030-01-10T19:00:00-06:00",
                "doors_at": "2030-01-10T18:00:00-06:00", "cover_cost": "$10", "flyer_url": "https://x.example/f.jpg",
                "venue": "V", "region": "tulsa", "age": "all-ages", "cancelled": True, "tags": ["music"], "guid": "V-1",
            },
        )

    def test_a_time_without_a_zone_is_refused(self):
        e = Event(title="Open Mic", start_at=datetime(2030, 1, 10, 19, 0), age="all-ages")
        with self.assertRaisesRegex(ValueError, "start_at has no timezone"):
            e.to_dict()
        e = Event(title="Open Mic", start_at=self.start, age="all-ages", doors_at=datetime(2030, 1, 10, 18, 0))
        with self.assertRaisesRegex(ValueError, "doors_at has no timezone"):
            e.to_dict()

    def test_guid_from_a_source_id(self):
        e = Event(title="t", start_at=self.start, age="a", venue="The Venue")
        e.update_guid("1001")
        self.assertEqual(e.guid, "The Venue-1001")
        e.update_guid("2002")
        self.assertEqual(e.guid, "The Venue-1001", "a guid that is set stays")

    def test_no_source_id_leaves_the_guid_to_the_server(self):
        e = Event(title="t", start_at=self.start, age="a", venue="V")
        e.update_guid(None)
        e.update_guid("")
        self.assertEqual(e.guid, "")
        self.assertNotIn("guid", e.to_dict())

    def test_validate_names_what_is_missing(self):
        with self.assertRaisesRegex(ValueError, "missing title, age, venue"):
            Event(title=" ", start_at=self.start, age="").validate()
        Event(title="t", start_at=self.start, age="a").validate(payload_venue="V")
        Event(title="t", start_at=self.start, age="a", venue="V").validate()

    def test_validate_requires_a_start(self):
        with self.assertRaisesRegex(ValueError, "missing start_at"):
            Event(title="t", start_at=None, age="a", venue="V").validate()
        with self.assertRaisesRegex(ValueError, "missing start_at"):
            Event(title="t", start_at=None, age="a", venue="V").to_dict()

    def test_cancellation_only_events(self):
        stub = Event(title="", start_at=None, age="", venue="V", cancelled=True, guid="V-abc")
        self.assertTrue(stub.is_cancellation_only())
        stub.validate()
        self.assertEqual(
            stub.to_dict(),
            {"title": "", "link": "", "cover_cost": "", "age": "", "venue": "V", "cancelled": True, "guid": "V-abc"},
        )
        no_age = Event(title="t", start_at=self.start, age="", cancelled=True, guid="V-1")
        self.assertTrue(no_age.is_cancellation_only())
        no_age.validate(payload_venue="V")

        full = Event(title="t", start_at=self.start, age="a", cancelled=True)
        self.assertFalse(full.is_cancellation_only())
        self.assertFalse(Event(title="t", start_at=None, age="a").is_cancellation_only())

    def test_cancellation_only_needs_a_guid_and_a_venue(self):
        with self.assertRaisesRegex(ValueError, "missing guid"):
            Event(title="", start_at=None, age="", venue="V", cancelled=True).validate()
        with self.assertRaisesRegex(ValueError, "missing venue"):
            Event(title="", start_at=None, age="", cancelled=True, guid="V-1").validate()

    def test_tags_do_not_leak_between_events(self):
        a, b = Event(title="a", start_at=self.start, age="x"), Event(title="b", start_at=self.start, age="x")
        a.add_tag("music")
        self.assertEqual(b.tags, [])


class IngestionOutputTest(unittest.TestCase):
    def test_wire_format(self):
        start = datetime(2030, 1, 10, 19, 0, tzinfo=timezone.utc)
        out = IngestionOutput(
            VenueInfo(name="V", link="https://v.example", default_region="tulsa"),
            [Event(title="t", start_at=start, age="a")],
        )
        self.assertEqual(
            out.to_dict(),
            {
                "venue_info": {
                    "name": "V", "link": "https://v.example", "is_auto_ingest": True,
                    "default_region": "tulsa", "skip_venue_save": False, "timezone": "",
                },
                "events": [out.events[0].to_dict()],
                "errors": [],
            },
        )

    def test_run_id_and_approval_only_when_set(self):
        out = IngestionOutput(VenueInfo(name="V", require_approval=True))
        d = out.to_dict(run_id=12)
        self.assertEqual(d["run_id"], 12)
        self.assertTrue(d["venue_info"]["require_approval"])
        self.assertNotIn("run_id", out.to_dict())

    def test_errors_are_prefixed_with_the_venue(self):
        out = IngestionOutput(VenueInfo(name="V"))
        out.append_error(ValueError("bad row"))
        out.append_error("fetching: timeout")
        self.assertEqual(out.errors, ["V: bad row", "V: fetching: timeout"])
        self.assertEqual(out.to_dict()["errors"], out.errors)


if __name__ == "__main__":
    unittest.main()
