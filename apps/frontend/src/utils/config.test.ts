import { describe, it, expect } from 'vitest'
import { changeNotice, count, regionUse } from './config'

describe('count', () => {
  it('uses the plural except for one', () => {
    expect(count(0, 'event')).toBe('0 events')
    expect(count(1, 'event')).toBe('1 event')
    expect(count(2, 'venue')).toBe('2 venues')
  })
})

describe('changeNotice', () => {
  it('says what happened to the events', () => {
    expect(changeNotice({ eventsMoved: 3, eventsDeleted: 0 }, 'Tulsa')).toBe('Moved 3 events to Tulsa.')
    expect(changeNotice({ eventsMoved: 1, eventsDeleted: 0 })).toBe('Moved 1 event.')
    expect(changeNotice({ eventsMoved: 0, eventsDeleted: 2 })).toBe('Deleted 2 events.')
  })

  it('is empty when no event was touched', () => {
    expect(changeNotice({ eventsMoved: 0, eventsDeleted: 0 })).toBe('')
  })
})

describe('regionUse', () => {
  it('counts events and venues', () => {
    expect(regionUse(1, 2)).toBe('1 event · 2 venues')
  })
})
