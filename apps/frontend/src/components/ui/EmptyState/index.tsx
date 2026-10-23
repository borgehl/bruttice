import type { ReactNode } from 'react';

interface EmptyStateProps {
  title: string;
  children?: ReactNode;
  // the one thing to do next, if there is one
  action?: ReactNode;
}

// Nothing to show, and why. Neutral on purpose: an empty list is not a failure.
export default function EmptyState({ title, children, action }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center gap-3 px-4 py-12 text-center">
      <div
        className="flex h-14 w-14 items-center justify-center rounded-full border-2 border-border text-text-faint"
        aria-hidden="true"
      >
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <rect x="4" y="5" width="16" height="15" rx="2" />
          <path d="M4 10h16M9 3v4M15 3v4" />
        </svg>
      </div>
      <h2 className="font-label text-lg font-bold text-text">{title}</h2>
      {children && <p className="max-w-sm text-[15px] leading-relaxed text-text-muted">{children}</p>}
      {action && <div className="mt-1 w-full max-w-xs">{action}</div>}
    </div>
  );
}
