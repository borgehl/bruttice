import { describe, it, expect } from 'vitest'
import { getAgeColor } from './index'

describe('getAgeColor', () => {
  it('returns the all-ages color', () => {
    expect(getAgeColor('all-ages')).toBe('text-age-all')
  })

  it('returns the 21+ color', () => {
    expect(getAgeColor('21+')).toBe('text-age-21')
  })

  it('returns the 18+ color', () => {
    expect(getAgeColor('18+')).toBe('text-age-18')
  })

  it('returns the 16+ color', () => {
    expect(getAgeColor('16+')).toBe('text-age-16')
  })

  it('returns faint text for unknown age', () => {
    expect(getAgeColor('unknown')).toBe('text-text-faint')
  })

  it('returns faint text for empty string', () => {
    expect(getAgeColor('')).toBe('text-text-faint')
  })
})
