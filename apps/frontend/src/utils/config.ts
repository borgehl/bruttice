import type { ConfigChange } from '../types/config';

export function count(n: number, noun: string): string {
  return `${n} ${noun}${n === 1 ? '' : 's'}`;
}

// What to tell an admin a change did. Empty if no event was touched.
export function changeNotice(change: ConfigChange, movedTo?: string): string {
  if (change.eventsDeleted > 0) return `Deleted ${count(change.eventsDeleted, 'event')}.`;
  if (change.eventsMoved > 0) return `Moved ${count(change.eventsMoved, 'event')}${movedTo ? ` to ${movedTo}` : ''}.`;
  return '';
}

// How much of the site a region is: what retiring it touches.
export function regionUse(events: number, venues: number): string {
  return `${count(events, 'event')} · ${count(venues, 'venue')}`;
}
