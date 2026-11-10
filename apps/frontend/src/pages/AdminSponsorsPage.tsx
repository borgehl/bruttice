import { useCallback, useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { fetchAdminSponsors } from '../services/api';
import { useAuthStore } from '../store/auth';
import type { AdminSponsor } from '../types/sponsor';
import { toLoadError, type LoadError } from '../utils/loadError';
import Page from '../components/ui/Page';
import Skeleton from '../components/ui/Skeleton';
import LoadFailed from '../components/ui/LoadFailed';
import InlineError from '../components/ui/InlineError';
import LoadingBlock from '../components/ui/LoadingBlock';
import AdminSponsors from '../components/features/AdminSponsors';
import AdminSubHeader from '../components/features/AdminSubHeader';

// The sponsor cards: add them, set the days they run, and see how often they were seen and clicked.
export default function AdminSponsorsPage() {
  const { isAuthenticated, isLoading: authLoading, checkSession } = useAuthStore();
  const navigate = useNavigate();
  const [sponsors, setSponsors] = useState<AdminSponsor[] | null>(null);
  const [loadError, setLoadError] = useState<LoadError | null>(null);
  // bumped to load again, after every change and by Try again
  const [attempt, setAttempt] = useState(0);

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
    fetchAdminSponsors()
      .then((loaded) => {
        if (cancelled) return;
        setSponsors(loaded);
        setLoadError(null);
      })
      .catch((err) => {
        if (cancelled) return;
        if (err instanceof Error && err.message === 'Authentication required') {
          navigate('/admin/login');
          return;
        }
        setLoadError(toLoadError(err));
      });
    return () => {
      cancelled = true;
    };
  }, [isAuthenticated, attempt, navigate]);

  const reload = useCallback(() => setAttempt((n) => n + 1), []);
  const retry = () => {
    setSponsors(null);
    setLoadError(null);
    setAttempt((n) => n + 1);
  };

  if (authLoading || !isAuthenticated) {
    return (
      <Page title="Sponsors">
        <LoadingBlock label="Checking your session" />
      </Page>
    );
  }

  return (
    <Page title="Sponsors" subtitle="Paid cards in the calendar." subheader={<AdminSubHeader current="/admin/sponsors" />}>
      {loadError && !sponsors && <LoadFailed error={loadError} onRetry={retry} />}
      {!sponsors && !loadError && <Skeleton variant="adminCard" count={2} label="Loading the sponsors" />}
      {sponsors && (
        <>
          {loadError && <InlineError message={`${loadError.title}. ${loadError.message}`} onRetry={retry} className="mb-3" />}
          <AdminSponsors sponsors={sponsors} reload={reload} />
        </>
      )}
    </Page>
  );
}
