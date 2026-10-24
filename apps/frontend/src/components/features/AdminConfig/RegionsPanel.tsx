import { useState } from 'react';
import type { ConfigRegion } from '../../../types/config';
import { changeAdminConfig } from '../../../services/api';
import { count, regionUse } from '../../../utils/config';
import { fieldClass } from '../../ui/Field';
import EmptyState from '../../ui/EmptyState';
import { AddForm, ChangeStatus, List, Meta, RenameForm, Row } from './shared';
import { dangerButtonClass, smallButtonClass } from './styles';
import { useChanger } from './useChanger';

interface RetireFormProps {
  region: ConfigRegion;
  others: ConfigRegion[];
  changer: ReturnType<typeof useChanger>;
  onDone: () => void;
}

// How a region is retired: its events and venues move to another region, or its events are deleted.
function RetireForm({ region, others, changer, onDone }: RetireFormProps) {
  const [choice, setChoice] = useState<'move' | 'delete'>('move');
  const [targetId, setTargetId] = useState<number>(others[0]?.id ?? 0);
  const [confirmed, setConfirmed] = useState(false);
  const target = others.find((r) => r.id === targetId);
  const deleteBlocked = region.venues > 0;

  const retire = async () => {
    const body = choice === 'move' ? { moveToId: targetId } : { deleteEvents: true };
    if (await changer.run(() => changeAdminConfig('POST', `regions/${region.id}/retire`, body), target?.name)) onDone();
  };

  const canRetire = choice === 'move' ? Boolean(target) : confirmed && !deleteBlocked;

  return (
    <div className="space-y-3 rounded-control border border-border bg-bg p-3">
      <p className="text-sm text-text-muted">
        Retiring <span className="font-medium text-text">{region.name}</span> takes it out of the site&apos;s filters and forms.
        The site has {regionUse(region.events, region.venues)} in it.
      </p>

      <fieldset className="space-y-2">
        <legend className="sr-only">What happens to the events of {region.name}</legend>
        <label className="flex cursor-pointer items-start gap-2 text-sm text-text">
          <input type="radio" name={`retire-${region.id}`} checked={choice === 'move'} onChange={() => setChoice('move')} className="mt-1 accent-[var(--color-accent)]" />
          <span>
            Move its events and venues to another region
            <span className="block text-xs text-text-faint">
              The name {region.name} keeps working: a scraper that still sends it lands in the new region.
            </span>
          </span>
        </label>
        {choice === 'move' && (
          <div className="pl-6">
            <label htmlFor={`target-${region.id}`} className="sr-only">
              Region to move to
            </label>
            <select id={`target-${region.id}`} value={targetId} onChange={(e) => setTargetId(Number(e.target.value))} className={`${fieldClass} h-10`}>
              {others.map((r) => (
                <option key={r.id} value={r.id}>
                  {r.name}
                </option>
              ))}
            </select>
          </div>
        )}

        <label className={`flex items-start gap-2 text-sm ${deleteBlocked ? 'cursor-not-allowed opacity-60' : 'cursor-pointer'} text-text`}>
          <input
            type="radio"
            name={`retire-${region.id}`}
            checked={choice === 'delete'}
            disabled={deleteBlocked}
            onChange={() => setChoice('delete')}
            className="mt-1 accent-[var(--color-accent)]"
          />
          <span>
            Delete its {count(region.events, 'event')}
            {deleteBlocked && (
              <span className="block text-xs text-text-faint">
                {count(region.venues, 'venue')} default to this region. Move them to another region first.
              </span>
            )}
          </span>
        </label>
        {choice === 'delete' && !deleteBlocked && (
          <label className="flex cursor-pointer items-start gap-2 pl-6 text-sm text-danger">
            <input type="checkbox" checked={confirmed} onChange={(e) => setConfirmed(e.target.checked)} className="mt-1 accent-[var(--color-danger)]" />
            <span>I understand these events are deleted and no new region takes them.</span>
          </label>
        )}
      </fieldset>

      <div className="flex gap-2">
        <button type="button" onClick={() => void retire()} disabled={changer.busy || !canRetire} className={dangerButtonClass}>
          {choice === 'move' ? `Retire and move to ${target?.name ?? '…'}` : 'Retire and delete events'}
        </button>
        <button type="button" onClick={onDone} className={smallButtonClass}>
          Cancel
        </button>
      </div>
    </div>
  );
}

interface RegionsPanelProps {
  regions: ConfigRegion[];
  reload: () => void;
}

export default function RegionsPanel({ regions, reload }: RegionsPanelProps) {
  const changer = useChanger(reload);
  const [editing, setEditing] = useState<{ id: number; what: 'rename' | 'retire' } | null>(null);
  const active = regions.filter((r) => !r.retired);
  const retired = regions.filter((r) => r.retired);

  return (
    <div>
      <p className="mb-4 text-sm text-text-muted">
        The areas events are in. Retiring a region needs a decision about its events: move them to another region, or delete them.
      </p>
      <AddForm label="New region" placeholder="Name" changer={changer} onAdd={(name) => changeAdminConfig('POST', 'regions', { name })} />
      <ChangeStatus changer={changer} />

      {active.length === 0 ? (
        <EmptyState title="No regions">Add the first one above.</EmptyState>
      ) : (
        <List>
          {active.map((region) => {
            const mode = editing?.id === region.id ? editing.what : null;
            return (
              <Row key={region.id}>
                {mode === 'rename' ? (
                  <RenameForm
                    label="Region name"
                    initial={region.name}
                    changer={changer}
                    onSave={(name) => changeAdminConfig('PATCH', `regions/${region.id}`, { name })}
                    onDone={() => setEditing(null)}
                  />
                ) : (
                  <div className="flex flex-wrap items-center justify-between gap-2">
                    <span>
                      <span className="text-sm font-medium text-text">{region.name}</span>{' '}
                      <Meta>{regionUse(region.events, region.venues)}</Meta>
                    </span>
                    <span className="flex gap-2">
                      <button type="button" onClick={() => setEditing({ id: region.id, what: 'rename' })} className={smallButtonClass}>
                        Rename
                      </button>
                      <button
                        type="button"
                        onClick={() => setEditing({ id: region.id, what: 'retire' })}
                        disabled={active.length < 2}
                        title={active.length < 2 ? 'At least one region has to stay' : undefined}
                        className={dangerButtonClass}
                      >
                        Retire
                      </button>
                    </span>
                  </div>
                )}
                {mode === 'retire' && (
                  <RetireForm region={region} others={active.filter((r) => r.id !== region.id)} changer={changer} onDone={() => setEditing(null)} />
                )}
              </Row>
            );
          })}
        </List>
      )}

      {retired.length > 0 && (
        <div className="mt-6">
          <h3 className="font-label mb-2 text-sm font-semibold text-text-muted">Retired</h3>
          <List>
            {retired.map((region) => (
              <Row key={region.id}>
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <span>
                    <span className="text-sm text-text-muted">{region.name}</span>{' '}
                    <Meta>{region.mergedInto ? `now ${region.mergedInto}` : 'events deleted'}</Meta>
                  </span>
                  <button
                    type="button"
                    disabled={changer.busy}
                    onClick={() => void changer.run(() => changeAdminConfig('POST', 'regions', { name: region.name }))}
                    className={smallButtonClass}
                  >
                    Restore
                  </button>
                </div>
              </Row>
            ))}
          </List>
          <p className="mt-2 text-xs text-text-faint">Restoring a region puts it back in use. The events that moved away stay where they went.</p>
        </div>
      )}
    </div>
  );
}
