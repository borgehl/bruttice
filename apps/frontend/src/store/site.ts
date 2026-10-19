import { create } from 'zustand';
import { fetchSiteConfig } from '../services/api';

// Settings that belong to one deployment of the site rather than to the code: its name, contact
// address and which optional pages it has. The server supplies them, and these defaults apply
// until it answers (or if it can't).
export interface SiteConfig {
  name: string;
  region: string;
  contactEmail: string;
  description: string;
  features: {
    popBooker: boolean;
  };
  // the calendar the Pop! Booker page embeds, empty if there is none
  popBookerUrl: string;
}

export const defaultSiteConfig: SiteConfig = {
  name: 'Bruttice',
  region: '',
  contactEmail: '',
  description: 'A community calendar for local events. A simple aggregate of events, with no algorithms.',
  features: { popBooker: false },
  popBookerUrl: '',
};

interface SiteState {
  site: SiteConfig;
  // false until the server has answered (or failed to), so a page that depends on a setting can wait
  loaded: boolean;
  load: () => Promise<void>;
}

export function mergeSiteConfig(partial: Partial<SiteConfig> | null | undefined): SiteConfig {
  return {
    ...defaultSiteConfig,
    ...partial,
    features: { ...defaultSiteConfig.features, ...partial?.features },
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
