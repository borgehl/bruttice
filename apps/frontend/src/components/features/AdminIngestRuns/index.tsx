import { useCallback, useEffect, useRef, useState } from 'react';
import { fetchIngestRuns, startIngest } from '../../../services/api';
import type { IngestRun, IngestSource } from '../../../types/event';

// Loads more when the bottom sentinel is within this distance of the visible area.
const LOAD_MARGIN = '200px 0px';

// How often to check for the new run's progress after starting an ingestion, and when to give up.
const POLL_INTERVAL_MS = 5000;
const POLL_TIMEOUT_MS = 5 * 60 * 1000;

const statusStyles: Record<IngestSource['status'], string> = {
  ok: 'text-success',
  errors: 'text-warning',
  failed: 'text-danger',
};

const statusLabels: Record<IngestSource['status'], string> = { ok: 'OK', errors: 'Errors', failed: 'Failed' };

function formatTime(unix: number): string {
  return new Date(unix * 1000).toLocaleString();
}

function formatDuration(run: IngestRun): string {
  if (!run.finishedAt) return 'in progress or interrupted';
  const secs = run.finishedAt - run.startedAt;
  return secs >= 60 ? `${Math.floor(secs / 60)}m ${secs % 60}s` : `${secs}s`;
}

function SourceRow({ source }: { source: IngestSource }) {
  return (
    <li className="py-2 border-t border-border first:border-t-0">
      <div className="flex flex-wrap justify-between gap-2 text-sm">
        <span className="font-medium text-text">{source.source}</span>
        <span className="text-text-muted">
          <span className={`font-semibold ${statusStyles[source.status]}`}>{statusLabels[source.status]}</span>
          {' · '}
          {source.scrapedCount} scraped · {source.savedCount} saved · {source.failedCount} failed
        </span>
      </div>
      {source.errors.length > 0 && (
        <details className="mt-1 text-xs text-text-muted">
          <summary className="cursor-pointer">
            {source.errors.length} error{source.errors.length === 1 ? '' : 's'}
          </summary>
          <ul className="mt-1 space-y-1 break-words">
            {source.errors.map((err, i) => (
              <li key={i}>{err}</li>
            ))}
          </ul>
        </details>
      )}
    </li>
  );
}

export default function AdminIngestRuns() {
  const [runs, setRuns] = useState<IngestRun[]>([]);
  const [hasMore, setHasMore] = useState(true);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const bottomRef = useRef<HTMLDivElement>(null);
  const inFlight = useRef(false);
  const [starting, setStarting] = useState(false);
  const [polling, setPolling] = useState(false);
  const [startError, setStartError] = useState<string | null>(null);
  // newest run id when ingestion was started, the run we are waiting for is anything newer
  const baselineId = useRef(0);

  const loadMore = useCallback(async () => {
    if (inFlight.current) return;
    inFlight.current = true;
    setLoading(true);
    setError(null);
    try {
      const oldest = runs.length > 0 ? runs[runs.length - 1].id : undefined;
      const page = await fetchIngestRuns(oldest);
      setRuns((prev) => [...prev, ...page.runs]);
      setHasMore(page.hasMore);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load ingest runs');
    } finally {
      inFlight.current = false;
      setLoading(false);
    }
  }, [runs]);

  const handleStart = async () => {
    setStarting(true);
    setStartError(null);
    try {
      await startIngest();
      baselineId.current = runs[0]?.id ?? 0;
      setPolling(true);
    } catch (err) {
      setStartError(err instanceof Error ? err.message : 'Failed to start ingestion');
    } finally {
      setStarting(false);
    }
  };

  // Poll the first page until the run we started has finished, merging it into the list.
  useEffect(() => {
    if (!polling) return;

    const startedAt = Date.now();
    const timer = setInterval(async () => {
      if (Date.now() - startedAt > POLL_TIMEOUT_MS) {
        setPolling(false);
        return;
      }
      try {
        const page = await fetchIngestRuns();
        setRuns((prev) => {
          const fresh = new Set(page.runs.map((r) => r.id));
          return [...page.runs, ...prev.filter((r) => !fresh.has(r.id))].sort((a, b) => b.id - a.id);
        });
        const newest = page.runs[0];
        if (newest && newest.id > baselineId.current && newest.finishedAt > 0) {
          setPolling(false);
        }
      } catch {
        // transient, try again on the next tick
      }
    }, POLL_INTERVAL_MS);

    return () => clearInterval(timer);
  }, [polling]);

  // The observer is rebuilt after every load: a new observer reports the current intersection,
  // which keeps loading while the sentinel is still on screen.
  useEffect(() => {
    const bottom = bottomRef.current;
    if (!bottom || !hasMore || error) return;

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries.some((entry) => entry.isIntersecting)) void loadMore();
      },
      // no root: the page's own scroll area clips the sentinel, which is what counts
      { rootMargin: LOAD_MARGIN },
    );
    observer.observe(bottom);

    return () => observer.disconnect();
  }, [hasMore, error, loadMore, loading]);

  return (
    <div>
      <div className="mb-4 flex flex-col gap-2">
        <button
          type="button"
          onClick={() => void handleStart()}
          disabled={starting || polling}
          className="font-label chamfer h-11 w-full cursor-pointer rounded-control bg-accent text-sm font-semibold text-accent-ink enabled:hover:bg-accent-hover disabled:cursor-not-allowed disabled:opacity-50"
        >
          {polling ? 'Ingestion running...' : 'Ingest now'}
        </button>
        {startError && <span className="text-sm text-danger">{startError}</span>}
      </div>
    <div className="space-y-3">
      {runs.map((run) => (
        <section key={run.id} className="rounded-control border border-border bg-surface p-4">
          <div className="flex flex-wrap justify-between gap-2 mb-2">
            <h3 className="font-data text-sm font-semibold text-text">{formatTime(run.startedAt)}</h3>
            <span className="font-data text-sm text-text-faint">{formatDuration(run)}</span>
          </div>
          <p className="text-sm text-text-muted mb-2">
            {run.sources.length} sources · {run.scraped} scraped · {run.saved} saved · {run.failed} failed
          </p>
          {run.dedupeError && <p className="text-sm text-danger mb-2">Deduplication failed: {run.dedupeError}</p>}
          <ul>
            {run.sources.map((source, i) => (
              <SourceRow key={`${source.source}-${i}`} source={source} />
            ))}
          </ul>
        </section>
      ))}

      <div ref={bottomRef} className="py-2 text-center text-[13px] text-text-faint">
        {loading && 'Loading more runs...'}
        {!hasMore && (runs.length === 0 ? 'No ingestion runs recorded yet.' : 'No more runs')}
      </div>

      {error && (
        <div className="rounded-control border border-danger/30 bg-danger-muted p-3 text-center">
          <p className="text-danger text-sm">{error}</p>
          <button type="button" onClick={() => void loadMore()} className="mt-1 text-sm text-accent hover:underline cursor-pointer">
            Try again
          </button>
        </div>
      )}
    </div>
    </div>
  );
}
