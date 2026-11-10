import { useCallback, useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { fetchAdminConfig } from '../services/api';
import { useAuthStore } from '../store/auth';
import type { AdminConfig } from '../types/config';
import { toLoadError, type LoadError } from '../utils/loadError';
import Page from '../components/ui/Page';
import Segmented from '../components/ui/Segmented';
import Skeleton from '../components/ui/Skeleton';
import LoadFailed from '../components/ui/LoadFailed';
import InlineError from '../components/ui/InlineError';
import LoadingBlock from '../components/ui/LoadingBlock';
import AdminSubHeader from '../components/features/AdminSubHeader';
import RegionsPanel from '../components/features/AdminConfig/RegionsPanel';
import TagsPanel from '../components/features/AdminConfig/TagsPanel';
import AgesPanel from '../components/features/AdminConfig/AgesPanel';
import VenuesPanel from '../components/features/AdminConfig/VenuesPanel';

type Section = 'regions' | 'tags' | 'ages' | 'venues';

// The lists the site is made of: regions, tags, age groups and venues. They are saved from the
// config file on the first start, and managed here after that.
export default function AdminConfigPage() {
  const { isAuthenticated, isLoading: authLoading, checkSession } = useAuthStore();
  const navigate = useNavigate();
  const [section, setSection] = useState<Section>('regions');
  const [config, setConfig] = useState<AdminConfig | null>(null);
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
    fetchAdminConfig()
      .then((loaded) => {
        if (cancelled) return;
        setConfig(loaded);
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
    setConfig(null);
    setLoadError(null);
    setAttempt((n) => n + 1);
  };

  if (authLoading || !isAuthenticated) {
    return (
      <Page title="Site lists">
        <LoadingBlock label="Checking your session" />
      </Page>
    );
  }

  return (
    <Page title="Site lists" subtitle="Regions, tags, age groups and venues." subheader={<AdminSubHeader current="/admin/config" />}>
      <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
        <Segmented
          label="List"
          value={section}
          onChange={setSection}
          options={[
            { value: 'regions', label: 'Regions' },
            { value: 'tags', label: 'Tags' },
            { value: 'ages', label: 'Ages' },
            { value: 'venues', label: 'Venues' },
          ]}
        />
      </div>

      {loadError && !config && <LoadFailed error={loadError} onRetry={retry} />}
      {!config && !loadError && <Skeleton variant="adminCard" count={3} label="Loading the lists" />}
      {config && (
        <>
          {loadError && <InlineError message={`${loadError.title}. ${loadError.message}`} onRetry={retry} className="mb-3" />}
          {section === 'regions' && <RegionsPanel regions={config.regions} reload={reload} />}
          {section === 'tags' && <TagsPanel tags={config.tags} reload={reload} />}
          {section === 'ages' && <AgesPanel ages={config.ages} reload={reload} />}
          {section === 'venues' && <VenuesPanel venues={config.venues} regions={config.regions} reload={reload} />}
        </>
      )}
    </Page>
  );
}
