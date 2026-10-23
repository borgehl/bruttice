import { useEffect, useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import Page from '../components/ui/Page';
import AdminDashboardDaySection from '../components/features/AdminDashboardDaySection';
import AdminIngestRuns from '../components/features/AdminIngestRuns';
import AdminInbox from '../components/features/AdminInbox';
import { useAuthStore } from '../store/auth';
import { fetchAdminDashboard, submitVote } from '../services/api';
import Skeleton from '../components/ui/Skeleton';
import EmptyState from '../components/ui/EmptyState';
import InlineError from '../components/ui/InlineError';
import LoadFailed from '../components/ui/LoadFailed';
import { toLoadError, type LoadError } from '../utils/loadError';
import type { AdminDashboard } from '../types/event';

export default function AdminDashboardPage() {
  const { isAuthenticated, isLoading: authLoading, email, name, logout, checkSession } = useAuthStore();
  const [dashboard, setDashboard] = useState<AdminDashboard | null>(null);
  const [loading, setLoading] = useState(true);
  // the dashboard could not be loaded at all, or a later action failed while it is on screen
  const [loadError, setLoadError] = useState<LoadError | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);
  // bumped by Try again to load the dashboard once more
  const [attempt, setAttempt] = useState(0);
  const [tab, setTab] = useState<'events' | 'inbox' | 'ingestion'>('events');
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
        setLoadError(null);
      } catch (err) {
        if (err instanceof Error && err.message === 'Authentication required') {
          navigate('/admin/login');
        } else {
          setLoadError(toLoadError(err));
        }
      } finally {
        setLoading(false);
      }
    }

    if (isAuthenticated) {
      loadDashboard();
    }
  }, [isAuthenticated, navigate, attempt]);

  const retryLoad = () => {
    setLoading(true);
    setAttempt((n) => n + 1);
  };

  const refreshDashboard = async () => {
    try {
      const data = await fetchAdminDashboard();
      setDashboard(data);
      setActionError(null);
    } catch (err) {
      console.error(err);
      setActionError('The dashboard could not be refreshed.');
    }
  };

  const handleVote = async (guid: string, vote: string) => {
    setActionError(null);
    try {
      await submitVote({ guid, vote: vote as 'Approve' | 'Deny' | 'Keep' | 'Remove' });
      await refreshDashboard();
    } catch (err) {
      console.error(err);
      setActionError('Your vote was not saved.');
    }
  };

  const handleLogout = async () => {
    await logout();
    navigate('/admin/login');
  };

  if (authLoading || loading) {
    return (
      <Page title="Admin">
        <Skeleton variant="adminCard" count={4} label="Loading the dashboard" />
      </Page>
    );
  }

  if (loadError) {
    return (
      <Page title="Admin">
        <LoadFailed error={loadError} onRetry={retryLoad} />
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
        <span className="font-medium text-text">{name || email}</span> ·{' '}
        <Link to="/admin/add-event">add event</Link> · <Link to="/admin/settings">settings</Link>
      </p>

      {actionError && <InlineError message={actionError} className="mx-4 mt-3" />}

      <div role="tablist" className="mt-3 flex border-b border-border px-4">
        {([
          ['events', 'Events'],
          ['inbox', 'Inbox'],
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
            {key === 'inbox' && (dashboard?.unreadMessages ?? 0) > 0 && (
              <span className="font-data ml-1.5 rounded-full bg-accent px-1.5 py-0.5 text-xs text-accent-ink">
                {dashboard?.unreadMessages}
                <span className="sr-only"> unread</span>
              </span>
            )}
          </button>
        ))}
      </div>

      {tab === 'inbox' ? (
        <div className="p-4">
          <AdminInbox onChange={refreshDashboard} />
        </div>
      ) : tab === 'ingestion' ? (
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
              <EmptyState title="Nothing to review">New submissions and scraped events show up here.</EmptyState>
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
