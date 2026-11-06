import { useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { fetchVisitorStats } from '../services/api';
import { useAuthStore } from '../store/auth';
import type { VisitorStats } from '../types/event';
import Page from '../components/ui/Page';
import InlineError from '../components/ui/InlineError';
import LoadingBlock from '../components/ui/LoadingBlock';
import { describeSubmitError } from '../utils/loadError';

const RANGES = [7, 30, 90, 365];

// "Portland, OR, US"; the parts that are not known are left out
function placeLabel(l: VisitorStats['locations'][number]): string {
  return [l.city, l.state, l.country].filter(Boolean).join(', ') || 'Unknown';
}

export default function AdminVisitorsPage() {
  const { isAuthenticated, isLoading: authLoading, checkSession } = useAuthStore();
  const navigate = useNavigate();
  const [days, setDays] = useState(30);
  const [stats, setStats] = useState<VisitorStats | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    checkSession();
  }, [checkSession]);

  useEffect(() => {
    if (!authLoading && !isAuthenticated) {
      navigate('/admin/login');
    }
  }, [authLoading, isAuthenticated, navigate]);

  useEffect(() => {
    if (!isAuthenticated) return;
    let cancelled = false;
    fetchVisitorStats(days)
      .then((result) => {
        if (cancelled) return;
        setStats(result);
        setError(null);
      })
      .catch((err) => {
        if (cancelled) return;
        if (err instanceof Error && err.message === 'Authentication required') {
          navigate('/admin/login');
          return;
        }
        console.error(err);
        setError(describeSubmitError(err));
      });
    return () => {
      cancelled = true;
    };
  }, [isAuthenticated, days, navigate]);

  if (authLoading || !isAuthenticated) {
    return (
      <Page title="Visitors">
        <LoadingBlock label="Checking your session" />
      </Page>
    );
  }

  // the stats shown belong to the range that was asked for, not to the previous one
  const ready = stats !== null && stats.days === days;
  const maxDaily = ready ? Math.max(1, ...stats.daily.map((d) => d.visitors)) : 1;
  const maxLocation = ready ? Math.max(1, ...stats.locations.map((l) => l.visitors)) : 1;

  return (
    <Page title="Visitors" subtitle="Only visitors who accepted cookies are counted.">
      <div role="group" aria-label="Time range" className="mb-4 flex gap-2">
        {RANGES.map((range) => (
          <button
            key={range}
            type="button"
            aria-pressed={range === days}
            onClick={() => setDays(range)}
            className={`font-label h-9 cursor-pointer rounded-control border px-3 text-sm font-semibold ${
              range === days ? 'border-accent text-text' : 'border-border text-text-muted hover:text-text'
            }`}
          >
            {range}d
          </button>
        ))}
      </div>

      {error && <InlineError message={error} />}
      {!ready && !error && <LoadingBlock label="Loading visitors" />}

      {ready && (
        <div className="space-y-6">
          <section aria-labelledby="visitors-by-date">
            <h2 id="visitors-by-date" className="mb-2 text-sm font-semibold text-text">
              Visitors by date
            </h2>
            <div className="flex h-32 items-end gap-px" role="img" aria-label="Visitors per day">
              {stats.daily.map((d) => (
                <div
                  key={d.date}
                  title={`${d.date}: ${d.visitors}`}
                  className="min-h-px flex-1 bg-accent"
                  style={{ height: `${(d.visitors / maxDaily) * 100}%`, opacity: d.visitors ? 1 : 0.2 }}
                />
              ))}
            </div>
            <div className="mt-1 flex justify-between text-xs text-text-faint">
              <span>{stats.daily[0]?.date}</span>
              <span>{stats.daily[stats.daily.length - 1]?.date}</span>
            </div>
            <details className="mt-2">
              <summary className="cursor-pointer text-xs text-text-muted">Table</summary>
              <table className="mt-1 w-full text-sm">
                <thead className="text-left text-xs text-text-faint">
                  <tr>
                    <th className="py-1 font-medium">Date</th>
                    <th className="py-1 text-right font-medium">Visitors</th>
                  </tr>
                </thead>
                <tbody>
                  {[...stats.daily].reverse().map((d) => (
                    <tr key={d.date} className="border-t border-border">
                      <td className="py-1">{d.date}</td>
                      <td className="py-1 text-right tabular-nums">{d.visitors}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </details>
          </section>

          <section aria-labelledby="visitors-by-location">
            <h2 id="visitors-by-location" className="mb-2 text-sm font-semibold text-text">
              Visitors by location
            </h2>
            {stats.locations.length === 0 ? (
              <p className="text-sm text-text-muted">No visitors in this time range.</p>
            ) : (
              <ul className="space-y-2">
                {stats.locations.map((l) => (
                  <li key={`${l.country}|${l.state}|${l.city}`} className="text-sm">
                    <div className="flex justify-between gap-3">
                      <span className="text-text">{placeLabel(l)}</span>
                      <span className="tabular-nums text-text-muted">{l.visitors}</span>
                    </div>
                    <div className="mt-1 h-1.5 bg-border">
                      <div className="h-full bg-accent" style={{ width: `${(l.visitors / maxLocation) * 100}%` }} />
                    </div>
                  </li>
                ))}
              </ul>
            )}
            <p className="mt-2 text-xs text-text-faint">
              Each visitor is counted once per place, however many days they came.
            </p>
          </section>
        </div>
      )}

      <p className="mt-6 text-center text-sm">
        <Link to="/admin">Back to the dashboard</Link>
      </p>
    </Page>
  );
}
