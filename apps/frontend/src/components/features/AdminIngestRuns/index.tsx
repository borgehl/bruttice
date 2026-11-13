import { useCallback, useEffect, useRef, useState } from 'react';
import Skeleton from '../../ui/Skeleton';
import Spinner from '../../ui/Spinner';
import EmptyState from '../../ui/EmptyState';
import InlineError from '../../ui/InlineError';
import { toLoadError } from '../../../utils/loadError';
import { sourceDetailSummary } from '../../../utils/ingestRuns';
import { fetchIngestRuns, startIngest } from '../../../services/api';
import type { IngestRun, IngestSource } from '../../../types/event';

// Loads more when the bottom sentinel is within this distance of the visible area.
const LOAD_MARGIN = '200px 0px';

// How often to check for the new run's progress after starting an ingestion, and when to give up.
const POLL_INTERVAL_MS = 5000;
const POLL_TIMEOUT_MS = 5 * 60 * 1000;

const statusStyles: Record<IngestSource['status'], string> = {
  ok: 'text-success',
  skipped: 'text-text-muted',
  errors: 'text-warning',
  failed: 'text-danger',
};

const statusLabels: Record<IngestSource['status'], string> = {
  ok: 'OK',
  skipped: 'Skipped',
  errors: 'Errors',
  failed: 'Failed',
};

function plural(n: number, word: string): string {
  return `${n} ${word}${n === 1 ? '' : 's'}`;
}

// The columns shared by a run's header, totals and source rows so the numbers line up. Below sm
// the source's name takes a row of its own, above the status and the four counts.
const ROW_GRID =
  'grid grid-cols-[4.5rem_repeat(4,minmax(0,1fr))] items-baseline gap-x-2 sm:grid-cols-[minmax(0,1fr)_4.5rem_repeat(4,4.5rem)]';

const COUNT_COLUMNS = ['Scraped', 'Saved', 'Failed', 'Skipped'] as const;

// The four count cells of a row, right aligned. A zero is dimmed so the other numbers stand out.
function CountCells({ counts }: { counts: [number, number, number, number] }) {
  return (
    <>
      {counts.map((n, i) => (
        <span key={COUNT_COLUMNS[i]} className={`text-right font-data tabular-nums ${n === 0 ? 'text-text-faint' : ''}`}>
          {n}
        </span>
      ))}
    </>
  );
}

function ColumnHeader() {
  return (
    <div className={`${ROW_GRID} border-b border-border pb-1 text-xs uppercase text-text-faint`}>
      <span className="hidden sm:block">Source</span>
      <span>Status</span>
      {COUNT_COLUMNS.map((name) => (
        <span key={name} className="text-right">
          {name}
        </span>
      ))}
    </div>
  );
}

function TotalsRow({ run }: { run: IngestRun }) {
  return (
    <div className={`${ROW_GRID} border-b border-border py-2 text-sm font-semibold text-text`}>
      <span className="col-span-full sm:col-span-1">{plural(run.sources.length, 'source')}</span>
      <span />
      <CountCells counts={[run.scraped, run.saved, run.failed, run.skipped]} />
    </div>
  );
}

// A section of the expanded source: its own heading, then one entry per line.
function DetailList({ title, items }: { title: string; items: string[] }) {
  return (
    <div className="mt-2 first:mt-1">
      <h4 className="font-semibold text-text">{title}</h4>
      <ul className="mt-1 space-y-1 break-words">
        {items.map((item, i) => (
          <li key={i} className="whitespace-pre-wrap">{item}</li>
        ))}
      </ul>
    </div>
  );
}

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
      <div className={`${ROW_GRID} text-sm`}>
        <span className="col-span-full break-words font-medium text-text sm:col-span-1">{source.source}</span>
        <span className={`font-semibold ${statusStyles[source.status]}`}>{statusLabels[source.status]}</span>
        <CountCells counts={[source.scrapedCount, source.savedCount, source.failedCount, source.skippedCount]} />
      </div>
      {(source.errors.length > 0 || source.skipped.length > 0) && (
        <details className="mt-1 text-xs text-text-muted">
          <summary className="cursor-pointer">
            {sourceDetailSummary(source)}
          </summary>
          {source.errors.length > 0 && <DetailList title="Errors" items={source.errors} />}
          {source.skipped.length > 0 && <DetailList title="Skipped" items={source.skipped} />}
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
      const { title, message } = toLoadError(err);
      setError(`${title}. ${message}`);
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
      {runs.length === 0 && loading && <Skeleton variant="runRow" label="Loading ingestion runs" />}
      {runs.map((run) => (
        <section key={run.id} className="rounded-control border border-border bg-surface p-4">
          <div className="flex flex-wrap justify-between gap-2 mb-2">
            <h3 className="font-data text-sm font-semibold text-text">{formatTime(run.startedAt)}</h3>
            <span className="font-data text-sm text-text-faint">{formatDuration(run)}</span>
          </div>
          {run.dedupeError && <p className="text-sm text-danger mb-2">Deduplication failed: {run.dedupeError}</p>}
          <ColumnHeader />
          <TotalsRow run={run} />
          <ul>
            {run.sources.map((source, i) => (
              <SourceRow key={`${source.source}-${i}`} source={source} />
            ))}
          </ul>
        </section>
      ))}

      <div ref={bottomRef} className="py-2 text-center text-[13px] text-text-faint">
        {loading && runs.length > 0 && (
          <span className="inline-flex items-center gap-2">
            <Spinner size="sm" />
            Loading more runs
          </span>
        )}
        {!hasMore && runs.length > 0 && 'No more runs'}
      </div>

      {!hasMore && runs.length === 0 && (
        <EmptyState title="No ingestion runs yet">Runs appear here after the first scrape.</EmptyState>
      )}

      {error && <InlineError message={error} onRetry={() => void loadMore()} />}
    </div>
    </div>
  );
}
