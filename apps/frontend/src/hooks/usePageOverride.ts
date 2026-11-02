import { useEffect, useState } from 'react';
import { fetchPage } from '../services/api';
import { hasPage, usePagesStore } from '../store/pages';
import type { SitePage } from '../types/event';

// The deployment's own page for a built-in one (about, privacy-policy), if it has one.
// `ready` is false until it is known whether there is one, so the built-in text doesn't flash first.
export function usePageOverride(slug: string): { page: SitePage | null; ready: boolean } {
  const { pages, loaded } = usePagesStore();
  const exists = hasPage(pages, slug);
  const [page, setPage] = useState<SitePage | null>(null);
  const [fetched, setFetched] = useState(false);

  useEffect(() => {
    if (!loaded || !exists) return;
    let cancelled = false;
    fetchPage(slug)
      .then((found) => {
        if (!cancelled) setPage(found);
      })
      .catch(() => {
        // fall back to the built-in text
      })
      .finally(() => {
        if (!cancelled) setFetched(true);
      });
    return () => {
      cancelled = true;
    };
  }, [slug, loaded, exists]);

  return { page: exists ? page : null, ready: loaded && (!exists || fetched) };
}
