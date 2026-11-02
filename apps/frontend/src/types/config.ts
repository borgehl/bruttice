// The lists an admin manages in the settings: what GET /api/admin/config returns.

export interface ConfigRegion {
  id: number;
  name: string;
  events: number;
  venues: number;
  retired: boolean;
  // where a retired region's events and old name went
  mergedInto?: string;
}

export interface ConfigTag {
  id: number;
  name: string;
  events: number;
}

export interface ConfigAge {
  id: number;
  name: string;
  value: number;
  events: number;
}

export interface ConfigVenue {
  id: number;
  name: string;
  link: string;
  region: string;
  // kept up to date by its scraper, so not editable here
  scraped: boolean;
  events: number;
}

export interface AdminConfig {
  regions: ConfigRegion[];
  tags: ConfigTag[];
  ages: ConfigAge[];
  venues: ConfigVenue[];
}

// what a change did to events
export interface ConfigChange {
  eventsMoved: number;
  eventsDeleted: number;
}
