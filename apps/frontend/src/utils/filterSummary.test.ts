import { describe, it, expect } from 'vitest'
import { ageLabel, describeFilter, describeFilterGroups } from './filterSummary'

describe('ageLabel', () => {
  it('calls the lowest age group all ages', () => {
    expect(ageLabel(1)).toBe('All ages')
    expect(ageLabel(0)).toBe('All ages')
  })

  it('adds a plus to the others', () => {
    expect(ageLabel(21)).toBe('21+')
  })
})

describe('describeFilter', () => {
  it('is empty without a filter', () => {
    expect(describeFilter({})).toEqual([])
  })

  it('lists included tags as they are and excluded ones with Not', () => {
    expect(describeFilter({ tags: ['music'], tagsInclude: true })).toEqual(['Music'])
    expect(describeFilter({ tags: ['music'], tagsInclude: false })).toEqual(['Not Music'])
  })

  it('says Max or Min for the age', () => {
    expect(describeFilter({ age: 18, agesLessThan: true })).toEqual(['Max 18+'])
    expect(describeFilter({ age: 18, agesLessThan: false })).toEqual(['Min 18+'])
  })

  it('lists regions as they are', () => {
    expect(describeFilter({ regions: ['tulsa', 'okc'] })).toEqual(['tulsa', 'okc'])
  })

  it('puts tags, then age, then regions, then venues', () => {
    expect(
      describeFilter({
        tags: ['Music'], tagsInclude: true,
        age: 1, agesLessThan: true,
        regions: ['okc'],
        venues: ["Cain's"], venuesInclude: false,
      }),
    ).toEqual(['Music', 'Max All ages', 'okc', "Not Cain's"])
  })
})

describe('describeFilterGroups', () => {
  it('is empty without a filter', () => {
    expect(describeFilterGroups({})).toEqual([])
  })

  it('names each kind the filter uses and leaves out the rest', () => {
    expect(
      describeFilterGroups({ tags: ['music'], tagsInclude: true, regions: ['okc'] }),
    ).toEqual([
      { title: 'Tags', parts: ['Music'] },
      { title: 'Regions', parts: ['okc'] },
    ])
  })
})
