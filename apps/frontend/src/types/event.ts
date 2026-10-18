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
  tags: string[];
}

export interface EventDay {
  day: string;
  events: Event[];
}

export interface Venue {
  name: string;
  slug: string;
  link: string;
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
