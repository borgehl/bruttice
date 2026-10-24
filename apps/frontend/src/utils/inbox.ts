import type { InboxMessage, InboxThread } from '../types/event';

const SOURCE_LABELS: Record<InboxMessage['source'], string> = {
  contact_form: 'Contact form',
  email: 'Email',
  admin_reply: 'Sent',
};

export function sourceLabel(source: InboxMessage['source']): string {
  return SOURCE_LABELS[source];
}

// Who a message is from, as the thread shows it. An admin's reply is the name that admin has set,
// or their email.
export function senderLabel(message: InboxMessage): string {
  return message.fromName || message.fromEmail || 'Admin';
}

// When an email an admin sent was delivered, empty if that is not known or the message was received.
export function deliveryLabel(
  message: Pick<InboxMessage, 'source' | 'deliveredAt'>,
  format: (unix: number) => string,
): string {
  if (message.source !== 'admin_reply' || message.deliveredAt <= 0) return '';
  return `Delivered ${format(message.deliveredAt)}`;
}

// One line of the latest message, to tell conversations apart in the list.
export function threadPreview(thread: InboxThread, max = 90): string {
  const latest = thread.messages[thread.messages.length - 1];
  if (!latest) return '';
  const line = latest.body.replace(/\s+/g, ' ').trim();
  return line.length > max ? `${line.slice(0, max - 1)}…` : line;
}
