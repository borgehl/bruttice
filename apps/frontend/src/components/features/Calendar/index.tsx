import { useEffect, useLayoutEffect, useRef, useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { useSiteStore } from '../../../store/site';
import { fetchFilterDetails, type FilterDetails } from '../../../services/api';
import { useInfiniteCalendar } from '../../../hooks/useInfiniteCalendar';
import { todayLocal } from '../../../utils/calendarWindow';
import { groupEventsByDay } from '../../../utils/time';
import { describeFilter } from '../../../utils/filterSummary';
import EventDaySection from '../EventDaySection';
import StartDateFilter, { START_DATE_PARAM } from '../StartDateFilter';
import ActiveFilters from '../ActiveFilters';
import Skeleton from '../../ui/Skeleton';
import Spinner from '../../ui/Spinner';
import EmptyState from '../../ui/EmptyState';
import InlineError from '../../ui/InlineError';
import LoadFailed from '../../ui/LoadFailed';
import { ghostButtonClass, primaryButtonClass } from '../../ui/Field';

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

  const { site } = useSiteStore();

  const hasActiveFilters = filterDetails && describeFilter(filterDetails).length > 0;

  // Clear keeps the date the calendar starts from and drops only the filter.
  const clearParams = new URLSearchParams(searchParams);
  clearParams.delete('filterKey');
  const clearQuery = clearParams.toString();

  return (
    <div className="flex min-h-0 flex-1 flex-col">
      <header className="flex min-h-[60px] shrink-0 items-center justify-between gap-3 px-4 py-2 md:justify-end">
        <h1 className="font-label min-w-0 truncate md:hidden text-[17px] font-bold leading-tight tracking-tight text-text in-data-[theme=neon]:text-[19px] in-data-[theme=neon]:text-accent in-data-[theme=neon]:glow-text">
          <span>{site.name}</span>
          {site.region && (
            <span className="font-medium text-text-muted before:content-['_·_'] in-data-[theme=neon]:block in-data-[theme=neon]:text-xs in-data-[theme=neon]:text-accent-2 in-data-[theme=neon]:before:content-['//_']">
              {site.region}
            </span>
          )}
        </h1>
        <StartDateFilter startDate={startDate} />
      </header>

      {hasActiveFilters && filterDetails && (
        <ActiveFilters filterDetails={filterDetails} clearTo={clearQuery ? `/?${clearQuery}` : '/'} />
      )}

      {/* keyed so changing the start date or filter starts a fresh scroll */}
      <CalendarScroll
        key={`${startDate}|${filterKey ?? ''}`}
        filterKey={filterKey}
        startDate={startDate}
        filtered={Boolean(hasActiveFilters)}
        clearTo={clearQuery ? `/?${clearQuery}` : '/'}
      />
    </div>
  );
}

interface CalendarScrollProps {
  filterKey?: string;
  startDate: string;
  // a filter is narrowing the list, which decides what an empty calendar says
  filtered: boolean;
  clearTo: string;
}

function CalendarScroll({ filterKey, startDate, filtered, clearTo }: CalendarScrollProps) {
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
      <div className="flex-1 border-t border-border">
        <Skeleton variant="eventCard" count={6} label="Loading events" />
      </div>
    );
  }

  if (error && events.length === 0) {
    return (
      <div className="flex-1 border-t border-border px-4">
        <LoadFailed error={error} onRetry={retry} />
      </div>
    );
  }

  if (events.length === 0 && reachedStart && reachedEnd) {
    return (
      <div className="flex-1 border-t border-border">
        {filtered ? (
          <EmptyState
            title="No events match your filters"
            action={
              <Link to={clearTo} className={primaryButtonClass}>
                Clear filters
              </Link>
            }
          >
            Try removing a tag or venue.
          </EmptyState>
        ) : (
          <EmptyState
            title="No upcoming events"
            action={
              <Link to="/add-event" className={ghostButtonClass}>
                Add event
              </Link>
            }
          >
            Check back soon, or add one.
          </EmptyState>
        )}
      </div>
    );
  }

  return (
    // overflowAnchor off: the browser's own scroll anchoring would shift the view for prepended events
    // on top of the manual correction above, jumping forward by the height of the added window.
    <div
      ref={containerRef}
      style={{ overflowAnchor: 'none' }}
      className="min-h-0 flex-1 overflow-y-auto border-t border-border"
    >
      <div ref={topRef} className="py-3 text-center text-[13px] text-text-faint">
        {loadingEarlier && (
          <span className="inline-flex items-center gap-2">
            <Spinner size="sm" />
            Loading earlier events
          </span>
        )}
        {reachedStart && 'No earlier events'}
      </div>

      {groupEventsByDay(events).map((eventDay) => (
        <EventDaySection key={eventDay.day} eventDay={eventDay} />
      ))}

      <div ref={bottomRef} className="px-4 pb-7 pt-5 text-center text-[13px] text-text-faint">
        {loadingLater && (
          <span className="inline-flex items-center gap-2">
            <Spinner size="sm" />
            Loading more events
          </span>
        )}
        {reachedEnd && 'No more events'}
      </div>

      {error && <InlineError message={`${error.title}. ${error.message}`} onRetry={retry} className="m-4" />}
    </div>
  );
}
