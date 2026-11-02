import { describe, it, expect } from 'vitest'
import { parseTheme, readTheme, writeTheme } from './theme'

describe('parseTheme', () => {
  it('accepts neon', () => {
    expect(parseTheme('neon')).toBe('neon')
  })

  it('defaults to info', () => {
    expect(parseTheme('info')).toBe('info')
    expect(parseTheme(null)).toBe('info')
    expect(parseTheme(undefined)).toBe('info')
    expect(parseTheme('')).toBe('info')
    expect(parseTheme('light')).toBe('info')
  })
})

describe('readTheme', () => {
  it('reads the saved theme', () => {
    expect(readTheme({ getItem: (k) => (k === 'theme' ? 'neon' : null) })).toBe('neon')
  })

  it('is info when nothing is saved', () => {
    expect(readTheme({ getItem: () => null })).toBe('info')
  })

  it('is info when the saved value is unknown', () => {
    expect(readTheme({ getItem: () => 'sepia' })).toBe('info')
  })

  it('is info when storage throws', () => {
    expect(readTheme({ getItem: () => { throw new Error('blocked') } })).toBe('info')
  })

  it('is info without storage', () => {
    expect(readTheme(undefined)).toBe('info')
  })
})

describe('writeTheme', () => {
  it('saves under the theme key', () => {
    const saved: Record<string, string> = {}
    writeTheme('neon', { setItem: (k, v) => { saved[k] = v } })
    expect(saved).toEqual({ theme: 'neon' })
  })

  it('does not throw when storage does', () => {
    expect(() => writeTheme('neon', { setItem: () => { throw new Error('full') } })).not.toThrow()
  })
})
