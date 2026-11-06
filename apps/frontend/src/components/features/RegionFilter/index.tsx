import { useCallback, useEffect, useId, useRef, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { useDismissable } from '../../../hooks/useDismissable';
import { fetchFilterOptions } from '../../../services/api';
import { regionLabel, regionsFromParams, withRegions } from '../../../utils/regionFilter';
import { toggle } from '../../../utils/venueFilter';

// The region picker of the calendar's top bar. The button says "All Regions" until some are picked,
// and opens a list of the site's regions to tick. The choice lives in the address (see
// utils/regionFilter), so the calendar reloads from it and a link carries it.
//
// Nothing is shown for a site with fewer than two regions: there is nothing to choose between.
export default function RegionFilter() {
  const [searchParams, setSearchParams] = useSearchParams();
  const [available, setAvailable] = useState<string[]>([]);
  const [open, setOpen] = useState(false);

  const rootRef = useRef<HTMLDivElement>(null);
  const buttonRef = useRef<HTMLButtonElement>(null);
  const panelId = useId();

  const selected = regionsFromParams(searchParams);

  useEffect(() => {
    let cancelled = false;

    // without the list the calendar still works, it just can't be narrowed
    fetchFilterOptions()
      .then((options) => {
        if (!cancelled) setAvailable(options.regions.map((region) => region.name));
      })
      .catch(() => {});

    return () => {
      cancelled = true;
    };
  }, []);

  const close = useCallback(() => setOpen(false), []);
  useDismissable(open, close, rootRef, buttonRef);

  if (available.length < 2) return null;

  // replace, so each tick isn't a step for the back button
  const setRegions = (regions: string[]) => setSearchParams(withRegions(searchParams, regions), { replace: true });

  const rowClass =
    'flex min-h-10 cursor-pointer items-center gap-3 rounded-control px-3 text-sm text-text hover:bg-raised';

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
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className="text-text-muted max-[359px]:hidden" aria-hidden="true">
          <path d="M12 21s-7-6.2-7-11a7 7 0 0 1 14 0c0 4.8-7 11-7 11z" />
          <circle cx="12" cy="10" r="2.5" />
        </svg>
        <span className="max-w-28 truncate whitespace-nowrap max-[359px]:max-w-20">{regionLabel(selected, available)}</span>
        <span className="sr-only">Choose regions</span>
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" className="text-text-muted" aria-hidden="true">
          <path d="m6 9 6 6 6-6" />
        </svg>
      </button>

      {open && (
        <div
          id={panelId}
          role="group"
          aria-label="Regions"
          className="absolute left-0 top-full z-20 mt-1 min-w-48 md:left-auto md:right-0 max-w-[calc(100vw-2rem)] rounded-control border border-border bg-surface p-1 shadow-[0_8px_24px_rgb(0_0_0/0.35)]"
        >
          <label className={rowClass}>
            <input
              type="checkbox"
              className="h-[18px] w-[18px] shrink-0 cursor-pointer accent-[var(--color-accent)]"
              checked={selected.length === 0}
              // there is nothing to untick: every region is what no region means
              onChange={() => setRegions([])}
            />
            All Regions
          </label>
          {available.map((region) => (
            <label key={region} className={rowClass}>
              <input
                type="checkbox"
                className="h-[18px] w-[18px] shrink-0 cursor-pointer accent-[var(--color-accent)]"
                checked={selected.includes(region)}
                onChange={() => setRegions(toggle(selected, region))}
              />
              <span className="min-w-0 truncate">{region}</span>
            </label>
          ))}
        </div>
      )}
    </div>
  );
}
