import type { Event } from '../../../types/event';
import CalendarVenue from '../../ui/CalendarVenue';
import CalendarTitle from '../../ui/CalendarTitle';
import CalendarAge from '../../ui/CalendarAge';
import CalendarCost from '../../ui/CalendarCost';
import CalendarTime from '../../ui/CalendarTime';
import CalendarTags from '../../ui/CalendarTags';

interface EventCardProps {
  event: Event;
}

export default function EventCard({ event }: EventCardProps) {
  return (
    <div className="py-3 border-b border-border last:border-b-0">
      <div className="mb-1">
        <CalendarVenue name={event.venue} link={event.venueLink} />
      </div>

      <CalendarTitle title={event.title} link={event.link} />

      <div className="flex flex-wrap items-center gap-2 text-sm mb-2">
        <CalendarAge age={event.age} />
        <CalendarCost cost={event.coverCost} />
        <CalendarTime startAt={event.startAt} doorsAt={event.doorsAt} />
      </div>

      <CalendarTags tags={event.tags} />
    </div>
  );
}
