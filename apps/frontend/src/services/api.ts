import { ApiError } from '../utils/loadError';
import type { SiteConfig } from '../store/site';
import type { Age, Event, FilterOptions, Venue, AdminDashboard, IngestRunsPage, InboxThread, PageSummary, SitePage } from '../types/event';

const API_BASE = '/api';

export interface CalendarParams {
  filterKey?: string;
  fromDate?: string;
  toDate?: string;
}

export async function fetchCalendarEvents(params?: CalendarParams): Promise<Event[]> {
  const searchParams = new URLSearchParams();

  if (params?.filterKey) {
    searchParams.set('filterKey', params.filterKey);
  }
  if (params?.fromDate) {
    searchParams.set('from-date', params.fromDate);
  }
  if (params?.toDate) {
    searchParams.set('to-date', params.toDate);
  }

  const url = `${API_BASE}/calendar${searchParams.toString() ? '?' + searchParams.toString() : ''}`;
  const response = await fetch(url);

  if (!response.ok) {
    throw new ApiError(response.status, 'Failed to fetch events');
  }

  return response.json();
}

export async function fetchFilterOptions(): Promise<FilterOptions> {
  const response = await fetch(`${API_BASE}/filter-options`);

  if (!response.ok) {
    throw new ApiError(response.status, 'Failed to fetch filter options');
  }

  return response.json();
}

export async function fetchVenues(): Promise<Venue[]> {
  const response = await fetch(`${API_BASE}/venues`);

  if (!response.ok) {
    throw new ApiError(response.status, 'Failed to fetch venues');
  }

  return response.json();
}

export async function fetchSubmissionVenues(): Promise<Venue[]> {
  const response = await fetch(`${API_BASE}/submission-venues`);

  if (!response.ok) {
    throw new ApiError(response.status, 'Failed to fetch venues');
  }

  return response.json();
}

export interface AddEventData {
  title: string;
  link: string;
  venue: string;
  region: string;
  age: string;
  date: string;
  startTime: string;
  doorsTime?: string;
  coverCost: string;
  tags: string[];
  email: string;
}

export async function submitEvent(data: AddEventData): Promise<{ success: boolean; message: string }> {
  const response = await fetch(`${API_BASE}/add-event`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data),
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.message || 'Failed to submit event');
  }

  return response.json();
}

export interface ContactData {
  name: string;
  email: string;
  message: string;
  // a field people never see, only a bot fills it in
  website: string;
}

export async function sendContactMessage(data: ContactData): Promise<void> {
  const response = await fetch(`${API_BASE}/contact`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data),
  });

  if (!response.ok) {
    // the server's reason is meant for the visitor: what to fix, or to wait
    const error = await response.json().catch(() => null);
    throw new ApiError(response.status, error?.error || 'Your message could not be sent');
  }
}

export interface CreateFilterData {
  tags: string[];
  tagsInclude: boolean;
  ages: number[];
  agesLessThan: boolean;
  venues: string[];
  venuesInclude: boolean;
}

export async function createFilter(data: CreateFilterData): Promise<{ filterKey: string }> {
  const response = await fetch(`${API_BASE}/filter`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data),
  });

  if (!response.ok) {
    throw new ApiError(response.status, 'Failed to create filter');
  }

  return response.json();
}

export interface FilterDetails {
  tags?: string[];
  tagsInclude?: boolean;
  age?: number;
  agesLessThan?: boolean;
  venues?: string[];
  venuesInclude?: boolean;
}

export async function fetchFilterDetails(filterKey: string): Promise<FilterDetails> {
  const response = await fetch(`${API_BASE}/filter?filterKey=${encodeURIComponent(filterKey)}`);

  if (!response.ok) {
    throw new ApiError(response.status, 'Failed to fetch filter details');
  }

  return response.json();
}

export async function postVisitorConsent(): Promise<void> {
  await fetch(`${API_BASE}/visitor/consent`, {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      path: window.location.pathname,
      query: window.location.search.replace(/^\?/, ''),
    }),
  });
}

// Admin API functions

export interface LoginCredentials {
  email: string;
  password: string;
}

export interface LoginResponse {
  success: boolean;
  email?: string;
  // how the admin is shown to other admins, empty if they have not set one
  name?: string;
  message?: string;
}

export async function adminLogin(credentials: LoginCredentials): Promise<LoginResponse> {
  const response = await fetch(`${API_BASE}/admin/login`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    credentials: 'include',
    body: JSON.stringify(credentials),
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Login failed');
  }

  return response.json();
}

// changes the signed in admin's own settings
export async function updateAdminSettings(settings: { name: string }): Promise<LoginResponse> {
  const response = await fetch(`${API_BASE}/admin/settings`, {
    method: 'PATCH',
    headers: {
      'Content-Type': 'application/json',
    },
    credentials: 'include',
    body: JSON.stringify(settings),
  });

  if (!response.ok) {
    if (response.status === 401) {
      throw new ApiError(response.status, 'Authentication required');
    }
    // the server's reason is meant for the admin, such as that the name is too long
    const error = await response.json().catch(() => null);
    throw new ApiError(response.status, error?.error || 'Failed to save the settings');
  }

  return response.json();
}

export async function adminLogout(): Promise<void> {
  const response = await fetch(`${API_BASE}/admin/logout`, {
    method: 'POST',
    credentials: 'include',
  });

  if (!response.ok) {
    throw new ApiError(response.status, 'Logout failed');
  }
}

export async function checkAdminSession(): Promise<LoginResponse> {
  const response = await fetch(`${API_BASE}/admin/session`, {
    credentials: 'include',
  });

  if (!response.ok) {
    return { success: false };
  }

  return response.json();
}

export async function fetchAdminDashboard(): Promise<AdminDashboard> {
  const response = await fetch(`${API_BASE}/admin/dashboard`, {
    credentials: 'include',
  });

  if (!response.ok) {
    if (response.status === 401) {
      throw new ApiError(response.status, 'Authentication required');
    }
    throw new ApiError(response.status, 'Failed to fetch dashboard');
  }

  return response.json();
}

// beforeId: only return runs older than this id
export async function fetchIngestRuns(beforeId?: number): Promise<IngestRunsPage> {
  const query = beforeId ? `?before=${beforeId}` : '';
  const response = await fetch(`${API_BASE}/admin/ingest-runs${query}`, {
    credentials: 'include',
  });

  if (!response.ok) {
    if (response.status === 401) {
      throw new ApiError(response.status, 'Authentication required');
    }
    throw new ApiError(response.status, 'Failed to fetch ingest runs');
  }

  return response.json();
}

export async function startIngest(): Promise<void> {
  const response = await fetch(`${API_BASE}/admin/ingest-run`, {
    method: 'POST',
    credentials: 'include',
  });

  if (!response.ok) {
    if (response.status === 401) {
      throw new ApiError(response.status, 'Authentication required');
    }
    if (response.status === 409) {
      throw new ApiError(response.status, 'An ingestion is already running');
    }
    throw new ApiError(response.status, 'Failed to start ingestion');
  }
}

export async function fetchInbox(includeArchived: boolean): Promise<InboxThread[]> {
  const response = await fetch(`${API_BASE}/admin/inbox${includeArchived ? '?archived=true' : ''}`, {
    credentials: 'include',
  });

  if (!response.ok) {
    if (response.status === 401) {
      throw new ApiError(response.status, 'Authentication required');
    }
    throw new ApiError(response.status, 'Failed to fetch the inbox');
  }

  return response.json();
}

// the action on a conversation: read, archive or unarchive
export async function updateInboxThread(threadId: string, action: 'read' | 'archive' | 'unarchive'): Promise<void> {
  const response = await fetch(`${API_BASE}/admin/inbox/${encodeURIComponent(threadId)}/${action}`, {
    method: 'POST',
    credentials: 'include',
  });

  if (!response.ok) {
    if (response.status === 401) {
      throw new ApiError(response.status, 'Authentication required');
    }
    throw new ApiError(response.status, 'Failed to update the conversation');
  }
}

// emails the person the conversation is with, and keeps the reply in it
export async function replyToInboxThread(threadId: string, body: string): Promise<void> {
  const response = await fetch(`${API_BASE}/admin/inbox/${encodeURIComponent(threadId)}/reply`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    credentials: 'include',
    body: JSON.stringify({ body }),
  });

  if (!response.ok) {
    if (response.status === 401) {
      throw new ApiError(response.status, 'Authentication required');
    }
    // the server's reason is meant for the admin: what to fix, or that email is not set up
    const error = await response.json().catch(() => null);
    throw new ApiError(response.status, error?.error || 'The reply could not be sent');
  }
}

export interface VoteData {
  guid: string;
  vote: 'Approve' | 'Deny' | 'Keep' | 'Remove';
}

export async function submitVote(data: VoteData): Promise<{ success: boolean }> {
  const response = await fetch(`${API_BASE}/admin/vote`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    credentials: 'include',
    body: JSON.stringify(data),
  });

  if (!response.ok) {
    throw new ApiError(response.status, 'Failed to submit vote');
  }

  return response.json();
}

export interface AdminAddEventData {
  title: string;
  link: string;
  venue: string;
  region: string;
  age: string;
  date: string;
  startTime: string;
  doorsTime?: string;
  coverCost: string;
  tags: string[];
}

export async function adminAddEvent(data: AdminAddEventData): Promise<{ success: boolean; message: string }> {
  const response = await fetch(`${API_BASE}/admin/add-event`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    credentials: 'include',
    body: JSON.stringify(data),
  });

  if (!response.ok) {
    if (response.status === 401) {
      throw new ApiError(response.status, 'Authentication required');
    }
    const error = await response.json();
    throw new Error(error.error || 'Failed to add event');
  }

  return response.json();
}

export interface AdminUpdateEventData {
  guid: string;
  title?: string;
  link?: string;
  coverCost?: string;
  // an empty string removes the flyer
  flyerUrl?: string;
  venue?: string;
  region?: string;
  age?: string;
  date?: string;
  startTime?: string;
  doorsTime?: string;
  tags?: string[];
}

export async function fetchAdminAges(): Promise<Age[]> {
  const response = await fetch(`${API_BASE}/admin/ages`, {
    credentials: 'include',
  });

  if (!response.ok) {
    throw new ApiError(response.status, 'Failed to fetch ages');
  }

  return response.json();
}

export async function adminUpdateEvent(data: AdminUpdateEventData): Promise<{ success: boolean; message: string }> {
  const response = await fetch(`${API_BASE}/admin/event`, {
    method: 'PATCH',
    headers: {
      'Content-Type': 'application/json',
    },
    credentials: 'include',
    body: JSON.stringify(data),
  });

  if (!response.ok) {
    if (response.status === 401) {
      throw new ApiError(response.status, 'Authentication required');
    }
    const error = await response.json();
    throw new Error(error.error || 'Failed to update event');
  }

  return response.json();
}

export async function verifyEvent(verifyKey: string): Promise<{ success: boolean; message: string }> {
  const response = await fetch(`${API_BASE}/verify-event?verifyKey=${encodeURIComponent(verifyKey)}`);

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to verify event');
  }

  return response.json();
}

export async function fetchSiteConfig(): Promise<Partial<SiteConfig>> {
  const response = await fetch(`${API_BASE}/site`);

  if (!response.ok) {
    throw new ApiError(response.status, 'Failed to fetch site config');
  }

  return response.json();
}

export async function fetchPages(): Promise<PageSummary[]> {
  const response = await fetch(`${API_BASE}/pages`);

  if (!response.ok) {
    throw new ApiError(response.status, 'Failed to fetch pages');
  }

  return response.json();
}

// null if the deployment has no page with this slug
export async function fetchPage(slug: string): Promise<SitePage | null> {
  const response = await fetch(`${API_BASE}/pages/${encodeURIComponent(slug)}`);

  if (response.status === 404) return null;
  if (!response.ok) {
    throw new ApiError(response.status, 'Failed to fetch page');
  }

  return response.json();
}
