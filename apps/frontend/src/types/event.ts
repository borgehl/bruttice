export interface Event {
  guid: string;
  title: string;
  link: string;
  venue: string;
  venueLink: string;
  age: string;
  region: string;
  startAt: string;
  doorsAt?: string;
  coverCost: string;
  flyerUrl?: string;
  tags: string[];
}

export interface EventDay {
  day: string;
  // "2026-10-22", "Thursday" and "Oct 22": the day split up for headers
  dateKey: string;
  weekday: string;
  date: string;
  events: Event[];
}

export interface Venue {
  name: string;
  slug: string;
  link: string;
  // the venue's region name, empty if it has none
  region: string;
}

export interface Tag {
  id: number;
  name: string;
}

export interface Age {
  id: number;
  name: string;
  value: number;
}

export interface Region {
  id: number;
  name: string;
}

export interface FilterOptions {
  tags: Tag[];
  venues: Venue[];
  ages: Age[];
  regions: Region[];
}

export interface CalendarFilter {
  tags: number[];
  tagsInclude: boolean;
  ages: number[];
  agesLessThan: boolean;
  venues: string[];
  venuesInclude: boolean;
  fromDate?: string;
  toDate?: string;
}

// Admin types
export interface AdminEvent {
  guid: string;
  title: string;
  link: string;
  venue: string;
  venueLink: string;
  age: string;
  region: string;
  date: string;
  time: string;
  startTime: string;
  doorsTime?: string;
  coverCost: string;
  flyerUrl?: string;
  tags: string[];
  email: string;
  score: number;
}

export interface AdminEventDay {
  day: string;
  events: AdminEvent[];
}

export interface AdminDashboard {
  email: string;
  awaitingApproval: AdminEventDay[];
  upcoming: AdminEventDay[];
}

export interface IngestSource {
  source: string;
  status: 'ok' | 'errors' | 'failed';
  scrapedCount: number;
  savedCount: number;
  failedCount: number;
  errors: string[];
}

export interface IngestRun {
  id: number;
  startedAt: number;
  finishedAt: number;
  dedupeError?: string;
  scraped: number;
  saved: number;
  failed: number;
  sources: IngestSource[];
}

export interface IngestRunsPage {
  runs: IngestRun[];
  hasMore: boolean;
}

// A page of the deployment's own, from its pages directory.
export interface PageSummary {
  slug: string;
  title: string;
  // listed in the navigation
  nav: boolean;
  order: number;
}

export interface SitePage extends PageSummary {
  // Markdown
  body: string;
  // an http(s) page shown in a frame under the body
  embed?: string;
  // a complete HTML document shown in a sandboxed frame under the body
  embedHtml?: string;
}
