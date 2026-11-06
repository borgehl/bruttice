import { describe, it, expect } from 'vitest'
import { ALL_REGIONS_LABEL, regionLabel, regionsFromParams, regionsKey, withRegions } from './regionFilter'

describe('regionsFromParams', () => {
  it('is empty when the address has no regions', () => {
    expect(regionsFromParams(new URLSearchParams('from-date=2026-11-05'))).toEqual([])
  })

  it('reads one parameter per region, in order', () => {
    expect(regionsFromParams(new URLSearchParams('regions=okc&regions=tulsa'))).toEqual(['okc', 'tulsa'])
  })

  it('drops empty and repeated names and trims the rest', () => {
    expect(regionsFromParams(new URLSearchParams('regions=&regions=%20okc%20&regions=okc&regions=tulsa'))).toEqual([
      'okc',
      'tulsa',
    ])
  })

  it('keeps a comma as part of a name', () => {
    expect(regionsFromParams(new URLSearchParams({ regions: 'Dallas, TX' }))).toEqual(['Dallas, TX'])
  })
})

describe('withRegions', () => {
  it('replaces the regions and keeps the other parameters', () => {
    const params = new URLSearchParams('filterKey=abc&regions=tulsa&from-date=2026-11-05')
    const next = withRegions(params, ['okc', 'wichita'])

    expect(next.getAll('regions')).toEqual(['okc', 'wichita'])
    expect(next.get('filterKey')).toBe('abc')
    expect(next.get('from-date')).toBe('2026-11-05')
  })

  it('leaves no parameter for every region', () => {
    expect(withRegions(new URLSearchParams('regions=tulsa&filterKey=abc'), []).toString()).toBe('filterKey=abc')
  })

  it('does not change the parameters it was given', () => {
    const params = new URLSearchParams('regions=tulsa')
    withRegions(params, ['okc'])
    expect(params.getAll('regions')).toEqual(['tulsa'])
  })
})

describe('regionLabel', () => {
  const available = ['tulsa', 'okc', 'wichita']

  it('says All Regions when none is picked', () => {
    expect(regionLabel([], available)).toBe(ALL_REGIONS_LABEL)
  })

  it('says the name of a single region', () => {
    expect(regionLabel(['okc'], available)).toBe('okc')
  })

  it('counts several regions', () => {
    expect(regionLabel(['okc', 'tulsa'], available)).toBe('2 regions')
  })

  it('says All Regions when every region is picked', () => {
    expect(regionLabel(['wichita', 'okc', 'tulsa'], available)).toBe(ALL_REGIONS_LABEL)
  })

  it('does not count a name the site does not have', () => {
    expect(regionLabel(['gone'], available)).toBe(ALL_REGIONS_LABEL)
    expect(regionLabel(['okc', 'gone'], available)).toBe('okc')
  })
})

describe('regionsKey', () => {
  it('is the same whatever order the regions were picked in', () => {
    expect(regionsKey(['okc', 'tulsa'])).toBe(regionsKey(['tulsa', 'okc']))
  })

  it('differs for different regions, and for none', () => {
    expect(regionsKey(['okc'])).not.toBe(regionsKey(['tulsa']))
    expect(regionsKey([])).not.toBe(regionsKey(['okc']))
  })

  it('does not reorder the list it was given', () => {
    const regions = ['tulsa', 'okc']
    regionsKey(regions)
    expect(regions).toEqual(['tulsa', 'okc'])
  })
})
