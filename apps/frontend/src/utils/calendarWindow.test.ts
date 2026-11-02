import { describe, expect, it } from 'vitest';
import type { Event } from '../types/event';
import { addDays, formatLocalDate, mergeEvents, windowAfter, windowBefore } from './calendarWindow';

function ev(guid: string): Event {
  return { guid, title: guid, link: '', venue: '', venueLink: '', age: '', region: '', startAt: '', coverCost: '', tags: [] };
}

describe('addDays', () => {
  it('crosses month and year boundaries', () => {
    expect(addDays('2026-01-31', 1)).toBe('2026-02-01');
    expect(addDays('2026-12-31', 1)).toBe('2027-01-01');
    expect(addDays('2026-03-01', -1)).toBe('2026-02-28');
  });

  it('is exact across DST changes', () => {
    expect(addDays('2026-03-07', 2)).toBe('2026-03-09');
    expect(addDays('2026-11-01', -1)).toBe('2026-10-31');
    expect(addDays('2026-01-01', 365)).toBe('2027-01-01');
  });
});

describe('formatLocalDate', () => {
  it('uses the local date, not UTC', () => {
    expect(formatLocalDate(new Date(2026, 9, 7, 23, 30))).toBe('2026-10-07');
    expect(formatLocalDate(new Date(2026, 0, 5, 0, 5))).toBe('2026-01-05');
  });
});

describe('windows', () => {
  it('windowAfter starts the day after and does not overlap', () => {
    expect(windowAfter('2026-10-06')).toEqual({ from: '2026-10-07', to: '2026-11-05' });
  });

  it('windowBefore ends the day before the first loaded day', () => {
    expect(windowBefore('2026-10-07')).toEqual({ from: '2026-09-07', to: '2026-10-06' });
  });

  it('adjacent windows are contiguous', () => {
    const first = windowAfter(addDays('2026-10-07', -1));
    const back = windowBefore(first.from);
    expect(addDays(back.to, 1)).toBe(first.from);
    expect(windowAfter(first.to).from).toBe(addDays(first.to, 1));
  });
});

describe('mergeEvents', () => {
  it('prepends and appends in order', () => {
    const base = [ev('b')];
    expect(mergeEvents(base, [ev('a')], 'prepend').map((e) => e.guid)).toEqual(['a', 'b']);
    expect(mergeEvents(base, [ev('c')], 'append').map((e) => e.guid)).toEqual(['b', 'c']);
  });

  it('drops duplicates by guid', () => {
    expect(mergeEvents([ev('a'), ev('b')], [ev('b'), ev('c')], 'append').map((e) => e.guid)).toEqual(['a', 'b', 'c']);
  });
});
