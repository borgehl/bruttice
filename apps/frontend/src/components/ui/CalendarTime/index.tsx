import { formatEventTime } from '../../../utils/time';

interface CalendarTimeProps {
  startAt: string;
  doorsAt?: string;
}

export default function CalendarTime({ startAt, doorsAt }: CalendarTimeProps) {
  return (
    <span className="text-text-muted">{formatEventTime(startAt, doorsAt)}</span>
  );
}
