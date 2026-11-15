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
  // when the event was found to be cancelled, absent if it is not
  cancelledAt?: string;
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
  cancelledAt?: string;
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
  // inbox conversations with a message nobody has read
  unreadMessages: number;
}

export interface InboxMessage {
  id: number;
  // who wrote it: the contact form, an email to the site, or an admin from the inbox (an answer or
  // a first message to someone)
  source: 'contact_form' | 'email' | 'admin_reply';
  fromName: string;
  fromEmail: string;
  // plain text
  body: string;
  // unix seconds
  createdAt: number;
  unread: boolean;
  // unix seconds Resend said an email an admin sent was delivered; 0 if it has not
  deliveredAt: number;
}

// A conversation, its messages oldest first.
export interface InboxThread {
  threadId: string;
  subject: string;
  // who it is with
  fromName: string;
  fromEmail: string;
  unread: boolean;
  archived: boolean;
  // unix seconds of the latest message
  lastAt: number;
  messages: InboxMessage[];
}

export interface IngestSource {
  source: string;
  status: 'ok' | 'skipped' | 'errors' | 'failed';
  scrapedCount: number;
  savedCount: number;
  failedCount: number;
  // events not saved because they were deleted on the server
  skippedCount: number;
  errors: string[];
  // one description per skipped event, these are not errors
  skipped: string[];
}

export interface IngestRun {
  id: number;
  startedAt: number;
  finishedAt: number;
  dedupeError?: string;
  scraped: number;
  saved: number;
  failed: number;
  skipped: number;
  sources: IngestSource[];
}

export interface VisitorStats {
  days: number;
  daily: { date: string; visitors: number }[];
  locations: { country: string; state: string; city: string; visitors: number }[];
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
