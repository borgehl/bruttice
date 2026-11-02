import { describe, it, expect } from 'vitest'
import { embedHeightFrom } from './embedHeight'

describe('embedHeightFrom', () => {
  it('reads the height of an embed message, rounded up', () => {
    expect(embedHeightFrom({ type: 'bruttice-embed-height', height: 512.2 })).toBe(513)
  })

  it('keeps the height within bounds', () => {
    expect(embedHeightFrom({ type: 'bruttice-embed-height', height: 1 })).toBe(120)
    expect(embedHeightFrom({ type: 'bruttice-embed-height', height: 99999 })).toBe(2400)
  })

  it('ignores anything else', () => {
    expect(embedHeightFrom(null)).toBeNull()
    expect(embedHeightFrom('bruttice-embed-height')).toBeNull()
    expect(embedHeightFrom({ type: 'other', height: 300 })).toBeNull()
    expect(embedHeightFrom({ type: 'bruttice-embed-height', height: '300' })).toBeNull()
    expect(embedHeightFrom({ type: 'bruttice-embed-height', height: NaN })).toBeNull()
  })
})
