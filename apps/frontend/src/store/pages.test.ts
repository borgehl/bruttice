import { describe, it, expect } from 'vitest'
import { hasPage, navPages } from './pages'
import type { PageSummary } from '../types/event'

const page = (slug: string, nav: boolean): PageSummary => ({ slug, title: slug, nav, order: 0 })

describe('navPages', () => {
  it('lists the pages marked for the navigation', () => {
    expect(navPages([page('faq', true), page('notes', false)]).map((p) => p.slug)).toEqual(['faq'])
  })

  it('leaves out the pages that replace a built-in one', () => {
    expect(navPages([page('about', true), page('privacy-policy', true), page('faq', true)]).map((p) => p.slug)).toEqual(['faq'])
  })

  it('keeps the order it was given', () => {
    expect(navPages([page('b', true), page('a', true)]).map((p) => p.slug)).toEqual(['b', 'a'])
  })
})

describe('hasPage', () => {
  it('finds a slug', () => {
    expect(hasPage([page('about', false)], 'about')).toBe(true)
    expect(hasPage([page('about', false)], 'faq')).toBe(false)
    expect(hasPage([], 'about')).toBe(false)
  })
})
