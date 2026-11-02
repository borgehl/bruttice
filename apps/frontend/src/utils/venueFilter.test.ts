import { describe, it, expect } from 'vitest'
import { dropHiddenVenues, toggle, venueHint, visibleVenues } from './venueFilter'
import type { Venue } from '../types/event'

const venue = (name: string, region: string): Venue => ({ name, slug: name.toLowerCase(), link: '', region })

const venues = [
  venue('The Hunt Club', 'Downtown'),
  venue('Guthrie Green', 'Downtown'),
  venue('Whittier Pub', 'Pearl District'),
  venue('Red Fork Room', 'Brookside'),
  venue('Unlisted', ''),
]

describe('visibleVenues', () => {
  it('shows every venue when no region is selected', () => {
    expect(visibleVenues(venues, [])).toHaveLength(5)
  })

  it('narrows to the selected regions', () => {
    expect(visibleVenues(venues, ['Downtown']).map((v) => v.name)).toEqual(['The Hunt Club', 'Guthrie Green'])
    expect(visibleVenues(venues, ['Downtown', 'Brookside'])).toHaveLength(3)
  })

  it('hides venues without a region once a region is selected', () => {
    expect(visibleVenues(venues, ['Downtown']).some((v) => v.name === 'Unlisted')).toBe(false)
  })

  it('searches the name, ignoring case and surrounding space', () => {
    expect(visibleVenues(venues, [], '  PUB ').map((v) => v.name)).toEqual(['Whittier Pub'])
  })

  it('applies the search inside the selected regions', () => {
    expect(visibleVenues(venues, ['Downtown'], 'pub')).toEqual([])
    expect(visibleVenues(venues, ['Downtown'], 'green').map((v) => v.name)).toEqual(['Guthrie Green'])
  })
})

describe('dropHiddenVenues', () => {
  it('keeps everything when no region is selected', () => {
    expect(dropHiddenVenues(['Whittier Pub', 'Unlisted'], venues, [])).toEqual(['Whittier Pub', 'Unlisted'])
  })

  it('drops selected venues that a region hides', () => {
    expect(dropHiddenVenues(['The Hunt Club', 'Whittier Pub'], venues, ['Downtown'])).toEqual(['The Hunt Club'])
  })

  it('drops a name that is not a known venue', () => {
    expect(dropHiddenVenues(['Gone'], venues, [])).toEqual([])
  })
})

describe('toggle', () => {
  it('adds and removes', () => {
    expect(toggle(['a'], 'b')).toEqual(['a', 'b'])
    expect(toggle(['a', 'b'], 'a')).toEqual(['b'])
  })
})

describe('venueHint', () => {
  it('names the regions', () => {
    expect(venueHint(['Downtown', 'Brookside'])).toBe('Showing venues in Downtown, Brookside')
  })

  it('asks for a region when there is none', () => {
    expect(venueHint([])).toBe('Pick a region to narrow this list')
  })
})
