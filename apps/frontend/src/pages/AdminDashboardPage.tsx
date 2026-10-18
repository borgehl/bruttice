import { useEffect, useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
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
      <div className="flex justify-center items-center min-h-[50vh]">
        <div className="text-text-muted">Loading...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="max-w-4xl mx-auto">
        <div className="p-4 bg-error-muted border border-error/30 rounded-lg">
          <p className="text-error">{error}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold text-text">Admin Dashboard</h1>
        <button
          onClick={handleLogout}
          className="px-4 py-2 text-sm bg-surface border border-border rounded hover:bg-surface-elevated text-text-muted"
        >
          Logout
        </button>
      </div>

      <div className="mb-6 p-4 bg-surface border border-border rounded-lg">
        <p className="text-text-muted">
          Welcome <span className="font-medium text-text">{email}</span>. The admin add event page is{' '}
          <Link to="/admin/add-event" className="text-primary hover:text-primary-hover">
            here
          </Link>
          .
        </p>
      </div>

      <div className="flex gap-2 mb-6 border-b border-border">
        {([
          ['events', 'Events'],
          ['ingestion', 'Ingestion'],
        ] as const).map(([key, label]) => (
          <button
            key={key}
            onClick={() => setTab(key)}
            className={`px-4 py-2 text-sm -mb-px border-b-2 ${
              tab === key ? 'border-primary text-text font-medium' : 'border-transparent text-text-muted hover:text-text'
            }`}
          >
            {label}
          </button>
        ))}
      </div>

      {tab === 'ingestion' ? (
        <AdminIngestRuns />
      ) : (
        <>
      <section className="mb-12">
        <h2 className="text-xl font-bold text-text mb-4">Events Awaiting Approval</h2>
        <div className="max-h-[500px] overflow-y-auto border border-border rounded-lg p-4 bg-surface">
          {dashboard?.awaitingApproval && dashboard.awaitingApproval.length > 0 ? (
            dashboard.awaitingApproval.map((day) => (
              <AdminDashboardDaySection key={day.day} eventDay={day} mode="approval" onVote={handleVote} onUpdate={refreshDashboard} />
            ))
          ) : (
            <p className="text-text-muted">No events awaiting approval.</p>
          )}
        </div>
      </section>

      <section>
        <h2 className="text-xl font-bold text-text mb-4">Upcoming Events Awaiting Fiery Judgement</h2>
        <div className="max-h-[600px] overflow-y-auto border border-border rounded-lg p-4 bg-surface">
          {dashboard?.upcoming && dashboard.upcoming.length > 0 ? (
            dashboard.upcoming.map((day) => (
              <AdminDashboardDaySection key={day.day} eventDay={day} mode="upcoming" onVote={handleVote} onUpdate={refreshDashboard} />
            ))
          ) : (
            <p className="text-text-muted">No upcoming events.</p>
          )}
        </div>
      </section>
        </>
      )}
    </div>
  );
}
