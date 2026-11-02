import type { Event } from '../types/event';

export const WINDOW_DAYS = 30;

// Number of consecutive empty windows to fetch before deciding there is nothing more in that direction.
export const MAX_EMPTY_WINDOWS = 6;

export interface DateWindow {
  from: string;
  to: string;
}

// Local calendar date as YYYY-MM-DD. toISOString() would give the UTC date, which is
// tomorrow's date in the evening for US timezones.
export function formatLocalDate(date: Date): string {
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, '0');
  const d = String(date.getDate()).padStart(2, '0');
  return `${y}-${m}-${d}`;
}

export function todayLocal(): string {
  return formatLocalDate(new Date());
}

export function addDays(dateStr: string, days: number): string {
  const [y, m, d] = dateStr.split('-').map(Number);
  return formatLocalDate(new Date(y, m - 1, d + days));
}

// The window of WINDOW_DAYS starting the day after `lastLoaded`.
export function windowAfter(lastLoaded: string): DateWindow {
  return { from: addDays(lastLoaded, 1), to: addDays(lastLoaded, WINDOW_DAYS) };
}

// The window of WINDOW_DAYS ending the day before `firstLoaded`.
export function windowBefore(firstLoaded: string): DateWindow {
  return { from: addDays(firstLoaded, -WINDOW_DAYS), to: addDays(firstLoaded, -1) };
}

// Merges a freshly fetched window into the list, dropping duplicates by guid.
export function mergeEvents(existing: Event[], incoming: Event[], position: 'prepend' | 'append'): Event[] {
  const seen = new Set(existing.map((e) => e.guid));
  const fresh = incoming.filter((e) => !seen.has(e.guid));
  return position === 'prepend' ? [...fresh, ...existing] : [...existing, ...fresh];
}
