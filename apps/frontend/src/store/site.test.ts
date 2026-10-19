import { describe, it, expect } from 'vitest'
import { defaultSiteConfig, mergeSiteConfig } from './site'

describe('mergeSiteConfig', () => {
  it('is the defaults without a response', () => {
    expect(mergeSiteConfig(null)).toEqual(defaultSiteConfig)
    expect(mergeSiteConfig(undefined)).toEqual(defaultSiteConfig)
  })

  it('overrides the fields it gets', () => {
    const site = mergeSiteConfig({ name: 'Do It Together', region: 'Tulsa' })
    expect(site.name).toBe('Do It Together')
    expect(site.region).toBe('Tulsa')
    expect(site.description).toBe(defaultSiteConfig.description)
  })

  it('keeps default features that the response leaves out', () => {
    const site = mergeSiteConfig({ features: {} as never })
    expect(site.features.popBooker).toBe(false)
  })

  it('turns Pop! Booker on', () => {
    expect(mergeSiteConfig({ features: { popBooker: true } }).features.popBooker).toBe(true)
  })
})
