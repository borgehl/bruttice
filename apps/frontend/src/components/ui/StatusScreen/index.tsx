import { useEffect, useRef, type ReactNode } from 'react';

interface StatusScreenProps {
  kind: 'success' | 'error';
  title: string;
  children: ReactNode;
  // what happens next, shown as a numbered list, the first step being the one to do now
  steps?: string[];
  action: ReactNode;
}

// The outcome of a submission or a verification link: an icon, what happened, what comes next.
export default function StatusScreen({ kind, title, children, steps, action }: StatusScreenProps) {
  const ok = kind === 'success';
  const heading = useRef<HTMLHeadingElement>(null);

  // an error replaces the view, so put focus on what happened
  useEffect(() => {
    if (!ok) heading.current?.focus();
  }, [ok]);

  return (
    <div role={ok ? undefined : 'alert'} className="flex flex-col items-center gap-4 py-6 text-center">
      <div
        className={`flex h-16 w-16 items-center justify-center rounded-full border-2 shadow-[0_0_24px_color-mix(in_srgb,currentColor_35%,transparent)] ${
          ok ? 'border-success text-success' : 'border-danger text-danger'
        }`}
        aria-hidden="true"
      >
        <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round">
          <path d={ok ? 'M5 12l5 5 9-10' : 'M6 6l12 12M18 6L6 18'} />
        </svg>
      </div>
      <h2 ref={heading} tabIndex={-1} className="font-label text-xl font-bold text-text outline-none">
        {title}
      </h2>
      <p className="max-w-sm text-[15px] leading-relaxed text-text-muted">{children}</p>

      {steps && steps.length > 0 && (
        <ol className="mt-2 w-full max-w-sm text-left">
          {steps.map((step, index) => (
            <li key={step} className="flex items-center gap-3.5 border-b border-border py-3">
              <span className={`font-data text-sm ${index === 0 ? 'text-accent-2' : 'text-text-faint'}`}>
                {String(index + 1).padStart(2, '0')}
              </span>
              <span className={`text-sm ${index === 0 ? 'font-medium text-text' : 'text-text-muted'}`}>{step}</span>
            </li>
          ))}
        </ol>
      )}

      <div className="mt-2 w-full max-w-sm">{action}</div>
    </div>
  );
}
