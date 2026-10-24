import json
import threading
import unittest
from datetime import datetime, timezone
from http.server import BaseHTTPRequestHandler, HTTPServer

from bruttice_ingest import Client, Event, IngestError, IngestionOutput, VenueInfo

KEY = "secret-key"
KNOWN_VENUE_FIELDS = {"name", "link", "is_auto_ingest", "require_approval", "default_region", "skip_venue_save", "timezone"}
KNOWN_EVENT_FIELDS = {"guid", "title", "link", "start_at", "doors_at", "cover_cost", "flyer_url", "venue", "region", "age", "cancelled", "tags"}


class FakeServer:
    """Stands in for the backend's /internal/ingest endpoints, with the same rules about the key
    and about fields it does not know."""

    def __init__(self):
        self.requests = []
        self.fail_with = None  # (status, body) for the next ingest request
        outer = self

        class Handler(BaseHTTPRequestHandler):
            def log_message(self, *args):
                pass

            def reply(self, status, body=None):
                data = b"" if body is None else json.dumps(body).encode()
                self.send_response(status)
                self.send_header("Content-Type", "application/json")
                self.send_header("Content-Length", str(len(data)))
                self.end_headers()
                self.wfile.write(data)

            def do_POST(self):
                length = int(self.headers.get("Content-Length") or 0)
                raw = self.rfile.read(length) if length else b""
                body = json.loads(raw) if raw else None
                outer.requests.append((self.path, body))
                if self.headers.get("Authorization") != "Bearer " + KEY:
                    return self.reply(401, {"error": "Unauthorized"})
                if self.path == "/internal/ingest/runs":
                    return self.reply(201, {"id": 12})
                if self.path == "/internal/ingest/runs/12/finish":
                    return self.reply(204)
                if self.path == "/internal/ingest":
                    if outer.fail_with:
                        status, message = outer.fail_with
                        return self.reply(status, {"error": message})
                    unknown = set(body["venue_info"]) - KNOWN_VENUE_FIELDS
                    for e in body["events"]:
                        unknown |= set(e) - KNOWN_EVENT_FIELDS
                    if unknown:
                        return self.reply(400, {"error": "unknown field " + sorted(unknown)[0]})
                    return self.reply(200, {"saved": len(body["events"]), "errors": body["errors"] + ["event x failed"]})
                self.reply(404, {"error": "no such route"})

        self.httpd = HTTPServer(("127.0.0.1", 0), Handler)
        self.url = f"http://127.0.0.1:{self.httpd.server_port}"
        threading.Thread(target=self.httpd.serve_forever, daemon=True).start()

    def close(self):
        self.httpd.shutdown()
        self.httpd.server_close()


def sample_output():
    start = datetime(2030, 1, 10, 19, 0, tzinfo=timezone.utc)
    return IngestionOutput(
        VenueInfo(name="V", default_region="tulsa"),
        [Event(title="t", start_at=start, age="all-ages", venue="V", tags=["music"], doors_at=start)],
        ["V: one problem"],
    )


class ClientTest(unittest.TestCase):
    def setUp(self):
        self.server = FakeServer()
        self.addCleanup(self.server.close)
        self.client = Client(self.server.url + "/", KEY, timeout=5)

    def paths(self):
        return [p for p, _ in self.server.requests]

    def test_the_whole_flow(self):
        run_id = self.client.start_run()
        result = self.client.send(sample_output(), run_id)
        self.client.finish_run(run_id)

        self.assertEqual(run_id, 12)
        self.assertEqual(result.saved, 1)
        self.assertEqual(result.errors, ["V: one problem", "event x failed"])
        self.assertEqual(self.paths(), ["/internal/ingest/runs", "/internal/ingest", "/internal/ingest/runs/12/finish"])
        sent = self.server.requests[1][1]
        self.assertEqual(sent["run_id"], 12)
        self.assertEqual(sent["events"][0]["start_at"], "2030-01-10T19:00:00+00:00")

    def test_a_send_without_a_run(self):
        self.client.send(sample_output())
        self.assertNotIn("run_id", self.server.requests[0][1])

    def test_a_wrong_key_is_refused(self):
        with self.assertRaises(IngestError) as caught:
            Client(self.server.url, "wrong", timeout=5).start_run()
        self.assertEqual(caught.exception.status, 401)
        self.assertIn("Unauthorized", str(caught.exception))

    def test_the_servers_reason_is_in_the_error(self):
        self.server.fail_with = (500, "venue could not be saved")
        with self.assertRaises(IngestError) as caught:
            self.client.send(sample_output())
        self.assertEqual(caught.exception.status, 500)
        self.assertIn("venue could not be saved", str(caught.exception))

    def test_a_server_that_cannot_be_reached(self):
        url = self.server.url
        self.server.close()
        with self.assertRaises(IngestError) as caught:
            Client(url, KEY, timeout=2).start_run()
        self.assertIsNone(caught.exception.status)

    def test_the_payload_has_only_fields_the_server_knows(self):
        # the fake refuses an unknown field with a 400, like the real server
        self.client.send(sample_output())


class RunTest(unittest.TestCase):
    def setUp(self):
        self.server = FakeServer()
        self.addCleanup(self.server.close)
        self.client = Client(self.server.url, KEY, timeout=5)

    def paths(self):
        return [p for p, _ in self.server.requests]

    def test_a_run_is_finished(self):
        with self.client.run() as run:
            result = run.send(sample_output())
        self.assertEqual(result.saved, 1)
        self.assertEqual(self.paths(), ["/internal/ingest/runs", "/internal/ingest", "/internal/ingest/runs/12/finish"])
        self.assertEqual(self.server.requests[1][1]["run_id"], 12)

    def test_a_run_is_finished_when_the_scrape_raises(self):
        with self.assertRaisesRegex(RuntimeError, "site is down"):
            with self.client.run():
                raise RuntimeError("site is down")
        self.assertEqual(self.paths()[-1], "/internal/ingest/runs/12/finish")

    def test_the_scrapes_failure_is_not_hidden_by_a_failed_finish(self):
        with self.assertRaisesRegex(RuntimeError, "site is down"):
            with self.client.run():
                self.server.close()  # finishing will fail too
                raise RuntimeError("site is down")


if __name__ == "__main__":
    unittest.main()
