import { useCallback, useId, useRef, useState } from 'react';
import { useDismissable } from '../../../hooks/useDismissable';

export interface AdminPageMenuItem {
  value: string;
  label: string;
  // a count to show next to the label, left out when there is nothing to count
  badge?: number;
}

interface AdminPageMenuProps {
  items: AdminPageMenuItem[];
  // the item that is on screen now
  current: string;
  onSelect: (value: string) => void;
}

function Badge({ count }: { count: number }) {
  return (
    <span className="font-data rounded-full bg-accent px-1.5 py-0.5 text-xs leading-none text-accent-ink">
      {count}
      <span className="sr-only"> unread</span>
    </span>
  );
}

// The admin pages as one list, in the look of the region picker: a small button that says where you
// are, and a panel of every page. The pages are all the same level, so there are no groups.
export default function AdminPageMenu({ items, current, onSelect }: AdminPageMenuProps) {
  const [open, setOpen] = useState(false);
  const rootRef = useRef<HTMLDivElement>(null);
  const buttonRef = useRef<HTMLButtonElement>(null);
  const panelId = useId();

  const close = useCallback(() => setOpen(false), []);
  useDismissable(open, close, rootRef, buttonRef);

  const currentItem = items.find((item) => item.value === current);
  const badgeTotal = items.reduce((sum, item) => sum + (item.badge ?? 0), 0);

  return (
    <div ref={rootRef} className="relative inline-block">
      <button
        ref={buttonRef}
        type="button"
        aria-haspopup="true"
        aria-expanded={open}
        aria-controls={panelId}
        onClick={() => setOpen((value) => !value)}
        className="flex h-9 cursor-pointer items-center gap-2 rounded-control border border-border bg-surface px-3 text-sm font-medium text-text"
      >
        <span className="sr-only">Admin page: </span>
        <span className="whitespace-nowrap">{currentItem?.label}</span>
        {/* counts show on the button too, so unread mail is seen without opening the list */}
        {badgeTotal > 0 && <Badge count={badgeTotal} />}
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" className="text-text-muted" aria-hidden="true">
          <path d="m6 9 6 6 6-6" />
        </svg>
      </button>

      {open && (
        <div
          id={panelId}
          role="group"
          aria-label="Admin pages"
          className="absolute left-0 top-full z-20 mt-1 min-w-44 max-w-[calc(100vw-2rem)] rounded-control border border-border bg-surface p-1 shadow-[0_8px_24px_rgb(0_0_0/0.35)]"
        >
          {items.map((item) => {
            const selected = item.value === current;
            return (
              <button
                key={item.value}
                type="button"
                aria-current={selected ? 'true' : undefined}
                onClick={() => {
                  setOpen(false);
                  onSelect(item.value);
                }}
                className={`flex min-h-9 w-full cursor-pointer items-center gap-2 rounded-control px-3 text-left text-sm hover:bg-raised ${
                  selected ? 'font-semibold text-accent' : 'text-text'
                }`}
              >
                <span className="min-w-0 flex-1 truncate">{item.label}</span>
                {item.badge ? <Badge count={item.badge} /> : null}
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
}
