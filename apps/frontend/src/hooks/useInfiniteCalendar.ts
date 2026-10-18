import { useCallback, useEffect, useRef, useState } from 'react';
import type { Event } from '../types/event';
import { fetchCalendarEvents } from '../services/api';
import {
  MAX_EMPTY_WINDOWS,
  addDays,
  mergeEvents,
  windowAfter,
  windowBefore,
} from '../utils/calendarWindow';

interface UseInfiniteCalendarOptions {
  filterKey?: string;
  startDate: string;
  // Called synchronously right before earlier events are prepended, so the caller
  // can record scroll metrics and keep the viewport from jumping.
  onBeforePrepend?: () => void;
}

// Loads the calendar in date windows around startDate: forward from it and backward before it.
// State is not reset when the inputs change, mount it with a key built from them instead.
export function useInfiniteCalendar({ filterKey, startDate, onBeforePrepend }: UseInfiniteCalendarOptions) {
  const [events, setEvents] = useState<Event[]>([]);
  const [initialDone, setInitialDone] = useState(false);
  const [loadingEarlier, setLoadingEarlier] = useState(false);
  const [loadingLater, setLoadingLater] = useState(false);
  const [reachedStart, setReachedStart] = useState(false);
  const [reachedEnd, setReachedEnd] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // first and last day (inclusive) already requested
  const loaded = useRef({ first: startDate, last: addDays(startDate, -1) });
  const busy = useRef({ earlier: false, later: false });
  const mounted = useRef(false);
  const onBeforePrependRef = useRef(onBeforePrepend);

  useEffect(() => {
    onBeforePrependRef.current = onBeforePrepend;
  });

  useEffect(() => {
    mounted.current = true;
    return () => {
      mounted.current = false;
    };
  }, []);

  const loadLater = useCallback(async () => {
    if (busy.current.later || reachedEnd || error) return;
    busy.current.later = true;
    setLoadingLater(true);

    try {
      for (let empty = 0; empty < MAX_EMPTY_WINDOWS; empty++) {
        const win = windowAfter(loaded.current.last);
        const data = await fetchCalendarEvents({ filterKey, fromDate: win.from, toDate: win.to });
        if (!mounted.current) return;

        loaded.current.last = win.to;
        if (data.length > 0) {
          setEvents((prev) => mergeEvents(prev, data, 'append'));
          return;
        }
      }
      setReachedEnd(true);
    } catch (err) {
      if (mounted.current) setError(err instanceof Error ? err.message : 'Failed to load events');
    } finally {
      busy.current.later = false;
      if (mounted.current) {
        setLoadingLater(false);
        setInitialDone(true);
      }
    }
  }, [filterKey, reachedEnd, error]);

  const loadEarlier = useCallback(async () => {
    if (busy.current.earlier || reachedStart || error) return;
    busy.current.earlier = true;
    setLoadingEarlier(true);

    try {
      for (let empty = 0; empty < MAX_EMPTY_WINDOWS; empty++) {
        const win = windowBefore(loaded.current.first);
        const data = await fetchCalendarEvents({ filterKey, fromDate: win.from, toDate: win.to });
        if (!mounted.current) return;

        loaded.current.first = win.from;
        if (data.length > 0) {
          onBeforePrependRef.current?.();
          setEvents((prev) => mergeEvents(prev, data, 'prepend'));
          return;
        }
      }
      setReachedStart(true);
    } catch (err) {
      if (mounted.current) setError(err instanceof Error ? err.message : 'Failed to load events');
    } finally {
      busy.current.earlier = false;
      if (mounted.current) setLoadingEarlier(false);
    }
  }, [filterKey, reachedStart, error]);

  const retry = useCallback(() => setError(null), []);

  // first page
  useEffect(() => {
    void loadLater();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return {
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
  };
}
