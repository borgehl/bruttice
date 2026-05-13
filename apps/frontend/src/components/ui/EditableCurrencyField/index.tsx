import { useState, useEffect, useRef } from 'react';

interface EditableCurrencyFieldProps {
  value: string;
  onChange: (value: string) => void;
  className?: string;
  placeholder?: string;
  variant?: 'inline' | 'form';
}

export default function EditableCurrencyField({ value, onChange, className = '', placeholder, variant = 'inline' }: EditableCurrencyFieldProps) {
  const [localValue, setLocalValue] = useState(value);
  const [validationError, setValidationError] = useState<string | null>(null);
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    setLocalValue(value);
    setValidationError(null);
  }, [value]);

  const validate = (val: string): boolean => {
    if (val === '') return true;
    return /^\d+(\.\d{1,2})?$/.test(val) && parseFloat(val) >= 0;
  };

  const handleChange = (newValue: string) => {
    setLocalValue(newValue);

    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
    }

    if (!validate(newValue)) {
      setValidationError('Enter a valid amount (e.g. 10.00)');
      return;
    }

    setValidationError(null);

    timeoutRef.current = setTimeout(() => {
      if (newValue !== value) {
        const formatted = newValue === '' ? '' : parseFloat(newValue).toFixed(2);
        onChange(formatted);
      }
    }, 500);
  };

  useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
    };
  }, []);

  const inputClass = variant === 'form'
    ? `w-full px-3 py-2 bg-surface border ${validationError ? 'border-error' : 'border-border'} rounded-md text-text placeholder-text-dim focus:outline-none focus:ring-2 focus:ring-primary`
    : `bg-transparent border ${validationError ? 'border-error' : 'border-transparent hover:border-border'} focus:border-primary focus:outline-none rounded px-1 -mx-1`;

  return (
    <div className={className}>
      <input
        type="text"
        inputMode="decimal"
        value={localValue}
        onChange={(e) => handleChange(e.target.value)}
        placeholder={placeholder ?? '0.00'}
        className={inputClass}
      />
      {validationError && <p className="text-xs text-error mt-1">{validationError}</p>}
    </div>
  );
}
