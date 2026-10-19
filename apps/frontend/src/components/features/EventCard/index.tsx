import type { Event } from '../../../types/event';
import CalendarVenue from '../../ui/CalendarVenue';
import CalendarTitle from '../../ui/CalendarTitle';
import CalendarAge from '../../ui/CalendarAge';
import CalendarCost from '../../ui/CalendarCost';
import CalendarTime from '../../ui/CalendarTime';
import CalendarTags from '../../ui/CalendarTags';
import FlyerThumb from '../../ui/FlyerThumb';

interface EventCardProps {
  event: Event;
}

// Time first, because that is what people scan for, then title, venue and one quiet line. The flyer
// sits at the right so the left edge never moves, with or without one.
export default function EventCard({ event }: EventCardProps) {
  return (
    <article className="flex gap-3.5 border-b border-border px-4 py-3.5">
      <CalendarTime startAt={event.startAt} doorsAt={event.doorsAt} />

      <div className="flex min-w-0 flex-1 flex-col gap-0.5">
        <CalendarTitle title={event.title} link={event.link} />
        <div>
          <CalendarVenue name={event.venue} link={event.venueLink} />
        </div>
        <div className="flex flex-wrap items-center gap-x-1.5 text-[13px] text-text-faint">
          <CalendarAge age={event.age} />
          {event.coverCost && (
            <>
              <span aria-hidden="true">·</span>
              <CalendarCost cost={event.coverCost} />
            </>
          )}
        </div>
        <CalendarTags tags={event.tags} />
      </div>

      {event.flyerUrl && <FlyerThumb url={event.flyerUrl} title={event.title} />}
    </article>
  );
}
