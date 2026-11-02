"""A client for the ingest API, see docs/writing-scrapers.md."""

from __future__ import annotations

import json
import urllib.error
import urllib.request
from contextlib import contextmanager
from typing import Any, Dict, Iterator, Optional

from .types import IngestionOutput, Result


class IngestError(Exception):
    """A request the server refused or that never reached it."""

    def __init__(self, message: str, status: Optional[int] = None):
        super().__init__(message)
        #: The HTTP status, None if the server could not be reached.
        self.status = status


class Client:
    """Posts scraper output to a bruttice backend's /internal/ingest endpoints.

    ``base_url`` is the backend itself, such as "http://bruttice-backend:8080": /internal is blocked
    at nginx. ``key`` is the backend's ingest-key. A client holds no connection, so it is safe to
    share between threads.
    """

    def __init__(self, base_url: str, key: str, timeout: float = 120.0):
        self.base_url = base_url.rstrip("/")
        self.key = key
        #: Seconds to wait for the server.
        self.timeout = timeout

    def start_run(self) -> int:
        """Opens a run. Payloads sent with its id are grouped in the admin ingestion log."""
        return int(self._post("/internal/ingest/runs")["id"])

    def finish_run(self, run_id: int) -> None:
        """Finishes a run. The server then emails the root admin any errors the run recorded."""
        self._post(f"/internal/ingest/runs/{int(run_id)}/finish")

    def send(self, output: IngestionOutput, run_id: int = 0) -> Result:
        """Posts one venue's output. A run id of 0 sends it outside any run, and the server
        records it as a run of its own."""
        body = self._post("/internal/ingest", output.to_dict(run_id))
        return Result(saved=int(body.get("saved", 0)), errors=list(body.get("errors") or []))

    @contextmanager
    def run(self) -> Iterator["Run"]:
        """A run that is always finished, even if the scrape raises::

            with client.run() as run:
                run.send(scrape())

        Finishing is what sends the error email, so it must happen after a failure too.
        """
        run = Run(self, self.start_run())
        try:
            yield run
        except BaseException:
            try:
                self.finish_run(run.id)
            except IngestError:
                pass  # the scrape's own failure is the one to show
            raise
        else:
            self.finish_run(run.id)

    def _post(self, path: str, body: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        data = None if body is None else json.dumps(body).encode()
        request = urllib.request.Request(self.base_url + path, data=data, method="POST")
        request.add_header("Authorization", "Bearer " + self.key)
        request.add_header("Content-Type", "application/json")
        try:
            with urllib.request.urlopen(request, timeout=self.timeout) as response:
                raw = response.read()
        except urllib.error.HTTPError as err:
            message = err.read(1024).decode("utf-8", "replace").strip()
            raise IngestError(f"POST {path}: {err.code} {err.reason}: {message}", err.code) from None
        except (urllib.error.URLError, OSError) as err:
            raise IngestError(f"POST {path}: {getattr(err, 'reason', err)}") from None
        return json.loads(raw) if raw.strip() else {}


class Run:
    """An open run, see ``Client.run``."""

    def __init__(self, client: Client, run_id: int):
        self._client = client
        self.id = run_id

    def send(self, output: IngestionOutput) -> Result:
        return self._client.send(output, self.id)
