import { Fragment } from 'react';
import type { EventDay } from '../../../types/event';
import type { Sponsor } from '../../../types/sponsor';
import { todayLocal } from '../../../utils/calendarWindow';
import EventCard from '../EventCard';
import SponsorCard from '../SponsorCard';

interface EventDaySectionProps {
  eventDay: EventDay;
  // the sponsor card, and the event it comes right after. It is drawn only in the section that has
  // that event.
  sponsor?: Sponsor | null;
  sponsorAfterGuid?: string | null;
}

export default function EventDaySection({ eventDay, sponsor, sponsorAfterGuid }: EventDaySectionProps) {
  const isToday = eventDay.dateKey === todayLocal();
  const count = eventDay.events.length;

  return (
    <section>
      <h2 className="sticky top-0 z-[1] flex items-baseline gap-2 border-b border-border bg-surface px-4 py-2.5 text-[13px] font-semibold text-text">
        <span className="font-label in-data-[theme=neon]:text-xl in-data-[theme=neon]:text-accent-2">
          {eventDay.weekday}
          <span className="in-data-[theme=neon]:hidden">,</span>
        </span>
        <span className="font-data -ml-1 in-data-[theme=neon]:ml-0">{eventDay.date}</span>
        {isToday && (
          <span className="flex items-center gap-1.5 text-accent-2 in-data-[theme=neon]:text-accent">
            <span className="h-[7px] w-[7px] rounded-full bg-current" aria-hidden="true" />
            Today
          </span>
        )}
        <span className="font-data ml-auto font-normal text-text-faint">
          {count} {count === 1 ? 'event' : 'events'}
        </span>
      </h2>
      {eventDay.events.map((event) => (
        <Fragment key={event.guid}>
          <EventCard event={event} />
          {sponsor && event.guid === sponsorAfterGuid && <SponsorCard sponsor={sponsor} />}
        </Fragment>
      ))}
    </section>
  );
}
