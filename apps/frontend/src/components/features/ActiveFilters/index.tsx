import { Link } from 'react-router-dom';
import type { FilterDetails } from '../../../services/api';
import { describeFilter } from '../../../utils/filterSummary';

interface ActiveFiltersProps {
  filterDetails: FilterDetails;
  // where Clear leads: the calendar without the filter
  clearTo: string;
}

// One line under the top bar: what the calendar is filtered by, and a way to drop it.
export default function ActiveFilters({ filterDetails, clearTo }: ActiveFiltersProps) {
  const summary = describeFilter(filterDetails).join(', ');

  return (
    <div className="flex shrink-0 items-center gap-3 px-4 pb-2.5 text-[13px] text-text-muted">
      <span className="min-w-0 flex-1 truncate">
        Filtered by <span className="text-text">{summary}</span>
      </span>
      <Link to={clearTo} className="font-label flex min-h-8 items-center font-semibold">
        Clear
      </Link>
    </div>
  );
}
