import type { ReactNode } from 'react';

// Classes shared by every text input, select and date input in the forms.
export const fieldClass =
  'h-12 w-full rounded-control border border-border bg-raised px-3 text-text placeholder-text-faint focus:border-accent focus:outline-none';

interface FieldProps {
  label: string;
  htmlFor?: string;
  hint?: string;
  children: ReactNode;
}

export default function Field({ label, htmlFor, hint, children }: FieldProps) {
  return (
    <div>
      <label htmlFor={htmlFor} className="mb-1.5 block text-[13px] font-medium text-text-muted">
        {label}
      </label>
      {children}
      {hint && <p className="mt-1.5 text-xs text-text-faint">{hint}</p>}
    </div>
  );
}

// The main button of a form or page.
export const primaryButtonClass =
  'font-label chamfer flex h-12 w-full cursor-pointer items-center justify-center rounded-control bg-accent text-sm font-semibold text-accent-ink enabled:hover:bg-accent-hover disabled:cursor-not-allowed disabled:opacity-50';

// An outlined button, for the secondary choice.
export const ghostButtonClass =
  'font-label flex h-12 w-full cursor-pointer items-center justify-center rounded-control border border-accent text-sm font-semibold text-accent hover:text-accent-hover';
