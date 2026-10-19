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
    <section>
      <h3 className="sticky top-0 z-[1] border-y border-border bg-surface px-4 py-2 text-[13px] font-semibold text-text">
        {eventDay.day}
      </h3>
      <div className="space-y-3 p-4">
        {eventDay.events.map((event) => (
          <AdminEventCard key={event.guid} event={event} mode={mode} onVote={onVote} onUpdate={onUpdate} />
        ))}
      </div>
    </section>
  );
}
