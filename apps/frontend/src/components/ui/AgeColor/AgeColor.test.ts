import { describe, it, expect } from 'vitest'
import { getAgeColor } from './index'

describe('getAgeColor', () => {
  it('returns green for all-ages', () => {
    expect(getAgeColor('all-ages')).toBe('text-green-400')
  })

  it('returns red for 21+', () => {
    expect(getAgeColor('21+')).toBe('text-red-400')
  })

  it('returns orange for 18+', () => {
    expect(getAgeColor('18+')).toBe('text-orange-400')
  })

  it('returns yellow for 16+', () => {
    expect(getAgeColor('16+')).toBe('text-yellow-400')
  })

  it('returns dim text for unknown age', () => {
    expect(getAgeColor('unknown')).toBe('text-text-dim')
  })

  it('returns dim text for empty string', () => {
    expect(getAgeColor('')).toBe('text-text-dim')
  })
})
