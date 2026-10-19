import { describe, it, expect } from 'vitest'
import { feedUrl } from './feeds'

describe('feedUrl', () => {
  it('is the feed of all events without a filter', () => {
    expect(feedUrl('https://x.example', 'rss', null)).toBe('https://x.example/api/feeds?feedType=rss')
  })

  it('adds the filter key', () => {
    expect(feedUrl('https://x.example', 'json', 'abc123')).toBe(
      'https://x.example/api/feeds?feedType=json&filterKey=abc123',
    )
  })

  it('escapes the filter key', () => {
    expect(feedUrl('https://x.example', 'atom', 'a b&c')).toBe(
      'https://x.example/api/feeds?feedType=atom&filterKey=a%20b%26c',
    )
  })
})
