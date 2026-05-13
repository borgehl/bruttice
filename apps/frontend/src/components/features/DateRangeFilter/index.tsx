import { useMemo } from 'react';
import { useSearchParams } from 'react-router-dom';
import DateInput from '../../ui/DateInput';

const TO_DATE_DEFAULT_DAYS = 30;

function formatDate(date: Date): string {
  return date.toISOString().split('T')[0];
}

function getDefaultDates() {
  const today = new Date();
  const toDate = new Date(today);
  toDate.setDate(toDate.getDate() + TO_DATE_DEFAULT_DAYS);

  return {
    fromDate: formatDate(today),
    toDate: formatDate(toDate),
  };
}

interface DateRangeFilterProps {
  onDateChange?: (fromDate: string, toDate: string) => void;
}

export default function DateRangeFilter({ onDateChange }: DateRangeFilterProps) {
  const [searchParams, setSearchParams] = useSearchParams();

  const defaults = useMemo(() => getDefaultDates(), []);

  const fromDate = searchParams.get('from-date') || defaults.fromDate;
  const toDate = searchParams.get('to-date') || defaults.toDate;

  const handleFromDateChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newFromDate = e.target.value;
    const newParams = new URLSearchParams(searchParams);

    if (newFromDate) {
      newParams.set('from-date', newFromDate);
    } else {
      newParams.delete('from-date');
    }

    setSearchParams(newParams);
    onDateChange?.(newFromDate, toDate);
  };

  const handleToDateChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newToDate = e.target.value;
    const newParams = new URLSearchParams(searchParams);

    if (newToDate) {
      newParams.set('to-date', newToDate);
    } else {
      newParams.delete('to-date');
    }

    setSearchParams(newParams);
    onDateChange?.(fromDate, newToDate);
  };

  return (
    <div className="flex flex-wrap items-center gap-4 mb-4">
      <div className="flex items-center gap-2">
        <DateInput
          id="from-date"
          name="from-date"
          value={fromDate}
          onChange={handleFromDateChange}
        />
      </div>
      <div className="flex items-center gap-2">
        <label htmlFor="to-date" className="text-sm text-text-muted">
          to
        </label>
        <DateInput
          id="to-date"
          name="to-date"
          value={toDate}
          onChange={handleToDateChange}
        />
      </div>
    </div>
  );
}
