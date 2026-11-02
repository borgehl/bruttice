import { useState, type FormEvent } from 'react';
import type { ConfigRegion, ConfigVenue } from '../../../types/config';
import { changeAdminConfig } from '../../../services/api';
import { count } from '../../../utils/config';
import Field, { fieldClass } from '../../ui/Field';
import EmptyState from '../../ui/EmptyState';
import { ChangeStatus, List, Meta, Row } from './shared';
import { dangerButtonClass, smallButtonClass, solidButtonClass } from './styles';
import { useChanger, type Changer } from './useChanger';

interface VenueFormProps {
  initial: { name: string; link: string; region: string };
  regions: ConfigRegion[];
  submitLabel: string;
  idPrefix: string;
  changer: Changer;
  onSave: (venue: { name: string; link: string; region: string }) => Promise<boolean>;
  onCancel?: () => void;
}

function VenueForm({ initial, regions, submitLabel, idPrefix, changer, onSave, onCancel }: VenueFormProps) {
  const [venue, setVenue] = useState(initial);
  const valid = venue.name.trim() !== '' && venue.region !== '';

  const submit = async (event: FormEvent) => {
    event.preventDefault();
    if (valid && (await onSave(venue)) && !onCancel) setVenue(initial);
  };

  return (
    <form onSubmit={(e) => void submit(e)} className="space-y-3 rounded-control border border-border bg-bg p-3">
      <Field label="Name" htmlFor={`${idPrefix}-name`}>
        <input id={`${idPrefix}-name`} type="text" maxLength={100} value={venue.name} onChange={(e) => setVenue({ ...venue, name: e.target.value })} className={`${fieldClass} h-11`} />
      </Field>
      <Field label="Website" htmlFor={`${idPrefix}-link`} hint="Optional. Starts with https://">
        <input id={`${idPrefix}-link`} type="url" maxLength={500} value={venue.link} onChange={(e) => setVenue({ ...venue, link: e.target.value })} className={`${fieldClass} h-11`} />
      </Field>
      <Field label="Region" htmlFor={`${idPrefix}-region`} hint="Events at this venue are put in this region.">
        <select id={`${idPrefix}-region`} value={venue.region} onChange={(e) => setVenue({ ...venue, region: e.target.value })} className={`${fieldClass} h-11`}>
          <option value="">Choose a region</option>
          {regions.map((r) => (
            <option key={r.id} value={r.name}>
              {r.name}
            </option>
          ))}
        </select>
      </Field>
      <div className="flex gap-2">
        <button type="submit" disabled={changer.busy || !valid} className={solidButtonClass}>
          {submitLabel}
        </button>
        {onCancel && (
          <button type="button" onClick={onCancel} className={smallButtonClass}>
            Cancel
          </button>
        )}
      </div>
    </form>
  );
}

interface VenuesPanelProps {
  venues: ConfigVenue[];
  regions: ConfigRegion[];
  reload: () => void;
}

export default function VenuesPanel({ venues, regions, reload }: VenuesPanelProps) {
  const changer = useChanger(reload);
  const [adding, setAdding] = useState(false);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [deletingId, setDeletingId] = useState<number | null>(null);
  const active = regions.filter((r) => !r.retired);

  return (
    <div>
      <p className="mb-4 text-sm text-text-muted">
        Places events happen. A venue that a scraper reports is kept up to date by it, so it can only be changed in the scraper. A venue with events can&apos;t
        be deleted.
      </p>
      {adding ? (
        <div className="mb-4">
          <VenueForm
            initial={{ name: '', link: '', region: '' }}
            regions={active}
            submitLabel="Add venue"
            idPrefix="new-venue"
            changer={changer}
            onSave={(venue) => changer.run(() => changeAdminConfig('POST', 'venues', venue)).then((ok) => (ok && setAdding(false), ok))}
            onCancel={() => setAdding(false)}
          />
        </div>
      ) : (
        <button type="button" onClick={() => setAdding(true)} className={`${smallButtonClass} mb-4`}>
          New venue
        </button>
      )}
      <ChangeStatus changer={changer} />

      {venues.length === 0 ? (
        <EmptyState title="No venues">Add the first one above.</EmptyState>
      ) : (
        <List>
          {venues.map((venue) => (
            <Row key={venue.id}>
              <div className="flex flex-wrap items-center justify-between gap-2">
                <span>
                  <span className="text-sm font-medium text-text">{venue.name}</span>{' '}
                  <Meta>
                    {venue.region} · {count(venue.events, 'event')}
                    {venue.scraped && ' · scraper'}
                  </Meta>
                </span>
                {!venue.scraped && (
                  <span className="flex gap-2">
                    <button type="button" onClick={() => setEditingId(editingId === venue.id ? null : venue.id)} className={smallButtonClass}>
                      Edit
                    </button>
                    <button
                      type="button"
                      disabled={venue.events > 0}
                      title={venue.events > 0 ? 'A venue with events can’t be deleted' : undefined}
                      onClick={() => setDeletingId(deletingId === venue.id ? null : venue.id)}
                      className={dangerButtonClass}
                    >
                      Delete
                    </button>
                  </span>
                )}
              </div>
              {editingId === venue.id && (
                <VenueForm
                  initial={{ name: venue.name, link: venue.link, region: venue.region }}
                  regions={active}
                  submitLabel="Save"
                  idPrefix={`venue-${venue.id}`}
                  changer={changer}
                  onSave={(next) => changer.run(() => changeAdminConfig('PATCH', `venues/${venue.id}`, next)).then((ok) => (ok && setEditingId(null), ok))}
                  onCancel={() => setEditingId(null)}
                />
              )}
              {deletingId === venue.id && (
                <div className="flex flex-wrap items-center gap-2 rounded-control border border-border bg-bg p-3">
                  <p className="flex-1 text-sm text-text-muted">
                    Delete <span className="font-medium text-text">{venue.name}</span>?
                  </p>
                  <button
                    type="button"
                    disabled={changer.busy}
                    onClick={() => void changer.run(() => changeAdminConfig('DELETE', `venues/${venue.id}`)).then((ok) => ok && setDeletingId(null))}
                    className={dangerButtonClass}
                  >
                    Delete venue
                  </button>
                  <button type="button" onClick={() => setDeletingId(null)} className={smallButtonClass}>
                    Cancel
                  </button>
                </div>
              )}
            </Row>
          ))}
        </List>
      )}
    </div>
  );
}
