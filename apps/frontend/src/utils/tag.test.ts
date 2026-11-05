import { describe, it, expect } from 'vitest'
import { displayTag, tagsToShow } from './tag'

describe('displayTag', () => {
  it('capitalizes the first letter', () => {
    expect(displayTag('music')).toBe('Music')
    expect(displayTag('recurring')).toBe('Recurring')
  })

  it('writes DIY in capitals', () => {
    expect(displayTag('diy')).toBe('DIY')
    expect(displayTag('DIY')).toBe('DIY')
  })

  it('leaves an empty name alone', () => {
    expect(displayTag('')).toBe('')
  })
})

describe('tagsToShow', () => {
  it('leaves the tags alone for an event that is not cancelled', () => {
    expect(tagsToShow(['music', 'free'])).toEqual(['music', 'free'])
  })

  it('puts a cancelled mark first', () => {
    expect(tagsToShow(['music'], '2030-01-10T19:00:00Z')).toEqual(['cancelled', 'music'])
    expect(tagsToShow([], '2030-01-10T19:00:00Z')).toEqual(['cancelled'])
  })

  it('does not show the mark twice, or for an event that is not cancelled', () => {
    expect(tagsToShow(['cancelled', 'music'], '2030-01-10T19:00:00Z')).toEqual(['cancelled', 'music'])
    expect(tagsToShow(['Cancelled', 'music'])).toEqual(['music'])
  })
})
