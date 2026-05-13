import { useEffect, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import type { Event } from '../../../types/event';
import { fetchCalendarEvents, fetchFilterDetails, type FilterDetails } from '../../../services/api';
import { groupEventsByDay } from '../../../utils/time';
import EventDaySection from '../EventDaySection';
import DateRangeFilter from '../DateRangeFilter';
import ActiveFilters from '../ActiveFilters';

export default function Calendar() {
  const [searchParams] = useSearchParams();
  const filterKey = searchParams.get('filterKey') || undefined;
  const fromDate = searchParams.get('from-date') || undefined;
  const toDate = searchParams.get('to-date') || undefined;

  const [events, setEvents] = useState<Event[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [filterDetails, setFilterDetails] = useState<FilterDetails | null>(null);

  useEffect(() => {
    async function loadEvents() {
      try {
        setLoading(true);
        setError(null);
        const data = await fetchCalendarEvents({ filterKey, fromDate, toDate });
        setEvents(data);

        if (filterKey) {
          const filter = await fetchFilterDetails(filterKey);
          setFilterDetails(filter);
        } else {
          setFilterDetails(null);
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load events');
      } finally {
        setLoading(false);
      }
    }

    loadEvents();
  }, [filterKey, fromDate, toDate]);

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

      <DateRangeFilter />

      {loading && (
        <div className="text-center py-12">
          <div className="inline-block animate-spin rounded-full h-8 w-8 border-4 border-primary border-t-transparent"></div>
          <p className="mt-4 text-text-muted">Loading events...</p>
        </div>
      )}

      {error && (
        <div className="p-4 bg-error-muted border border-error/30 rounded-lg">
          <p className="text-error">{error}</p>
        </div>
      )}

      {!loading && !error && events.length === 0 && (
        <div className="text-center py-12">
          <p className="text-text-muted">No events found.</p>
        </div>
      )}

      {!loading && !error && events.length > 0 && (
        <div className="max-h-[600px] overflow-y-auto border border-border rounded-lg p-4 bg-surface">
          {groupEventsByDay(events).map((eventDay) => (
            <EventDaySection key={eventDay.day} eventDay={eventDay} />
          ))}
        </div>
      )}
    </div>
  );
}
