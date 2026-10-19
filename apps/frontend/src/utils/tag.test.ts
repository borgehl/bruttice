import { describe, it, expect } from 'vitest'
import { displayTag } from './tag'

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
