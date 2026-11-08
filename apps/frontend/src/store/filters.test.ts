import { describe, it, expect } from 'vitest'
import { FILTERS_KEY, filtersFromParams, parseFilters, readFilters, restoreParams, writeFilters } from './filters'

describe('parseFilters', () => {
  it('reads what writeFilters saves', () => {
    const raw = JSON.stringify({ filterKey: 'abc', regions: ['tulsa', 'okc'] })
    expect(parseFilters(raw)).toEqual({ filterKey: 'abc', regions: ['tulsa', 'okc'] })
  })

  it('is no filters for nothing, bad JSON or the wrong shape', () => {
    const none = { regions: [] }
    expect(parseFilters(null)).toEqual(none)
    expect(parseFilters('')).toEqual(none)
    expect(parseFilters('{nope')).toEqual(none)
    expect(parseFilters('42')).toEqual(none)
    expect(parseFilters('null')).toEqual(none)
  })

  it('drops values of the wrong type', () => {
    expect(parseFilters(JSON.stringify({ filterKey: 7, regions: ['tulsa', 3, '', ' '] }))).toEqual({
      filterKey: undefined,
      regions: ['tulsa'],
    })
    expect(parseFilters(JSON.stringify({ regions: 'tulsa' }))).toEqual({ filterKey: undefined, regions: [] })
  })
})

describe('readFilters', () => {
  it('reads the saved filters', () => {
    const storage = { getItem: (k: string) => (k === FILTERS_KEY ? '{"filterKey":"abc","regions":["okc"]}' : null) }
    expect(readFilters(storage)).toEqual({ filterKey: 'abc', regions: ['okc'] })
  })

  it('is no filters when storage throws or is missing', () => {
    expect(readFilters({ getItem: () => { throw new Error('blocked') } })).toEqual({ regions: [] })
    expect(readFilters(undefined)).toEqual({ regions: [] })
  })
})

describe('writeFilters', () => {
  it('saves under the filters key', () => {
    const saved: Record<string, string> = {}
    writeFilters({ filterKey: 'abc', regions: ['okc'] }, { setItem: (k, v) => { saved[k] = v } })
    expect(JSON.parse(saved[FILTERS_KEY])).toEqual({ filterKey: 'abc', regions: ['okc'] })
  })

  it('does not throw when storage does', () => {
    expect(() => writeFilters({ regions: [] }, { setItem: () => { throw new Error('full') } })).not.toThrow()
  })
})

describe('filtersFromParams', () => {
  it('reads the key and the regions', () => {
    expect(filtersFromParams(new URLSearchParams('filterKey=abc&regions=okc&from-date=2026-11-05'))).toEqual({
      filterKey: 'abc',
      regions: ['okc'],
    })
  })

  it('is empty for a bare address', () => {
    expect(filtersFromParams(new URLSearchParams())).toEqual({ filterKey: undefined, regions: [] })
  })
})

describe('restoreParams', () => {
  const saved = { filterKey: 'abc', regions: ['okc'] }

  it('puts the saved key and regions into a bare address, keeping other parameters', () => {
    const next = restoreParams(new URLSearchParams('from-date=2026-11-05'), saved)
    expect(next?.get('filterKey')).toBe('abc')
    expect(next?.getAll('regions')).toEqual(['okc'])
    expect(next?.get('from-date')).toBe('2026-11-05')
  })

  it('restores regions without a key', () => {
    const next = restoreParams(new URLSearchParams(), { regions: ['okc', 'tulsa'] })
    expect(next?.has('filterKey')).toBe(false)
    expect(next?.getAll('regions')).toEqual(['okc', 'tulsa'])
  })

  it('leaves an address that names a key alone', () => {
    expect(restoreParams(new URLSearchParams('filterKey=other'), saved)).toBeNull()
  })

  it('leaves an address that names regions alone', () => {
    expect(restoreParams(new URLSearchParams('regions=tulsa'), saved)).toBeNull()
  })

  it('does nothing when nothing is saved', () => {
    expect(restoreParams(new URLSearchParams(), { regions: [] })).toBeNull()
  })
})
