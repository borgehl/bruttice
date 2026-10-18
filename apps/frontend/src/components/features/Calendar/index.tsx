import { useEffect, useLayoutEffect, useRef, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { fetchFilterDetails, type FilterDetails } from '../../../services/api';
import { useInfiniteCalendar } from '../../../hooks/useInfiniteCalendar';
import { todayLocal } from '../../../utils/calendarWindow';
import { groupEventsByDay } from '../../../utils/time';
import EventDaySection from '../EventDaySection';
import StartDateFilter, { START_DATE_PARAM } from '../StartDateFilter';
import ActiveFilters from '../ActiveFilters';

// Loads more when a sentinel is within this distance of the visible area.
const LOAD_MARGIN = '200px 0px';

export default function Calendar() {
  const [searchParams] = useSearchParams();
  const filterKey = searchParams.get('filterKey') || undefined;
  const startDate = searchParams.get(START_DATE_PARAM) || todayLocal();

  const [filterDetails, setFilterDetails] = useState<FilterDetails | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function loadFilter() {
      if (!filterKey) {
        setFilterDetails(null);
        return;
      }
      try {
        const filter = await fetchFilterDetails(filterKey);
        if (!cancelled) setFilterDetails(filter);
      } catch {
        if (!cancelled) setFilterDetails(null);
      }
    }

    loadFilter();
    return () => {
      cancelled = true;
    };
  }, [filterKey]);

  const hasActiveFilters = filterDetails && (
    (filterDetails.tags && filterDetails.tags.length > 0) ||
    (filterDetails.venues && filterDetails.venues.length > 0) ||
    (filterDetails.age && filterDetails.age > 0)
  );

  return (
    <div>
      {hasActiveFilters && filterKey && (
        <ActiveFilters filterKey={filterKey} filterDetails={filterDetails} />
      )}

      <StartDateFilter startDate={startDate} />

      {/* keyed so changing the start date or filter starts a fresh scroll */}
      <CalendarScroll key={`${startDate}|${filterKey ?? ''}`} filterKey={filterKey} startDate={startDate} />
    </div>
  );
}

interface CalendarScrollProps {
  filterKey?: string;
  startDate: string;
}

function CalendarScroll({ filterKey, startDate }: CalendarScrollProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const topRef = useRef<HTMLDivElement>(null);
  const bottomRef = useRef<HTMLDivElement>(null);
  // scrollHeight before earlier events were prepended, null when nothing is pending
  const heightBeforePrepend = useRef<number | null>(null);

  const {
    events,
    initialDone,
    loadingEarlier,
    loadingLater,
    reachedStart,
    reachedEnd,
    error,
    loadEarlier,
    loadLater,
    retry,
  } = useInfiniteCalendar({
    filterKey,
    startDate,
    onBeforePrepend: () => {
      heightBeforePrepend.current = containerRef.current?.scrollHeight ?? null;
    },
  });

  // Content added above the viewport pushes everything down, so shift the scroll by the added height.
  useLayoutEffect(() => {
    const el = containerRef.current;
    const before = heightBeforePrepend.current;
    if (!el || before === null) return;

    el.scrollTop += el.scrollHeight - before;
    heightBeforePrepend.current = null;
  }, [events]);

  // The observer is rebuilt after every load: a new observer reports the current intersection,
  // which keeps loading while a sentinel is still on screen (short or sparse stretches).
  useEffect(() => {
    const root = containerRef.current;
    const top = topRef.current;
    const bottom = bottomRef.current;
    if (!initialDone || !root || !top || !bottom) return;

    const observer = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          if (!entry.isIntersecting) continue;
          if (entry.target === top) void loadEarlier();
          else void loadLater();
        }
      },
      { root, rootMargin: LOAD_MARGIN },
    );
    observer.observe(top);
    observer.observe(bottom);

    return () => observer.disconnect();
  }, [initialDone, events, loadEarlier, loadLater, loadingEarlier, loadingLater, reachedStart, reachedEnd]);

  if (!initialDone) {
    return (
      <div className="text-center py-12">
        <div className="inline-block animate-spin rounded-full h-8 w-8 border-4 border-primary border-t-transparent"></div>
        <p className="mt-4 text-text-muted">Loading events...</p>
      </div>
    );
  }

  if (error && events.length === 0) {
    return (
      <div className="p-4 bg-error-muted border border-error/30 rounded-lg">
        <p className="text-error">{error}</p>
        <button type="button" onClick={retry} className="mt-2 text-sm text-primary hover:underline cursor-pointer">
          Try again
        </button>
      </div>
    );
  }

  if (events.length === 0 && reachedStart && reachedEnd) {
    return (
      <div className="text-center py-12">
        <p className="text-text-muted">No events found.</p>
      </div>
    );
  }

  return (
    // overflowAnchor off: the browser's own scroll anchoring would shift the view for prepended events
    // on top of the manual correction above, jumping forward by the height of the added window.
    <div
      ref={containerRef}
      style={{ overflowAnchor: 'none' }}
      className="max-h-[600px] overflow-y-auto border border-border rounded-lg p-4 bg-surface"
    >
      <div ref={topRef} className="text-center text-sm text-text-muted py-2">
        {loadingEarlier && 'Loading earlier events...'}
        {reachedStart && 'No earlier events'}
      </div>

      {groupEventsByDay(events).map((eventDay) => (
        <EventDaySection key={eventDay.day} eventDay={eventDay} />
      ))}

      <div ref={bottomRef} className="text-center text-sm text-text-muted py-2">
        {loadingLater && 'Loading more events...'}
        {reachedEnd && 'No more events'}
      </div>

      {error && (
        <div className="p-3 bg-error-muted border border-error/30 rounded-lg text-center">
          <p className="text-error text-sm">{error}</p>
          <button type="button" onClick={retry} className="mt-1 text-sm text-primary hover:underline cursor-pointer">
            Try again
          </button>
        </div>
      )}
    </div>
  );
}
