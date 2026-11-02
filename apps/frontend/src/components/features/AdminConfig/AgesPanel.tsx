import { useState } from 'react';
import type { ConfigAge } from '../../../types/config';
import { changeAdminConfig } from '../../../services/api';
import { count } from '../../../utils/config';
import { fieldClass } from '../../ui/Field';
import EmptyState from '../../ui/EmptyState';
import { AddForm, ChangeStatus, List, Meta, RenameForm, Row } from './shared';
import { dangerButtonClass, smallButtonClass } from './styles';
import { useChanger } from './useChanger';

interface AgeValueProps {
  id: string;
  value: string;
  onChange: (value: string) => void;
}

function AgeValue({ id, value, onChange }: AgeValueProps) {
  return (
    <div className="w-24">
      <label htmlFor={id} className="mb-1.5 block text-[13px] font-medium text-text-muted">
        Age
      </label>
      <input id={id} type="number" min={0} max={255} value={value} onChange={(e) => onChange(e.target.value)} className={`${fieldClass} h-11`} />
    </div>
  );
}

interface AgesPanelProps {
  ages: ConfigAge[];
  reload: () => void;
}

export default function AgesPanel({ ages, reload }: AgesPanelProps) {
  const changer = useChanger(reload);
  const [newValue, setNewValue] = useState('18');
  const [editing, setEditing] = useState<{ id: number; value: string } | null>(null);
  const [deleting, setDeleting] = useState<{ id: number; replacementId: number } | null>(null);

  return (
    <div>
      <p className="mb-4 text-sm text-text-muted">
        Age restrictions. The age is the youngest person allowed. Events and scrapers pick an age group by its name, so renaming one makes a scraper that
        still sends the old name have its events refused until it is updated.
      </p>
      <AddForm
        label="New age group"
        placeholder="Name, for example 18+"
        changer={changer}
        onAdd={(name) => changeAdminConfig('POST', 'ages', { name, value: Number(newValue) })}
      >
        <AgeValue id="new-age-value" value={newValue} onChange={setNewValue} />
      </AddForm>
      <ChangeStatus changer={changer} />

      {ages.length === 0 ? (
        <EmptyState title="No age groups">Add the first one above.</EmptyState>
      ) : (
        <List>
          {ages.map((age) => {
            const others = ages.filter((a) => a.id !== age.id);
            return (
              <Row key={age.id}>
                {editing?.id === age.id ? (
                  <RenameForm
                    label="Age group name"
                    initial={age.name}
                    changer={changer}
                    // the age is saved with the name, even when only it changed
                    onSave={(name) => changeAdminConfig('PATCH', `ages/${age.id}`, { name, value: Number(editing.value) })}
                    onDone={() => setEditing(null)}
                  >
                    <AgeValue id={`age-value-${age.id}`} value={editing.value} onChange={(value) => setEditing({ id: age.id, value })} />
                  </RenameForm>
                ) : (
                  <div className="flex flex-wrap items-center justify-between gap-2">
                    <span>
                      <span className="text-sm font-medium text-text">{age.name}</span> <Meta>age {age.value} · {count(age.events, 'event')}</Meta>
                    </span>
                    <span className="flex gap-2">
                      <button type="button" onClick={() => setEditing({ id: age.id, value: String(age.value) })} className={smallButtonClass}>
                        Edit
                      </button>
                      <button
                        type="button"
                        disabled={others.length === 0}
                        title={others.length === 0 ? 'At least one age group has to stay' : undefined}
                        onClick={() => setDeleting(deleting?.id === age.id ? null : { id: age.id, replacementId: others[0]?.id ?? 0 })}
                        className={dangerButtonClass}
                      >
                        Delete
                      </button>
                    </span>
                  </div>
                )}
                {deleting?.id === age.id && (
                  <div className="space-y-2 rounded-control border border-border bg-bg p-3">
                    {age.events > 0 ? (
                      <>
                        <p className="text-sm text-text-muted">
                          {count(age.events, 'event')} have <span className="font-medium text-text">{age.name}</span>. Pick the age group they get instead.
                        </p>
                        <label htmlFor={`replace-${age.id}`} className="sr-only">
                          Age group to move to
                        </label>
                        <select
                          id={`replace-${age.id}`}
                          value={deleting.replacementId}
                          onChange={(e) => setDeleting({ id: age.id, replacementId: Number(e.target.value) })}
                          className={`${fieldClass} h-10`}
                        >
                          {others.map((a) => (
                            <option key={a.id} value={a.id}>
                              {a.name}
                            </option>
                          ))}
                        </select>
                      </>
                    ) : (
                      <p className="text-sm text-text-muted">
                        No event has <span className="font-medium text-text">{age.name}</span>. Delete it?
                      </p>
                    )}
                    <div className="flex gap-2">
                      <button
                        type="button"
                        disabled={changer.busy}
                        onClick={() => {
                          const query = age.events > 0 ? `?replacementId=${deleting.replacementId}` : '';
                          const to = others.find((a) => a.id === deleting.replacementId)?.name;
                          void changer.run(() => changeAdminConfig('DELETE', `ages/${age.id}${query}`), to).then((ok) => ok && setDeleting(null));
                        }}
                        className={dangerButtonClass}
                      >
                        Delete age group
                      </button>
                      <button type="button" onClick={() => setDeleting(null)} className={smallButtonClass}>
                        Cancel
                      </button>
                    </div>
                  </div>
                )}
              </Row>
            );
          })}
        </List>
      )}
    </div>
  );
}
