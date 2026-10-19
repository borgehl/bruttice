import type { ReactNode } from 'react';

interface FilterGroupProps {
  title: string;
  // how many are selected, shown next to the title when there are any
  count: number;
  // a switch for how the group applies (Include/Exclude, Max/Min)
  mode?: ReactNode;
  hint?: string;
  children: ReactNode;
}

export default function FilterGroup({ title, count, mode, hint, children }: FilterGroupProps) {
  return (
    <section className="border-b border-border px-4 py-4">
      <div className="mb-3 flex min-h-9 items-center justify-between gap-3">
        <h2 className="font-label flex items-baseline gap-2 text-base font-semibold text-text">
          {title}
          {count > 0 && <span className="font-data text-xs font-normal text-text-faint">{count} selected</span>}
        </h2>
        {mode}
      </div>
      {hint && <p className="-mt-1 mb-3 text-[13px] text-text-faint">{hint}</p>}
      {children}
    </section>
  );
}
