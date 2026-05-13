import type { EventDay } from '../../../types/event';
import EventCard from '../EventCard';

interface EventDaySectionProps {
  eventDay: EventDay;
}

export default function EventDaySection({ eventDay }: EventDaySectionProps) {
  return (
    <section className="mb-8">
      <h2 className="sticky -top-4 z-10 text-xl font-semibold text-text pt-4 pb-3 px-4 -mx-4 mb-4 border-b-2 border-primary bg-surface-elevated">
        {eventDay.day}
      </h2>
      <div className="space-y-1">
        {eventDay.events.map((event) => (
          <EventCard key={event.guid} event={event} />
        ))}
      </div>
    </section>
  );
}
