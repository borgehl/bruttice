import { useCallback, useId, useRef, useState } from 'react';
import { Link } from 'react-router-dom';
import { useDismissable } from '../../../hooks/useDismissable';
import type { FilterDetails } from '../../../services/api';
import { describeFilterGroups } from '../../../utils/filterSummary';

interface CurrentFilterProps {
  // the saved filter the calendar is showing, null while it loads or when it can't be read
  filterDetails: FilterDetails | null;
  // where Clear leads: the calendar without the filter
  clearTo: string;
}

// The top bar's button for a calendar shown through a saved filter (?filterKey=). It opens a popup
// with what the filter is made of and a way to drop it, which stays even when the filter itself
// can't be read, so the visitor is never stuck on it.
export default function CurrentFilter({ filterDetails, clearTo }: CurrentFilterProps) {
  const [open, setOpen] = useState(false);

  const rootRef = useRef<HTMLDivElement>(null);
  const buttonRef = useRef<HTMLButtonElement>(null);
  const panelId = useId();

  const close = useCallback(() => setOpen(false), []);
  useDismissable(open, close, rootRef, buttonRef);

  const groups = filterDetails ? describeFilterGroups(filterDetails) : [];

  return (
    <div ref={rootRef} className="relative">
      <button
        ref={buttonRef}
        type="button"
        aria-haspopup="true"
        aria-expanded={open}
        aria-controls={panelId}
        onClick={() => setOpen((current) => !current)}
        className="flex h-10 cursor-pointer items-center gap-2 rounded-control border border-border bg-surface px-3 text-sm font-medium text-text"
      >
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className="text-text-muted" aria-hidden="true">
          <path d="M3 5h18l-7 8v6l-4 2v-8z" />
        </svg>
        Filter
      </button>

      {open && (
        <div
          id={panelId}
          role="dialog"
          aria-label="Current filter"
          className="absolute left-0 top-full z-20 mt-1 w-72 md:left-auto md:right-0 max-w-[calc(100vw-2rem)] rounded-control border border-border bg-surface p-3 shadow-[0_8px_24px_rgb(0_0_0/0.35)]"
        >
          <h2 className="font-label mb-2 text-sm font-semibold text-text">Current filter</h2>

          {groups.length > 0 ? (
            <dl className="mb-3 space-y-2 text-sm">
              {groups.map((group) => (
                <div key={group.title}>
                  <dt className="text-[13px] text-text-muted">{group.title}</dt>
                  <dd className="text-text">{group.parts.join(', ')}</dd>
                </div>
              ))}
            </dl>
          ) : (
            <p className="mb-3 text-sm text-text-muted">
              {filterDetails ? 'This filter does not narrow anything.' : 'This filter could not be loaded.'}
            </p>
          )}

          <Link
            to={clearTo}
            onClick={close}
            className="font-label flex min-h-10 items-center justify-center rounded-control border border-border text-sm font-semibold text-text hover:bg-raised"
          >
            Clear filter
          </Link>
        </div>
      )}
    </div>
  );
}
