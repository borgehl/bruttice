import { formatClock, formatDoors } from '../../../utils/time';

interface CalendarTimeProps {
  startAt: string;
  doorsAt?: string;
}

// The time column of an event row: the start time, with when doors open under it.
export default function CalendarTime({ startAt, doorsAt }: CalendarTimeProps) {
  return (
    <div className="w-[68px] shrink-0">
      <div className="font-data text-sm font-semibold leading-snug text-text">{formatClock(startAt)}</div>
      {doorsAt && <div className="mt-0.5 text-xs text-text-faint">{formatDoors(doorsAt, startAt)}</div>}
    </div>
  );
}
