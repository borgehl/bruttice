import { describe, it, expect } from 'vitest'
import { defaultSiteConfig, mergeSiteConfig, pageSiteName } from './site'

describe('pageSiteName', () => {
  const page = (content: string | null) => ({
    querySelector: (selector: string) =>
      selector === 'meta[property="og:site_name"]' && content !== null
        ? { getAttribute: (name: string) => (name === 'content' ? content : null) }
        : null,
  })

  it('is the name the page was served with', () => {
    expect(pageSiteName(page('Do It Together Tulsa'))).toBe('Do It Together Tulsa')
    expect(pageSiteName(page('  Padded  '))).toBe('Padded')
  })

  it('is empty without a document or a tag', () => {
    expect(pageSiteName(undefined)).toBe('')
    expect(pageSiteName(page(null))).toBe('')
  })

  it('is Bruttice in the defaults when there is no page', () => {
    expect(defaultSiteConfig.name).toBe('Bruttice')
  })
})

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
