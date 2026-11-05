import { describe, it, expect } from 'vitest'
import { getTagColor } from './index'

describe('getTagColor', () => {
  const cases: [string, string][] = [
    ['music',      'tag tag-music'],
    ['comedy',     'tag tag-comedy'],
    ['community',  'tag tag-community'],
    ['literature', 'tag tag-literature'],
    ['diy',        'tag tag-diy'],
    ['free',       'tag tag-free'],
    ['recurring',  'tag tag-recurring'],
    ['art',        'tag tag-art'],
    ['sports',     'tag tag-sports'],
    ['film',       'tag tag-film'],
    ['cancelled',  'tag tag-cancelled'],
  ]

  it.each(cases)('returns correct classes for "%s"', (tag, expected) => {
    expect(getTagColor(tag)).toBe(expected)
  })

  it('is case-insensitive', () => {
    expect(getTagColor('Music')).toBe('tag tag-music')
    expect(getTagColor('COMEDY')).toBe('tag tag-comedy')
  })

  it('returns fallback classes for unknown tag', () => {
    expect(getTagColor('unknown')).toBe('tag tag-recurring')
  })

  it('returns fallback for empty string', () => {
    expect(getTagColor('')).toBe('tag tag-recurring')
  })
})
