import { create } from 'zustand';
import { fetchSiteConfig } from '../services/api';

// Settings that belong to one deployment of the site rather than to the code: its name, region,
// description, and whether it has a contact form. The server supplies them, and these defaults apply
// until it answers (or if it can't).
export interface SiteConfig {
  name: string;
  region: string;
  // true if messages from the contact form can be delivered
  contactForm: boolean;
  description: string;
}

// The part of a document that pageSiteName reads, so a test can pass a stand-in.
interface MetaReader {
  querySelector(selector: string): { getAttribute(name: string): string | null } | null;
}

// The site's name as the page was served. The production image puts it in the og:site_name tag
// (docker-site-meta.sh), so the first paint can show it before /api/site answers. Empty if the page
// has none, as in development.
export function pageSiteName(doc: MetaReader | undefined): string {
  return doc?.querySelector('meta[property="og:site_name"]')?.getAttribute('content')?.trim() ?? '';
}

export const defaultSiteConfig: SiteConfig = {
  name: pageSiteName(typeof document === 'undefined' ? undefined : document) || 'Bruttice',
  region: '',
  contactForm: false,
  description: 'A community calendar for local events. A simple aggregate of events, with no algorithms.',
};

interface SiteState {
  site: SiteConfig;
  // false until the server has answered (or failed to), so a page that depends on a setting can wait
  loaded: boolean;
  load: () => Promise<void>;
}

// A setting the server leaves empty keeps its default, so a config without a `site` block still
// gives the pages a name.
export function mergeSiteConfig(partial: Partial<SiteConfig> | null | undefined): SiteConfig {
  return {
    name: partial?.name || defaultSiteConfig.name,
    region: partial?.region || defaultSiteConfig.region,
    description: partial?.description || defaultSiteConfig.description,
    contactForm: partial?.contactForm === true,
  };
}

export const useSiteStore = create<SiteState>((set) => ({
  site: defaultSiteConfig,
  loaded: false,
  load: async () => {
    try {
      set({ site: mergeSiteConfig(await fetchSiteConfig()), loaded: true });
    } catch {
      // keep the defaults
      set({ loaded: true });
    }
  },
}));
