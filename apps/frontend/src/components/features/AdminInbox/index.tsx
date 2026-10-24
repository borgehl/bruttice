import { useCallback, useEffect, useState, type FormEvent } from 'react';
import type { InboxThread } from '../../../types/event';
import { composeInboxMessage, fetchInbox, replyToInboxThread, updateInboxThread } from '../../../services/api';
import { deliveryLabel, senderLabel, sourceLabel, threadPreview } from '../../../utils/inbox';
import { describeSubmitError, toLoadError, type LoadError } from '../../../utils/loadError';
import Skeleton from '../../ui/Skeleton';
import EmptyState from '../../ui/EmptyState';
import InlineError from '../../ui/InlineError';
import LoadFailed from '../../ui/LoadFailed';
import Spinner from '../../ui/Spinner';
import Field, { fieldClass, primaryButtonClass } from '../../ui/Field';

interface AdminInboxProps {
  // called when a conversation was read, archived or answered, so the unread count can follow
  onChange?: () => void;
}

function formatTime(unix: number): string {
  return new Date(unix * 1000).toLocaleString([], { dateStyle: 'medium', timeStyle: 'short' });
}

// The messages visitors write through the contact form and the mail people send to the site,
// grouped into conversations, with a reply that is emailed and kept in the conversation. An admin
// can also write to someone first.
export default function AdminInbox({ onChange }: AdminInboxProps) {
  const [threads, setThreads] = useState<InboxThread[] | null>(null);
  const [loadError, setLoadError] = useState<LoadError | null>(null);
  const [showArchived, setShowArchived] = useState(false);
  const [openId, setOpenId] = useState<string | null>(null);
  const [drafts, setDrafts] = useState<Record<string, string>>({});
  const [sendingId, setSendingId] = useState<string | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);
  const [composing, setComposing] = useState(false);
  const [compose, setCompose] = useState({ to: '', subject: '', body: '' });
  const [composeSending, setComposeSending] = useState(false);
  const [composeError, setComposeError] = useState<string | null>(null);
  // bumped to load the inbox again, by Try again and after every change
  const [attempt, setAttempt] = useState(0);

  useEffect(() => {
    let cancelled = false;
    fetchInbox(showArchived)
      .then((loaded) => {
        if (cancelled) return;
        setThreads(loaded);
        setLoadError(null);
      })
      .catch((err) => {
        if (!cancelled) setLoadError(toLoadError(err));
      });
    return () => {
      cancelled = true;
    };
  }, [showArchived, attempt]);

  const reload = useCallback(() => {
    setAttempt((n) => n + 1);
    onChange?.();
  }, [onChange]);

  const retry = () => {
    setThreads(null);
    setLoadError(null);
    setAttempt((n) => n + 1);
  };

  const toggle = async (thread: InboxThread) => {
    const closing = openId === thread.threadId;
    setOpenId(closing ? null : thread.threadId);
    setActionError(null);
    if (closing || !thread.unread) return;
    try {
      await updateInboxThread(thread.threadId, 'read');
      reload();
    } catch (err) {
      console.error(err);
      setActionError('The conversation could not be marked as read.');
    }
  };

  const archive = async (thread: InboxThread) => {
    setActionError(null);
    try {
      await updateInboxThread(thread.threadId, thread.archived ? 'unarchive' : 'archive');
      if (!thread.archived) setOpenId(null);
      reload();
    } catch (err) {
      console.error(err);
      setActionError('The conversation could not be updated.');
    }
  };

  const send = async (event: FormEvent, thread: InboxThread) => {
    event.preventDefault();
    const body = (drafts[thread.threadId] ?? '').trim();
    if (!body) return;
    setSendingId(thread.threadId);
    setActionError(null);
    try {
      await replyToInboxThread(thread.threadId, body);
      setDrafts((current) => ({ ...current, [thread.threadId]: '' }));
      reload();
    } catch (err) {
      console.error(err);
      setActionError(describeSubmitError(err));
    } finally {
      setSendingId(null);
    }
  };

  const sendCompose = async (event: FormEvent) => {
    event.preventDefault();
    const to = compose.to.trim();
    const subject = compose.subject.trim();
    const body = compose.body.trim();
    if (!to || !subject || !body) return;
    setComposeSending(true);
    setComposeError(null);
    try {
      const threadId = await composeInboxMessage(to, subject, body);
      setCompose({ to: '', subject: '', body: '' });
      setComposing(false);
      setShowArchived(false);
      setOpenId(threadId);
      reload();
    } catch (err) {
      console.error(err);
      setComposeError(describeSubmitError(err));
    } finally {
      setComposeSending(false);
    }
  };

  const toolbar = (
    <div className="mb-3 flex flex-wrap items-center justify-between gap-x-4">
      <label className="flex min-h-11 cursor-pointer items-center gap-2 text-sm text-text-muted">
        <input
          type="checkbox"
          checked={showArchived}
          onChange={(e) => {
            setThreads(null);
            setShowArchived(e.target.checked);
          }}
          className="h-[18px] w-[18px] accent-[var(--color-accent)]"
        />
        Show archived
      </label>
      {!composing && (
        <button
          type="button"
          onClick={() => setComposing(true)}
          className="font-label h-11 cursor-pointer rounded-control border border-border px-4 text-sm font-semibold text-text-muted hover:text-text"
        >
          New email
        </button>
      )}
    </div>
  );

  const composeForm = composing && (
    <form onSubmit={(e) => void sendCompose(e)} className="mb-4 space-y-3 rounded-control border border-border bg-surface p-4">
      <Field label="To *" htmlFor="compose-to" hint="Their answer comes back here, in this conversation.">
        <input
          id="compose-to"
          type="email"
          required
          autoComplete="off"
          value={compose.to}
          onChange={(e) => setCompose((c) => ({ ...c, to: e.target.value }))}
          className={fieldClass}
        />
      </Field>
      <Field label="Subject *" htmlFor="compose-subject">
        <input
          id="compose-subject"
          type="text"
          required
          maxLength={200}
          value={compose.subject}
          onChange={(e) => setCompose((c) => ({ ...c, subject: e.target.value }))}
          className={fieldClass}
        />
      </Field>
      <Field label="Message *" htmlFor="compose-body">
        <textarea
          id="compose-body"
          required
          rows={6}
          maxLength={10000}
          value={compose.body}
          onChange={(e) => setCompose((c) => ({ ...c, body: e.target.value }))}
          className={`${fieldClass} h-auto resize-y py-3`}
        />
      </Field>
      {composeError && <InlineError message={composeError} />}
      <div className="flex gap-2">
        <button
          type="submit"
          disabled={composeSending || !compose.to.trim() || !compose.subject.trim() || !compose.body.trim()}
          className={`${primaryButtonClass} h-11 flex-1`}
        >
          {composeSending ? (
            <span className="inline-flex items-center gap-2">
              <Spinner size="sm" /> Sending
            </span>
          ) : (
            'Send email'
          )}
        </button>
        <button
          type="button"
          onClick={() => {
            setComposing(false);
            setComposeError(null);
          }}
          className="font-label h-11 cursor-pointer rounded-control border border-border px-4 text-sm font-semibold text-text-muted hover:text-text"
        >
          Cancel
        </button>
      </div>
    </form>
  );

  if (loadError && !threads) return <LoadFailed error={loadError} onRetry={retry} />;
  if (!threads) return <Skeleton variant="adminCard" count={3} label="Loading the inbox" />;

  return (
    <div>
      {toolbar}
      {composeForm}
      {actionError && <InlineError message={actionError} className="mb-3" />}
      {loadError && <InlineError message={`${loadError.title}. ${loadError.message}`} onRetry={retry} className="mb-3" />}

      {threads.length === 0 ? (
        <EmptyState title={showArchived ? 'Nothing here' : 'No messages'}>
          {showArchived
            ? 'No conversation has been archived.'
            : 'Messages from the contact form and email sent to the site show up here. Use New email to write to someone first.'}
        </EmptyState>
      ) : (
        <ul className="rounded-control border border-border bg-surface">
          {threads.map((thread) => {
            const open = openId === thread.threadId;
            const draft = drafts[thread.threadId] ?? '';
            return (
              <li key={thread.threadId} className="border-b border-border last:border-b-0">
                <button
                  type="button"
                  onClick={() => void toggle(thread)}
                  aria-expanded={open}
                  className="flex min-h-14 w-full cursor-pointer items-start gap-3 px-4 py-3 text-left"
                >
                  <span
                    className={`mt-1.5 h-2 w-2 shrink-0 rounded-full ${thread.unread ? 'bg-accent' : 'bg-transparent'}`}
                    aria-label={thread.unread ? 'Unread' : undefined}
                  />
                  <span className="min-w-0 flex-1">
                    <span className="flex items-baseline justify-between gap-3">
                      <span className={`truncate text-sm ${thread.unread ? 'font-semibold text-text' : 'text-text-muted'}`}>
                        {thread.fromName || thread.fromEmail}
                      </span>
                      <span className="font-data shrink-0 text-xs text-text-faint">{formatTime(thread.lastAt)}</span>
                    </span>
                    <span className="block truncate text-sm text-text">{thread.subject}</span>
                    {!open && <span className="block truncate text-[13px] text-text-faint">{threadPreview(thread)}</span>}
                  </span>
                </button>

                {open && (
                  <div className="space-y-3 px-4 pb-4">
                    <p className="text-[13px] text-text-faint">
                      {thread.fromName && `${thread.fromName} · `}
                      {thread.fromEmail}
                    </p>

                    {thread.messages.map((message) => (
                      <div
                        key={message.id}
                        className={`rounded-control border border-border p-3 ${message.source === 'admin_reply' ? 'bg-raised' : 'bg-bg'}`}
                      >
                        <p className="mb-1 flex flex-wrap items-baseline justify-between gap-x-3 text-[13px]">
                          <span className="font-medium text-text">
                            {senderLabel(message)}
                            <span className="ml-2 font-normal text-text-faint">{sourceLabel(message.source)}</span>
                          </span>
                          <span className="font-data text-xs text-text-faint">{formatTime(message.createdAt)}</span>
                        </p>
                        {/* plain text only: what people send is never shown as markup */}
                        <p className="whitespace-pre-wrap break-words text-sm text-text-muted">{message.body}</p>
                        {deliveryLabel(message, formatTime) && (
                          <p className="font-data mt-2 text-xs text-text-faint">{deliveryLabel(message, formatTime)}</p>
                        )}
                      </div>
                    ))}

                    <form onSubmit={(e) => void send(e, thread)} className="space-y-2">
                      <label htmlFor={`reply-${thread.threadId}`} className="sr-only">
                        Reply to {thread.fromName || thread.fromEmail}
                      </label>
                      <textarea
                        id={`reply-${thread.threadId}`}
                        rows={4}
                        maxLength={10000}
                        value={draft}
                        onChange={(e) => setDrafts((current) => ({ ...current, [thread.threadId]: e.target.value }))}
                        placeholder="Write a reply. It is emailed, and their answer comes back here."
                        className={`${fieldClass} h-auto resize-y py-3`}
                      />
                      <div className="flex gap-2">
                        <button
                          type="submit"
                          disabled={sendingId === thread.threadId || draft.trim() === ''}
                          className={`${primaryButtonClass} h-11 flex-1`}
                        >
                          {sendingId === thread.threadId ? (
                            <span className="inline-flex items-center gap-2">
                              <Spinner size="sm" /> Sending
                            </span>
                          ) : (
                            'Send reply'
                          )}
                        </button>
                        <button
                          type="button"
                          onClick={() => void archive(thread)}
                          className="font-label h-11 cursor-pointer rounded-control border border-border px-4 text-sm font-semibold text-text-muted hover:text-text"
                        >
                          {thread.archived ? 'Unarchive' : 'Archive'}
                        </button>
                      </div>
                    </form>
                  </div>
                )}
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}
