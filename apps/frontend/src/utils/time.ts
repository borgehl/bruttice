import type { Event, EventDay } from '../types/event';

const DEFAULT_TIMEZONE = 'America/Chicago';

// Mirrors the server's DisplayDateFormat but using the browser's local timezone.
// e.g. "Monday Jan. 20, 2025" or "Thursday May 1, 2025" (no period for May)
export function formatEventDay(startAt: string): string {
  const tz = Intl.DateTimeFormat().resolvedOptions().timeZone || DEFAULT_TIMEZONE;
  const date = new Date(startAt);

  const weekday = new Intl.DateTimeFormat('en-US', { weekday: 'long', timeZone: tz }).format(date);
  const monthShort = new Intl.DateTimeFormat('en-US', { month: 'short', timeZone: tz }).format(date);
  const monthLong = new Intl.DateTimeFormat('en-US', { month: 'long', timeZone: tz }).format(date);
  const day = new Intl.DateTimeFormat('en-US', { day: 'numeric', timeZone: tz }).format(date);
  const year = new Intl.DateTimeFormat('en-US', { year: 'numeric', timeZone: tz }).format(date);

  const period = monthShort !== monthLong ? '.' : '';
  return `${weekday} ${monthShort}${period} ${day}, ${year}`;
}

function formatTime(date: Date, timezone: string): string {
  const parts = new Intl.DateTimeFormat('en-US', {
    hour: 'numeric',
    minute: '2-digit',
    hour12: true,
    timeZone: timezone,
  }).formatToParts(date);

  const hour = parts.find((p) => p.type === 'hour')?.value ?? '';
  const minute = parts.find((p) => p.type === 'minute')?.value ?? '';
  const dayPeriod = parts.find((p) => p.type === 'dayPeriod')?.value?.toLowerCase() ?? '';

  return `${hour}:${minute}${dayPeriod}`;
}

export function groupEventsByDay(events: Event[]): EventDay[] {
  const groups: EventDay[] = [];
  for (const event of events) {
    const dayLabel = formatEventDay(event.startAt);
    const last = groups[groups.length - 1];
    if (!last || last.day !== dayLabel) {
      groups.push({ day: dayLabel, events: [event] });
    } else {
      last.events.push(event);
    }
  }
  return groups;
}

export function formatEventTime(startAt: string, doorsAt?: string): string {
  const tz = Intl.DateTimeFormat().resolvedOptions().timeZone || DEFAULT_TIMEZONE;
  const startDate = new Date(startAt);

  if (doorsAt) {
    return `Doors ${formatTime(new Date(doorsAt), tz)}, Start ${formatTime(startDate, tz)}`;
  }

  return formatTime(startDate, tz);
}
