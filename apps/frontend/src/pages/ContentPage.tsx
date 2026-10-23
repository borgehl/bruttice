import { useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { fetchPage } from '../services/api';
import type { SitePage } from '../types/event';
import Page from '../components/ui/Page';
import Markdown from '../components/ui/Markdown';
import EmbedFrame from '../components/ui/EmbedFrame';
import LoadingBlock from '../components/ui/LoadingBlock';
import LoadFailed from '../components/ui/LoadFailed';
import { describeLoadError, type LoadError } from '../utils/loadError';

type State = { status: 'loading' } | { status: 'missing' } | { status: 'error'; error: LoadError } | { status: 'ready'; page: SitePage };

// A page the deployment supplies, at /p/<slug>.
export default function ContentPage() {
  const { slug = '' } = useParams();
  const [loaded, setLoaded] = useState<{ slug: string; state: State } | null>(null);
  // bumped by Try again to fetch the page once more
  const [attempt, setAttempt] = useState(0);

  useEffect(() => {
    let cancelled = false;
    fetchPage(slug)
      .then((page) => {
        if (!cancelled) setLoaded({ slug, state: page ? { status: 'ready', page } : { status: 'missing' } });
      })
      .catch((err) => {
        console.error(err);
        if (!cancelled) setLoaded({ slug, state: { status: 'error', error: describeLoadError(err) } });
      });
    return () => {
      cancelled = true;
    };
  }, [slug, attempt]);

  const retry = () => {
    setLoaded(null);
    setAttempt((n) => n + 1);
  };

  const state: State = loaded && loaded.slug === slug ? loaded.state : { status: 'loading' };

  if (state.status === 'loading') {
    return (
      <Page title="">
        <LoadingBlock label="Loading page" />
      </Page>
    );
  }

  if (state.status === 'error') {
    return (
      <Page title="">
        <LoadFailed error={state.error} onRetry={retry} />
      </Page>
    );
  }

  if (state.status === 'missing') {
    return (
      <Page title="Page not found">
        <p className="text-text-muted">
          There is no page here. <Link to="/">Back to the calendar</Link>
        </p>
      </Page>
    );
  }

  const { page } = state;
  return (
    <Page title={page.title}>
      <div className="space-y-4">
        <Markdown>{page.body}</Markdown>
        {(page.embed || page.embedHtml) && <EmbedFrame url={page.embed} html={page.embedHtml} title={page.title} />}
      </div>
    </Page>
  );
}
