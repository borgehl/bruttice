import { useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { fetchPage } from '../services/api';
import type { SitePage } from '../types/event';
import Page from '../components/ui/Page';
import Markdown from '../components/ui/Markdown';
import EmbedFrame from '../components/ui/EmbedFrame';

type State = { status: 'loading' } | { status: 'missing' } | { status: 'error' } | { status: 'ready'; page: SitePage };

// A page the deployment supplies, at /p/<slug>.
export default function ContentPage() {
  const { slug = '' } = useParams();
  const [loaded, setLoaded] = useState<{ slug: string; state: State } | null>(null);

  useEffect(() => {
    let cancelled = false;
    fetchPage(slug)
      .then((page) => {
        if (!cancelled) setLoaded({ slug, state: page ? { status: 'ready', page } : { status: 'missing' } });
      })
      .catch(() => {
        if (!cancelled) setLoaded({ slug, state: { status: 'error' } });
      });
    return () => {
      cancelled = true;
    };
  }, [slug]);

  const state: State = loaded && loaded.slug === slug ? loaded.state : { status: 'loading' };

  if (state.status === 'loading') {
    return (
      <Page title="">
        <p className="py-12 text-center text-text-muted">Loading...</p>
      </Page>
    );
  }

  if (state.status !== 'ready') {
    return (
      <Page title={state.status === 'missing' ? 'Page not found' : 'Something went wrong'}>
        <p className="text-text-muted">
          {state.status === 'missing' ? 'There is no page here.' : 'The page could not be loaded.'}{' '}
          <Link to="/">Back to the calendar</Link>
        </p>
      </Page>
    );
  }

  const { page } = state;
  return (
    <Page title={page.title}>
      <div className="space-y-4">
        <Markdown>{page.body}</Markdown>
        {page.embed && <EmbedFrame url={page.embed} title={page.title} />}
      </div>
    </Page>
  );
}
