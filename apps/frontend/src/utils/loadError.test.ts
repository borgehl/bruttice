import { describe, it, expect } from 'vitest'
import { ApiError, describeLoadError, describeSubmitError } from './loadError'

describe('describeLoadError', () => {
  it('says the server cannot be reached when fetch itself fails', () => {
    expect(describeLoadError(new TypeError('Failed to fetch')).title).toBe("Can't reach the server")
  })

  it('blames the server for 5xx answers', () => {
    expect(describeLoadError(new ApiError(503, 'Failed to fetch events')).message).toContain('on our side')
  })

  it('uses the general copy for other answers and other errors', () => {
    expect(describeLoadError(new ApiError(404, 'x')).title).toBe('Something went wrong')
    expect(describeLoadError(new Error('x')).title).toBe('Something went wrong')
    expect(describeLoadError('x').title).toBe('Something went wrong')
  })

  it('never shows the server or error text', () => {
    const { title, message } = describeLoadError(new ApiError(500, 'pq: connection refused'))
    expect(title + message).not.toContain('pq')
  })
})

describe('describeSubmitError', () => {
  it('shows the reason the server gives for a refusal', () => {
    expect(describeSubmitError(new ApiError(400, 'Message is too short'))).toBe('Message is too short')
    expect(describeSubmitError(new ApiError(429, 'Too many messages. Try again later'))).toBe('Too many messages. Try again later')
  })

  it('does not show the text of a server failure or a lost connection', () => {
    expect(describeSubmitError(new ApiError(502, 'pq: boom'))).not.toContain('pq')
    expect(describeSubmitError(new TypeError('Failed to fetch'))).toContain("Can't reach the server")
  })
})
