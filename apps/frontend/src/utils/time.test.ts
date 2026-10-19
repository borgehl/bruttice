import { describe, it, expect } from 'vitest'
import { dateKey, dayParts, formatClock, formatDoors, groupEventsByDay } from './time'
import type { Event } from '../types/event'

const tz = 'America/Chicago'

describe('time formatting', () => {
  it('formats the clock', () => {
    expect(formatClock('2026-10-22T19:00:00-05:00', tz)).toBe('7:00 PM')
    expect(formatClock('2026-10-22T11:30:00-05:00', tz)).toBe('11:30 AM')
  })

  it('leaves the period off doors that are in the same half of the day', () => {
    expect(formatDoors('2026-10-22T18:30:00-05:00', '2026-10-22T19:00:00-05:00', tz)).toBe('Doors 6:30')
  })

  it('adds the period when doors are in a different half of the day', () => {
    expect(formatDoors('2026-10-22T11:30:00-05:00', '2026-10-22T13:00:00-05:00', tz)).toBe('Doors 11:30 AM')
  })

  it('splits a day into weekday and date', () => {
    expect(dayParts('2026-10-22T19:00:00-05:00', tz)).toEqual({ weekday: 'Thursday', date: 'Oct 22' })
  })

  it('keys a day by the local date', () => {
    expect(dateKey('2026-10-22T23:30:00-05:00', tz)).toBe('2026-10-22')
    expect(dateKey('2026-10-23T00:30:00-05:00', tz)).toBe('2026-10-23')
  })
})

describe('groupEventsByDay', () => {
  const event = (guid: string, startAt: string): Event => ({
    guid, title: guid, link: '', venue: '', venueLink: '', age: '', region: '',
    startAt, coverCost: '', tags: [],
  })

  it('groups by day and fills the day parts', () => {
    const days = groupEventsByDay([
      event('a', '2026-10-22T19:00:00Z'),
      event('b', '2026-10-22T21:00:00Z'),
      event('c', '2026-10-24T19:00:00Z'),
    ])
    expect(days).toHaveLength(2)
    expect(days[0].events.map((e) => e.guid)).toEqual(['a', 'b'])
    expect(days[0].weekday).toBeTruthy()
    expect(days[0].dateKey).toMatch(/^\d{4}-\d{2}-\d{2}$/)
  })
})
