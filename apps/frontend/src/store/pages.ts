import { create } from 'zustand';
import { fetchPages } from '../services/api';
import type { PageSummary } from '../types/event';

// Pages the site already has a place for. A page with one of these slugs replaces that page's text
// and is not listed in the navigation a second time.
export const BUILT_IN_SLUGS = ['about', 'privacy-policy'];

// The pages to list in the navigation, in the order the deployment gave them.
export function navPages(pages: PageSummary[]): PageSummary[] {
  return pages.filter((page) => page.nav && !BUILT_IN_SLUGS.includes(page.slug));
}

export function hasPage(pages: PageSummary[], slug: string): boolean {
  return pages.some((page) => page.slug === slug);
}

interface PagesState {
  pages: PageSummary[];
  // false until the server has answered (or failed to)
  loaded: boolean;
  load: () => Promise<void>;
}

export const usePagesStore = create<PagesState>((set) => ({
  pages: [],
  loaded: false,
  load: async () => {
    try {
      set({ pages: await fetchPages(), loaded: true });
    } catch {
      // no extra pages
      set({ loaded: true });
    }
  },
}));
