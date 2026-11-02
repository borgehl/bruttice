import { useState, type FormEvent, type ReactNode } from 'react';
import type { ConfigChange } from '../../../types/config';
import { smallButtonClass, solidButtonClass } from './styles';
import type { Changer } from './useChanger';
import InlineError from '../../ui/InlineError';
import Spinner from '../../ui/Spinner';
import { fieldClass } from '../../ui/Field';

export function ChangeStatus({ changer }: { changer: Changer }) {
  return (
    <>
      {changer.error && <InlineError message={changer.error} className="mb-3" />}
      <p role="status" className="mb-3 min-h-5 text-sm text-success">
        {changer.notice}
      </p>
    </>
  );
}

interface AddFormProps {
  label: string;
  placeholder: string;
  changer: Changer;
  // adds the thing; the form is cleared if it worked
  onAdd: (name: string) => Promise<ConfigChange>;
  children?: ReactNode;
}

// A name and an Add button, with room for more fields between them.
export function AddForm({ label, placeholder, changer, onAdd, children }: AddFormProps) {
  const [name, setName] = useState('');

  const submit = async (event: FormEvent) => {
    event.preventDefault();
    if (!name.trim()) return;
    if (await changer.run(() => onAdd(name))) setName('');
  };

  return (
    <form onSubmit={(e) => void submit(e)} className="mb-4 flex flex-wrap items-end gap-2">
      <div className="min-w-40 flex-1">
        <label htmlFor={`add-${label}`} className="mb-1.5 block text-[13px] font-medium text-text-muted">
          {label}
        </label>
        <input
          id={`add-${label}`}
          type="text"
          maxLength={100}
          value={name}
          placeholder={placeholder}
          onChange={(e) => setName(e.target.value)}
          className={`${fieldClass} h-11`}
        />
      </div>
      {children}
      <button type="submit" disabled={changer.busy || !name.trim()} className={`${solidButtonClass} h-11`}>
        {changer.busy ? <Spinner size="sm" /> : 'Add'}
      </button>
    </form>
  );
}

interface RenameFormProps {
  label: string;
  initial: string;
  changer: Changer;
  onSave: (name: string) => Promise<ConfigChange>;
  onDone: () => void;
  children?: ReactNode;
}

// Edits the name of one row in place.
export function RenameForm({ label, initial, changer, onSave, onDone, children }: RenameFormProps) {
  const [name, setName] = useState(initial);

  const submit = async (event: FormEvent) => {
    event.preventDefault();
    if (!name.trim() || name.trim() === initial) {
      onDone();
      return;
    }
    if (await changer.run(() => onSave(name))) onDone();
  };

  return (
    <form onSubmit={(e) => void submit(e)} className="flex flex-wrap items-end gap-2">
      <div className="min-w-40 flex-1">
        <label htmlFor={`rename-${label}-${initial}`} className="sr-only">
          {label}
        </label>
        <input
          id={`rename-${label}-${initial}`}
          type="text"
          maxLength={100}
          value={name}
          onChange={(e) => setName(e.target.value)}
          className={`${fieldClass} h-10`}
          autoFocus
        />
      </div>
      {children}
      <button type="submit" disabled={changer.busy} className={solidButtonClass}>
        Save
      </button>
      <button type="button" onClick={onDone} className={smallButtonClass}>
        Cancel
      </button>
    </form>
  );
}

export function Row({ children }: { children: ReactNode }) {
  return <li className="space-y-2 border-b border-border px-4 py-3 last:border-b-0">{children}</li>;
}

export function List({ children }: { children: ReactNode }) {
  return <ul className="rounded-control border border-border bg-surface">{children}</ul>;
}

export function Meta({ children }: { children: ReactNode }) {
  return <span className="font-data text-xs text-text-faint">{children}</span>;
}
