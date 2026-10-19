import { useEffect, useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import Page from '../components/ui/Page';
import AdminDashboardDaySection from '../components/features/AdminDashboardDaySection';
import AdminIngestRuns from '../components/features/AdminIngestRuns';
import { useAuthStore } from '../store/auth';
import { fetchAdminDashboard, submitVote } from '../services/api';
import type { AdminDashboard } from '../types/event';

export default function AdminDashboardPage() {
  const { isAuthenticated, isLoading: authLoading, email, logout, checkSession } = useAuthStore();
  const [dashboard, setDashboard] = useState<AdminDashboard | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [tab, setTab] = useState<'events' | 'ingestion'>('events');
  const navigate = useNavigate();

  useEffect(() => {
    checkSession();
  }, [checkSession]);

  useEffect(() => {
    if (!authLoading && !isAuthenticated) {
      navigate('/admin/login');
    }
  }, [authLoading, isAuthenticated, navigate]);

  useEffect(() => {
    async function loadDashboard() {
      if (!isAuthenticated) return;

      try {
        const data = await fetchAdminDashboard();
        setDashboard(data);
      } catch (err) {
        if (err instanceof Error && err.message === 'Authentication required') {
          navigate('/admin/login');
        } else {
          setError(err instanceof Error ? err.message : 'Failed to load dashboard');
        }
      } finally {
        setLoading(false);
      }
    }

    if (isAuthenticated) {
      loadDashboard();
    }
  }, [isAuthenticated, navigate]);

  const refreshDashboard = async () => {
    try {
      const data = await fetchAdminDashboard();
      setDashboard(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to refresh dashboard');
    }
  };

  const handleVote = async (guid: string, vote: string) => {
    try {
      await submitVote({ guid, vote: vote as 'Approve' | 'Deny' | 'Keep' | 'Remove' });
      await refreshDashboard();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to submit vote');
    }
  };

  const handleLogout = async () => {
    await logout();
    navigate('/admin/login');
  };

  if (authLoading || loading) {
    return (
      <Page title="Admin">
        <p className="py-12 text-center text-text-muted">Loading...</p>
      </Page>
    );
  }

  if (error) {
    return (
      <Page title="Admin">
        <div className="rounded-control border border-danger/30 bg-danger-muted p-4">
          <p className="text-danger">{error}</p>
        </div>
      </Page>
    );
  }

  const awaiting = dashboard?.awaitingApproval ?? [];
  const awaitingCount = awaiting.reduce((total, day) => total + day.events.length, 0);
  const upcoming = dashboard?.upcoming ?? [];
  const upcomingCount = upcoming.reduce((total, day) => total + day.events.length, 0);

  return (
    <Page
      title="Admin"
      action={
        <button
          type="button"
          onClick={handleLogout}
          className="font-label h-9 cursor-pointer rounded-control border border-border px-3 text-sm font-semibold text-text-muted hover:text-text"
        >
          Logout
        </button>
      }
      flush
    >
      <p className="px-4 pt-3 text-sm text-text-muted">
        <span className="font-medium text-text">{email}</span> ·{' '}
        <Link to="/admin/add-event">add event</Link>
      </p>

      <div role="tablist" className="mt-3 flex border-b border-border px-4">
        {([
          ['events', 'Events'],
          ['ingestion', 'Ingestion'],
        ] as const).map(([key, label]) => (
          <button
            key={key}
            type="button"
            role="tab"
            aria-selected={tab === key}
            onClick={() => setTab(key)}
            className={`font-label -mb-px min-h-11 flex-1 cursor-pointer border-b-2 text-sm font-semibold ${
              tab === key ? 'border-accent text-accent' : 'border-transparent text-text-muted hover:text-text'
            }`}
          >
            {label}
          </button>
        ))}
      </div>

      {tab === 'ingestion' ? (
        <div className="p-4">
          <AdminIngestRuns />
        </div>
      ) : (
        <>
          <section>
            <h2 className="font-label flex items-baseline justify-between px-4 pb-2 pt-4 text-base font-semibold text-text">
              Awaiting approval
              <span className="font-data text-xs font-normal text-text-faint">{awaitingCount}</span>
            </h2>
            {awaiting.length > 0 ? (
              awaiting.map((day) => (
                <AdminDashboardDaySection key={day.day} eventDay={day} mode="approval" onVote={handleVote} onUpdate={refreshDashboard} />
              ))
            ) : (
              <p className="px-4 pb-4 text-sm text-text-muted">No events awaiting approval.</p>
            )}
          </section>

          <section>
            <h2 className="font-label flex items-baseline justify-between border-t border-border px-4 pb-2 pt-4 text-base font-semibold text-text">
              Upcoming awaiting judgement
              <span className="font-data text-xs font-normal text-text-faint">{upcomingCount}</span>
            </h2>
            {upcoming.length > 0 ? (
              upcoming.map((day) => (
                <AdminDashboardDaySection key={day.day} eventDay={day} mode="upcoming" onVote={handleVote} onUpdate={refreshDashboard} />
              ))
            ) : (
              <p className="px-4 pb-4 text-sm text-text-muted">No upcoming events.</p>
            )}
          </section>
        </>
      )}
    </Page>
  );
}
