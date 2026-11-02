import { useSearchParams } from 'react-router-dom';
import DateInput from '../../ui/DateInput';
import { todayLocal } from '../../../utils/calendarWindow';

export const START_DATE_PARAM = 'from-date';

interface StartDateFilterProps {
  startDate: string;
}

// "Oct 22" from "2026-10-22", read as a calendar day so the timezone can't move it.
function formatShortDate(date: string): string {
  const [year, month, day] = date.split('-').map(Number);
  return new Intl.DateTimeFormat('en-US', { month: 'short', day: 'numeric', timeZone: 'UTC' }).format(
    new Date(Date.UTC(year, month - 1, day)),
  );
}

// The date picker and the Today button of the calendar's top bar.
export default function StartDateFilter({ startDate }: StartDateFilterProps) {
  const [searchParams, setSearchParams] = useSearchParams();

  const setStartDate = (value: string) => {
    const newParams = new URLSearchParams(searchParams);
    if (value) {
      newParams.set(START_DATE_PARAM, value);
    } else {
      newParams.delete(START_DATE_PARAM);
    }
    // an old shared link may still carry an end date, the calendar no longer uses it
    newParams.delete('to-date');
    setSearchParams(newParams);
  };

  const isToday = startDate === todayLocal();

  return (
    <div className="flex items-center gap-1">
      <label className="relative flex h-10 cursor-pointer items-center gap-2 rounded-control border border-border bg-surface px-3 text-sm font-medium text-text">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className="text-text-muted max-[359px]:hidden" aria-hidden="true">
          <rect x="3" y="5" width="18" height="16" rx="2" />
          <path d="M3 10h18M8 3v4M16 3v4" />
        </svg>
        <span className="font-data whitespace-nowrap">{formatShortDate(startDate)}</span>
        <span className="sr-only">Go to date</span>
        <DateInput
          id="start-date"
          name="start-date"
          value={startDate}
          onChange={(e) => setStartDate(e.target.value)}
          className="absolute inset-0 h-full w-full cursor-pointer opacity-0"
        />
      </label>
      <button
        type="button"
        onClick={() => setStartDate('')}
        disabled={isToday}
        className="font-label chamfer h-10 cursor-pointer rounded-control px-3 text-sm font-semibold text-accent enabled:hover:text-accent-hover disabled:cursor-default disabled:text-text-faint in-data-[theme=neon]:bg-accent in-data-[theme=neon]:text-accent-ink in-data-[theme=neon]:enabled:hover:text-accent-ink in-data-[theme=neon]:disabled:text-accent-ink in-data-[theme=neon]:disabled:opacity-60"
      >
        Today
      </button>
    </div>
  );
}
