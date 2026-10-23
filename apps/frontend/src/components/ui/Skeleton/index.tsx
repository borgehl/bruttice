import { useAfterDelay } from '../../../hooks/useAfterDelay';

type Variant = 'eventCard' | 'adminCard' | 'runRow';

interface SkeletonProps {
  variant: Variant;
  // read out to screen readers
  label: string;
  count?: number;
}

function Bar({ className }: { className: string }) {
  return <div className={`h-3 rounded-control bg-raised ${className}`} />;
}

// Same shape as EventCard, minus the flyer: that is optional, and the left edge must not move.
function EventCardSkeleton() {
  return (
    <div className="flex gap-3.5 border-b border-border px-4 py-3.5">
      <div className="flex w-14 shrink-0 flex-col gap-2 pt-0.5">
        <Bar className="w-full" />
        <Bar className="w-2/3" />
      </div>
      <div className="flex min-w-0 flex-1 flex-col gap-2.5">
        <Bar className="w-[70%]" />
        <Bar className="w-[35%]" />
        <Bar className="w-1/4" />
      </div>
    </div>
  );
}

function AdminCardSkeleton() {
  return (
    <div className="flex flex-col gap-2.5 border-b border-border px-4 py-4">
      <Bar className="w-[60%]" />
      <Bar className="w-[40%]" />
      <Bar className="w-1/4" />
    </div>
  );
}

function RunRowSkeleton() {
  return (
    <div className="flex flex-col gap-2.5 rounded-control border border-border bg-surface p-4">
      <Bar className="w-1/2" />
      <Bar className="w-[80%]" />
      <Bar className="w-1/3" />
    </div>
  );
}

function DayHeaderSkeleton() {
  return (
    <div className="flex items-center border-b border-border bg-surface px-4 py-3">
      <Bar className="w-28" />
    </div>
  );
}

// Placeholder rows with the shape of the loaded list, so nothing moves when the data arrives.
// Mount it only while loading: it appears after a short delay, so a quick response shows nothing.
export default function Skeleton({ variant, label, count = 3 }: SkeletonProps) {
  const show = useAfterDelay();
  if (!show) return null;

  const rows = Array.from({ length: count }, (_, index) => {
    switch (variant) {
      case 'eventCard':
        // a day header before every third card
        return (
          <div key={index}>
            {index % 3 === 0 && <DayHeaderSkeleton />}
            <EventCardSkeleton />
          </div>
        );
      case 'adminCard':
        return <AdminCardSkeleton key={index} />;
      case 'runRow':
        return <RunRowSkeleton key={index} />;
    }
  });

  return (
    <div aria-busy="true">
      <p role="status" className="sr-only">
        {label}
      </p>
      <div aria-hidden="true" className={`skeleton-pulse ${variant === 'runRow' ? 'space-y-3' : ''}`}>
        {rows}
      </div>
    </div>
  );
}
