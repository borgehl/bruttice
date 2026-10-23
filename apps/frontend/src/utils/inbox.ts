import type { InboxMessage, InboxThread } from '../types/event';

const SOURCE_LABELS: Record<InboxMessage['source'], string> = {
  contact_form: 'Contact form',
  email: 'Email',
  admin_reply: 'Reply',
};

export function sourceLabel(source: InboxMessage['source']): string {
  return SOURCE_LABELS[source];
}

// Who a message is from, as the thread shows it. An admin's reply is the name that admin has set,
// or their email.
export function senderLabel(message: InboxMessage): string {
  return message.fromName || message.fromEmail || 'Admin';
}

// One line of the latest message, to tell conversations apart in the list.
export function threadPreview(thread: InboxThread, max = 90): string {
  const latest = thread.messages[thread.messages.length - 1];
  if (!latest) return '';
  const line = latest.body.replace(/\s+/g, ' ').trim();
  return line.length > max ? `${line.slice(0, max - 1)}…` : line;
}
