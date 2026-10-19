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

function browserTimezone(): string {
  return Intl.DateTimeFormat().resolvedOptions().timeZone || DEFAULT_TIMEZONE;
}

// The calendar day an event falls on in the given timezone, as "2026-10-22".
export function dateKey(startAt: string, tz: string = browserTimezone()): string {
  return new Intl.DateTimeFormat('en-CA', { year: 'numeric', month: '2-digit', day: '2-digit', timeZone: tz }).format(
    new Date(startAt),
  );
}

// The pieces of a day header: "Thursday" and "Oct 22". The theme decides how they are laid out.
export function dayParts(startAt: string, tz: string = browserTimezone()): { weekday: string; date: string } {
  const date = new Date(startAt);
  return {
    weekday: new Intl.DateTimeFormat('en-US', { weekday: 'long', timeZone: tz }).format(date),
    date: new Intl.DateTimeFormat('en-US', { month: 'short', day: 'numeric', timeZone: tz }).format(date),
  };
}

function clockParts(date: Date, tz: string): { time: string; period: string } {
  const parts = new Intl.DateTimeFormat('en-US', {
    hour: 'numeric',
    minute: '2-digit',
    hour12: true,
    timeZone: tz,
  }).formatToParts(date);
  const hour = parts.find((p) => p.type === 'hour')?.value ?? '';
  const minute = parts.find((p) => p.type === 'minute')?.value ?? '';
  const period = parts.find((p) => p.type === 'dayPeriod')?.value?.toUpperCase() ?? '';
  return { time: `${hour}:${minute}`, period };
}

// "7:00 PM"
export function formatClock(startAt: string, tz: string = browserTimezone()): string {
  const { time, period } = clockParts(new Date(startAt), tz);
  return `${time} ${period}`;
}

// "Doors 6:30". The period is only added when doors open in a different half of the day than the
// show starts, so "Doors 11:30 AM" is not mistaken for the evening.
export function formatDoors(doorsAt: string, startAt: string, tz: string = browserTimezone()): string {
  const doors = clockParts(new Date(doorsAt), tz);
  const start = clockParts(new Date(startAt), tz);
  return doors.period === start.period ? `Doors ${doors.time}` : `Doors ${doors.time} ${doors.period}`;
}

export function groupEventsByDay(events: Event[]): EventDay[] {
  const groups: EventDay[] = [];
  for (const event of events) {
    const dayLabel = formatEventDay(event.startAt);
    const last = groups[groups.length - 1];
    if (!last || last.day !== dayLabel) {
      groups.push({ day: dayLabel, dateKey: dateKey(event.startAt), ...dayParts(event.startAt), events: [event] });
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
