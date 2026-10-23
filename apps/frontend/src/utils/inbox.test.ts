import { describe, it, expect } from 'vitest'
import { senderLabel, sourceLabel, threadPreview } from './inbox'
import type { InboxMessage, InboxThread } from '../types/event'

function message(overrides: Partial<InboxMessage> = {}): InboxMessage {
  return { id: 1, source: 'contact_form', fromName: 'Sam', fromEmail: 'sam@example.com', body: 'Hi', createdAt: 1, unread: true, ...overrides }
}

function thread(messages: InboxMessage[]): InboxThread {
  return { threadId: 't', subject: 'Hi', fromName: 'Sam', fromEmail: 'sam@example.com', unread: true, archived: false, lastAt: 1, messages }
}

describe('senderLabel', () => {
  it('is the name, or the address when there is no name', () => {
    expect(senderLabel(message())).toBe('Sam')
    expect(senderLabel(message({ fromName: '' }))).toBe('sam@example.com')
  })

  it('is the name of the admin for a reply', () => {
    expect(senderLabel(message({ source: 'admin_reply', fromName: 'Kim' }))).toBe('Kim')
  })

  it('has a label even when nothing is known', () => {
    expect(senderLabel(message({ source: 'admin_reply', fromName: '', fromEmail: '' }))).toBe('Admin')
  })
})

describe('sourceLabel', () => {
  it('names where a message came from', () => {
    expect(sourceLabel('contact_form')).toBe('Contact form')
    expect(sourceLabel('email')).toBe('Email')
  })
})

describe('threadPreview', () => {
  it('is the latest message on one line', () => {
    const t = thread([message({ body: 'first' }), message({ body: 'Is Friday\n\n  all   ages?' })])
    expect(threadPreview(t)).toBe('Is Friday all ages?')
  })

  it('cuts a long message', () => {
    const preview = threadPreview(thread([message({ body: 'a'.repeat(200) })]), 20)
    expect(preview).toHaveLength(20)
    expect(preview.endsWith('…')).toBe(true)
  })

  it('is empty for a thread without messages', () => {
    expect(threadPreview(thread([]))).toBe('')
  })
})
