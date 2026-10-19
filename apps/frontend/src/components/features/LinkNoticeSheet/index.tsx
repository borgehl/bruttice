import { useState } from 'react';
import { saveSkipLinkNotice, useLinkNoticeStore } from '../../../store/linkNotice';
import { useDialog } from '../../../hooks/useDialog';

// Shown once for every outbound event link, until someone ticks "Don't show this again".
export default function LinkNoticeSheet() {
  const url = useLinkNoticeStore((state) => state.pendingUrl);
  if (!url) return null;
  return <Sheet url={url} />;
}

function Sheet({ url }: { url: string }) {
  const close = useLinkNoticeStore((state) => state.close);
  const [skipNext, setSkipNext] = useState(false);
  const ref = useDialog<HTMLDivElement>(close);

  return (
    <div className="fixed inset-0 z-50 flex items-end justify-center bg-scrim md:items-center" onClick={close}>
      <div
        ref={ref}
        role="dialog"
        aria-modal="true"
        aria-labelledby="link-notice-title"
        onClick={(event) => event.stopPropagation()}
        className="flex w-full max-w-[480px] flex-col gap-3.5 border-t border-border bg-surface p-4 shadow-sheet md:rounded-control md:border"
      >
        <h2 id="link-notice-title" className="font-label text-lg font-bold text-text">
          Before you go
        </h2>
        <p className="text-sm leading-relaxed text-text-muted">
          Always confirm details by checking event links. Event information comes from venue websites and
          community submissions, and can be out of date.
        </p>
        <label className="flex min-h-10 cursor-pointer items-center gap-3 text-sm text-text">
          <input
            type="checkbox"
            checked={skipNext}
            onChange={(event) => setSkipNext(event.target.checked)}
            className="h-5 w-5 cursor-pointer accent-[var(--color-accent)]"
          />
          Don&apos;t show this again
        </label>
        <div className="flex gap-2.5">
          <button
            type="button"
            onClick={close}
            className="font-label h-12 flex-1 cursor-pointer rounded-control border border-border text-sm font-semibold text-text-muted"
          >
            Cancel
          </button>
          <a
            href={url}
            target="_blank"
            rel="noopener noreferrer"
            data-autofocus
            onClick={() => {
              if (skipNext) saveSkipLinkNotice();
              close();
            }}
            className="font-label chamfer flex h-12 flex-1 items-center justify-center rounded-control bg-accent text-sm font-semibold text-accent-ink hover:text-accent-ink"
          >
            Open
          </a>
        </div>
      </div>
    </div>
  );
}
