import { useEffect, useLayoutEffect, useRef, useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { useSiteStore } from '../../../store/site';
import { fetchFilterDetails, type FilterDetails } from '../../../services/api';
import { useInfiniteCalendar } from '../../../hooks/useInfiniteCalendar';
import { useSponsor } from '../../../hooks/useSponsor';
import { scrollTopAfterPrepend } from '../../../utils/prependScroll';
import { SPONSOR_CARD_HEIGHT, sponsorAnchor } from '../../../utils/sponsor';
import { todayLocal } from '../../../utils/calendarWindow';
import { groupEventsByDay } from '../../../utils/time';
import { describeFilter } from '../../../utils/filterSummary';
import { REGIONS_PARAM, regionsFromParams, regionsKey, withRegions } from '../../../utils/regionFilter';
import { usePersistedFilters } from '../../../hooks/usePersistedFilters';
import CurrentFilter from '../CurrentFilter';
import EventDaySection from '../EventDaySection';
import RegionFilter from '../RegionFilter';
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
  const [searchParams, setSearchParams] = useSearchParams();
  const filterKey = searchParams.get('filterKey') || undefined;
  const startDate = searchParams.get(START_DATE_PARAM) || todayLocal();
  const regions = regionsFromParams(searchParams);

  const [filterDetails, setFilterDetails] = useState<FilterDetails | null>(null);
  const keepsRegions = usePersistedFilters();

  useEffect(() => {
    let cancelled = false;

    async function loadFilter() {
      if (!filterKey) {
        setFilterDetails(null);
        return;
      }
      try {
        const filter = await fetchFilterDetails(filterKey);
        if (cancelled) return;
        setFilterDetails(filter);

        // The picker starts at the key's regions, and the visitor can change it from there. Regions
        // restored along with this key are the visitor's own, so they stay.
        if (!keepsRegions(filterKey)) {
          setSearchParams((params) => withRegions(params, filter.regions ?? []), { replace: true });
        }
      } catch {
        if (!cancelled) setFilterDetails(null);
      }
    }

    loadFilter();
    return () => {
      cancelled = true;
    };
  }, [filterKey, keepsRegions, setSearchParams]);

  const { site } = useSiteStore();

  const hasActiveFilters = filterDetails && describeFilter(filterDetails).length > 0;

  // Clearing the filter keeps the date the calendar starts from and the regions picked, they have
  // their own controls. The empty calendar's way out drops the regions too.
  const withoutFilter = new URLSearchParams(searchParams);
  withoutFilter.delete('filterKey');
  const withoutFilterQuery = withoutFilter.toString();

  const withoutAll = new URLSearchParams(withoutFilter);
  withoutAll.delete(REGIONS_PARAM);
  const withoutAllQuery = withoutAll.toString();

  return (
    <div className="flex min-h-0 flex-1 flex-col">
      {/* on a phone the title takes the first line and the controls the second, so they never cover it */}
      <header className="flex min-h-[60px] shrink-0 flex-wrap items-center gap-x-3 gap-y-2 px-4 py-2 md:flex-nowrap md:justify-end">
        <h1 className="font-label w-full min-w-0 truncate md:hidden text-[17px] font-bold leading-tight tracking-tight text-text in-data-[theme=neon]:text-[19px] in-data-[theme=neon]:text-accent in-data-[theme=neon]:glow-text">
          <span>{site.name}</span>
          {site.region && (
            <span className="font-medium text-text-muted before:content-['_·_'] in-data-[theme=neon]:block in-data-[theme=neon]:text-xs in-data-[theme=neon]:text-accent-2 in-data-[theme=neon]:before:content-['//_']">
              {site.region}
            </span>
          )}
        </h1>
        <div className="flex w-full flex-wrap items-center gap-2 md:w-auto md:shrink-0 md:flex-nowrap">
          {filterKey && (
            <CurrentFilter filterDetails={filterDetails} clearTo={withoutFilterQuery ? `/?${withoutFilterQuery}` : '/'} />
          )}
          <RegionFilter />
          <StartDateFilter startDate={startDate} />
        </div>
      </header>

      {hasActiveFilters && filterDetails && (
        <ActiveFilters
          filterDetails={filterDetails}
          clearTo={withoutFilterQuery ? `/?${withoutFilterQuery}` : '/'}
        />
      )}

      {/* keyed so changing the start date, filter or regions starts a fresh scroll */}
      <CalendarScroll
        key={`${startDate}|${filterKey ?? ''}|${regionsKey(regions)}`}
        filterKey={filterKey}
        regions={regions}
        startDate={startDate}
        filtered={Boolean(hasActiveFilters) || regions.length > 0}
        clearTo={withoutAllQuery ? `/?${withoutAllQuery}` : '/'}
      />
    </div>
  );
}

interface CalendarScrollProps {
  filterKey?: string;
  // region names to show, all of them when empty
  regions: string[];
  startDate: string;
  // a filter is narrowing the list, which decides what an empty calendar says
  filtered: boolean;
  clearTo: string;
}

function CalendarScroll({ filterKey, regions, startDate, filtered, clearTo }: CalendarScrollProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const topRef = useRef<HTMLDivElement>(null);
  const bottomRef = useRef<HTMLDivElement>(null);
  // What the scroll box looked like just before earlier events were prepended, null when nothing is
  // pending: how tall it was, how far it was scrolled, and the first event on the page.
  const beforePrepend = useRef<{ height: number; scrollTop: number; firstGuid: string | null; initial: boolean } | null>(null);
  // whether earlier events have been prepended yet, so the first time can be told from the rest
  const hasPrepended = useRef(false);

  // One sponsor card per page load, right after the last event of the first load. This component is
  // remounted when the start date, filter or regions change, so each of those gets a fresh pick.
  const sponsor = useSponsor();

  // The event the card follows, chosen once from what is on screen when the events first show, so
  // the card is in view on arrival. It never changes after that: earlier or later events, and
  // scrolling, don't move it.
  const [anchorGuid, setAnchorGuid] = useState<string | null>(null);
  const anchorChosen = useRef(false);

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
    regions,
    startDate,
    onBeforePrepend: () => {
      const el = containerRef.current;
      if (!el) return;
      beforePrepend.current = {
        height: el.scrollHeight,
        scrollTop: el.scrollTop,
        firstGuid: el.querySelector<HTMLElement>('[data-event-guid]')?.dataset.eventGuid ?? null,
        initial: !hasPrepended.current,
      };
      hasPrepended.current = true;
    },
  });

  // Content added above the viewport pushes everything down, so the scroll is shifted to keep the view
  // where it was. The first time, the view goes to the start of the day the page opened at instead, see
  // scrollTopAfterPrepend.
  useLayoutEffect(() => {
    const el = containerRef.current;
    const before = beforePrepend.current;
    if (!el || !before) return;
    beforePrepend.current = null;

    let sectionOffset: number | null = null;
    if (before.firstGuid) {
      const row = Array.from(el.querySelectorAll<HTMLElement>('[data-event-guid]')).find((r) => r.dataset.eventGuid === before.firstGuid);
      const section = row?.closest('section');
      if (section) sectionOffset = section.getBoundingClientRect().top - el.getBoundingClientRect().top;
    }

    el.scrollTop = scrollTopAfterPrepend(
      {
        initial: before.initial,
        scrollTopBefore: before.scrollTop,
        heightBefore: before.height,
        heightAfter: el.scrollHeight,
        sectionOffset,
      },
      el.scrollTop,
    );
  }, [events]);

  // The first time the events are on the page, before anything earlier is loaded above them, find the
  // last event that is fully in the scroll box with room under it for the card.
  useLayoutEffect(() => {
    const root = containerRef.current;
    if (anchorChosen.current || !root || events.length === 0) return;
    anchorChosen.current = true;

    const view = root.getBoundingClientRect();
    const boxes = Array.from(root.querySelectorAll<HTMLElement>('[data-event-guid]')).map((row) => {
      const rect = row.getBoundingClientRect();
      return { guid: row.dataset.eventGuid ?? '', top: rect.top, bottom: rect.bottom };
    });
    setAnchorGuid(sponsorAnchor(boxes, { top: view.top, bottom: view.bottom }, SPONSOR_CARD_HEIGHT));
  }, [events, initialDone]);

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
            Try removing a tag, venue or region.
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
        <EventDaySection key={eventDay.day} eventDay={eventDay} sponsor={sponsor} sponsorAfterGuid={anchorGuid} />
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
