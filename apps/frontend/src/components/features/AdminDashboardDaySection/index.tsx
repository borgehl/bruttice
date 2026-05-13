import type { AdminEventDay } from '../../../types/event';
import { AdminEventCard, type CardMode } from '../AdminDashboardEventCard';

interface AdminDashboardDaySectionProps {
  eventDay: AdminEventDay;
  mode: CardMode;
  onVote: (guid: string, vote: string) => void;
  onUpdate?: () => void;
}

export default function AdminDashboardDaySection({
  eventDay,
  mode,
  onVote,
  onUpdate,
}: AdminDashboardDaySectionProps) {
  return (
    <section className="mb-8">
      <h3 className="sticky -top-4 z-10 text-xl font-semibold text-text pt-4 pb-3 px-4 -mx-4 mb-4 border-b-2 border-primary bg-surface">
        {eventDay.day}
      </h3>
      <div className="space-y-4">
        {eventDay.events.map((event) => (
          <AdminEventCard key={event.guid} event={event} mode={mode} onVote={onVote} onUpdate={onUpdate} />
        ))}
      </div>
    </section>
  );
}
