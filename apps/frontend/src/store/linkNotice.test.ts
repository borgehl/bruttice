import { describe, it, expect } from 'vitest'
import { saveSkipLinkNotice, shouldSkipLinkNotice } from './linkNotice'

describe('shouldSkipLinkNotice', () => {
  it('is false by default', () => {
    expect(shouldSkipLinkNotice({ getItem: () => null })).toBe(false)
  })

  it('is true once saved', () => {
    expect(shouldSkipLinkNotice({ getItem: (k) => (k === 'skipLinkNotice' ? '1' : null) })).toBe(true)
  })

  it('ignores other values', () => {
    expect(shouldSkipLinkNotice({ getItem: () => '0' })).toBe(false)
    expect(shouldSkipLinkNotice({ getItem: () => 'true' })).toBe(false)
  })

  it('is false when storage throws or is missing', () => {
    expect(shouldSkipLinkNotice({ getItem: () => { throw new Error('blocked') } })).toBe(false)
    expect(shouldSkipLinkNotice(undefined)).toBe(false)
  })
})

describe('saveSkipLinkNotice', () => {
  it('saves under the key', () => {
    const saved: Record<string, string> = {}
    saveSkipLinkNotice({ setItem: (k, v) => { saved[k] = v } })
    expect(saved).toEqual({ skipLinkNotice: '1' })
  })

  it('does not throw when storage does', () => {
    expect(() => saveSkipLinkNotice({ setItem: () => { throw new Error('full') } })).not.toThrow()
  })
})
