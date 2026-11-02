import { describe, it, expect } from 'vitest'
import { isActive, linkFor } from './nav'

describe('isActive', () => {
  it('matches the calendar on / and its alias only', () => {
    expect(isActive('/', '/', ['/calendar'])).toBe(true)
    expect(isActive('/calendar', '/', ['/calendar'])).toBe(true)
    expect(isActive('/filter', '/', ['/calendar'])).toBe(false)
  })

  it('matches a path and what is under it', () => {
    expect(isActive('/admin', '/admin')).toBe(true)
    expect(isActive('/admin/add-event', '/admin')).toBe(true)
    expect(isActive('/administrator', '/admin')).toBe(false)
  })

  it('matches the extra paths', () => {
    expect(isActive('/about', '/more', ['/about', '/p'])).toBe(true)
  })
})

describe('linkFor', () => {
  it('keeps the whole query for the calendar', () => {
    expect(linkFor('/', '?from-date=2026-10-22&filterKey=abc', 'abc')).toEqual({
      pathname: '/',
      search: '?from-date=2026-10-22&filterKey=abc',
    })
  })

  it('carries only the filter to Feeds', () => {
    expect(linkFor('/feeds', '?from-date=2026-10-22&filterKey=a b', 'a b')).toEqual({
      pathname: '/feeds',
      search: '?filterKey=a%20b',
    })
  })

  it('leaves other links alone', () => {
    expect(linkFor('/feeds', '', null)).toBe('/feeds')
    expect(linkFor('/filter', '?filterKey=abc', 'abc')).toBe('/filter')
  })
})
