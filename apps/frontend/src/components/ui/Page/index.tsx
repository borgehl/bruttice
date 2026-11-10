import type { ReactNode } from 'react';

interface PageProps {
  title: string;
  subtitle?: string;
  children: ReactNode;
  // pinned under the scrolling content, for the main action
  footer?: ReactNode;
  // no padding around the content, for pages made of full-width sections
  flush?: boolean;
  // a control at the right of the title, such as Logout
  action?: ReactNode;
  // pinned between the title and the scrolling content, such as the admin pages' menu
  subheader?: ReactNode;
}

// A scrolling page under a fixed title, for everything except the calendar, which scrolls itself.
export default function Page({ title, subtitle, children, footer, flush, action, subheader }: PageProps) {
  return (
    <div className="flex min-h-0 flex-1 flex-col">
      <header className="shrink-0 border-b border-border px-4 pb-3.5 pt-4">
        <div className="flex items-center justify-between gap-3">
          <h1 className="font-label text-2xl font-bold leading-tight text-text">{title}</h1>
          {action}
        </div>
        {subtitle && <p className="mt-1 text-sm text-text-muted">{subtitle}</p>}
      </header>
      {subheader && <div className="shrink-0">{subheader}</div>}
      <div className={`min-h-0 flex-1 overflow-y-auto ${flush ? '' : 'px-4 py-4'}`}>{children}</div>
      {footer && <div className="shrink-0 border-t border-border bg-bg p-4">{footer}</div>}
    </div>
  );
}
