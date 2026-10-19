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

  it('keeps a default the server leaves empty', () => {
    const site = mergeSiteConfig({ name: '', region: 'Tulsa', description: '' })
    expect(site.name).toBe(defaultSiteConfig.name)
    expect(site.description).toBe(defaultSiteConfig.description)
    expect(site.region).toBe('Tulsa')
  })
})
