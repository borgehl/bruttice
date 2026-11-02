import { describe, it, expect } from 'vitest'
import { defaultSiteConfig, mergeSiteConfig } from './site'

describe('mergeSiteConfig', () => {
  it('is the defaults without a response', () => {
    expect(mergeSiteConfig(null)).toEqual(defaultSiteConfig)
    expect(mergeSiteConfig(undefined)).toEqual(defaultSiteConfig)
  })

  it('overrides the fields it gets', () => {
    const site = mergeSiteConfig({ name: 'Example Events', region: 'Tulsa' })
    expect(site.name).toBe('Example Events')
    expect(site.region).toBe('Tulsa')
    expect(site.description).toBe(defaultSiteConfig.description)
  })

  it('has a contact form only if the server says so', () => {
    expect(mergeSiteConfig(null).contactForm).toBe(false)
    expect(mergeSiteConfig({ contactForm: false }).contactForm).toBe(false)
    expect(mergeSiteConfig({ contactForm: true }).contactForm).toBe(true)
  })

  it('keeps a default the server leaves empty', () => {
    const site = mergeSiteConfig({ name: '', region: 'Tulsa', description: '' })
    expect(site.name).toBe(defaultSiteConfig.name)
    expect(site.description).toBe(defaultSiteConfig.description)
    expect(site.region).toBe('Tulsa')
  })
})
