import { useSearchParams } from 'react-router-dom';
import DateInput from '../../ui/DateInput';
import { todayLocal } from '../../../utils/calendarWindow';

export const START_DATE_PARAM = 'from-date';

interface StartDateFilterProps {
  startDate: string;
}

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

  return (
    <div className="flex flex-wrap items-center gap-4 mb-4">
      <label htmlFor="start-date" className="text-sm text-text-muted">
        Go To
      </label>
      <DateInput
        id="start-date"
        name="start-date"
        value={startDate}
        onChange={(e) => setStartDate(e.target.value)}
      />
      {startDate !== todayLocal() && (
        <button
          type="button"
          onClick={() => setStartDate('')}
          className="text-sm text-primary hover:underline cursor-pointer"
        >
          Today
        </button>
      )}
    </div>
  );
}
