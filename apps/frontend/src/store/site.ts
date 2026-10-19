import { create } from 'zustand';
import { fetchSiteConfig } from '../services/api';

// Settings that belong to one deployment of the site rather than to the code: its name, contact
// address and description. The server supplies them, and these defaults apply
// until it answers (or if it can't).
export interface SiteConfig {
  name: string;
  region: string;
  contactEmail: string;
  description: string;
}

export const defaultSiteConfig: SiteConfig = {
  name: 'Bruttice',
  region: '',
  contactEmail: '',
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
  const site = { ...defaultSiteConfig };
  for (const key of Object.keys(site) as (keyof SiteConfig)[]) {
    const value = partial?.[key];
    if (value) site[key] = value;
  }
  return site;
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
