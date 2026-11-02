import { type InputHTMLAttributes } from 'react';

type DateInputType = 'date' | 'time';

interface DateInputProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'type' | 'onClick'> {
  type?: DateInputType;
}

export default function DateInput({
  type = 'date',
  className = '',
  ...props
}: DateInputProps) {
  const baseClasses = 'px-2 py-1.5 text-sm bg-surface border border-border rounded-control focus:outline-none focus:border-accent cursor-pointer w-auto';
  const combinedClasses = className ? `${baseClasses} ${className}` : baseClasses;

  return (
    <input
      type={type}
      onClick={(e) => e.currentTarget.showPicker()}
      className={combinedClasses}
      {...props}
    />
  );
}
