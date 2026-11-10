import { useState, type ChangeEvent, type FormEvent } from 'react';
import type { AdminSponsor, SponsorForm } from '../../../types/sponsor';
import { changeAdminSponsor } from '../../../services/api';
import { count } from '../../../utils/config';
import { imageProblem, readImage } from '../../../utils/sponsor';
import Field, { fieldClass } from '../../ui/Field';
import EmptyState from '../../ui/EmptyState';
import { ChangeStatus, List, Meta, Row } from '../AdminConfig/shared';
import { dangerButtonClass, smallButtonClass, solidButtonClass } from '../AdminConfig/styles';
import { useChanger, type Changer } from '../AdminConfig/useChanger';

const EMPTY: SponsorForm = { name: '', text: '', link: '', startsOn: '', endsOn: '', enabled: true };

interface SponsorFormProps {
  initial: SponsorForm;
  // the picture the sponsor has now, empty if none
  currentImageUrl: string;
  submitLabel: string;
  idPrefix: string;
  changer: Changer;
  onSave: (sponsor: SponsorForm) => Promise<boolean>;
  onCancel?: () => void;
}

function SponsorFormFields({ initial, currentImageUrl, submitLabel, idPrefix, changer, onSave, onCancel }: SponsorFormProps) {
  const [sponsor, setSponsor] = useState(initial);
  // a picture chosen in this form, as base64
  const [image, setImage] = useState<string | undefined>(undefined);
  const [removeImage, setRemoveImage] = useState(false);
  const [imageError, setImageError] = useState('');
  const valid = sponsor.name.trim() !== '' && sponsor.link.trim() !== '';

  const choose = async (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    setImage(undefined);
    setImageError('');
    if (!file) return;
    const problem = imageProblem(file);
    if (problem) {
      setImageError(problem);
      event.target.value = '';
      return;
    }
    try {
      setImage(await readImage(file));
      setRemoveImage(false);
    } catch {
      setImageError('The picture could not be read.');
    }
  };

  const submit = async (event: FormEvent) => {
    event.preventDefault();
    if (!valid) return;
    const saved = await onSave({ ...sponsor, ...(image !== undefined ? { image } : {}), ...(removeImage ? { removeImage: true } : {}) });
    if (saved && !onCancel) {
      setSponsor(initial);
      setImage(undefined);
      setRemoveImage(false);
    }
  };

  return (
    <form onSubmit={(e) => void submit(e)} className="space-y-3 rounded-control border border-border bg-bg p-3">
      <Field label="Name" htmlFor={`${idPrefix}-name`}>
        <input id={`${idPrefix}-name`} type="text" maxLength={100} value={sponsor.name} onChange={(e) => setSponsor({ ...sponsor, name: e.target.value })} className={`${fieldClass} h-11`} />
      </Field>
      <Field label="Text" htmlFor={`${idPrefix}-text`} hint="Optional. One short line under the name.">
        <input id={`${idPrefix}-text`} type="text" maxLength={160} value={sponsor.text} onChange={(e) => setSponsor({ ...sponsor, text: e.target.value })} className={`${fieldClass} h-11`} />
      </Field>
      <Field label="Link" htmlFor={`${idPrefix}-link`} hint="Where a click goes. Starts with https://">
        <input id={`${idPrefix}-link`} type="url" maxLength={500} value={sponsor.link} onChange={(e) => setSponsor({ ...sponsor, link: e.target.value })} className={`${fieldClass} h-11`} />
      </Field>
      <div className="grid gap-3 sm:grid-cols-2">
        <Field label="First day" htmlFor={`${idPrefix}-starts`} hint="Optional. Empty means it starts right away.">
          <input id={`${idPrefix}-starts`} type="date" value={sponsor.startsOn} onChange={(e) => setSponsor({ ...sponsor, startsOn: e.target.value })} className={`${fieldClass} h-11`} />
        </Field>
        <Field label="Last day" htmlFor={`${idPrefix}-ends`} hint="Optional. Empty means it has no end.">
          <input id={`${idPrefix}-ends`} type="date" value={sponsor.endsOn} onChange={(e) => setSponsor({ ...sponsor, endsOn: e.target.value })} className={`${fieldClass} h-11`} />
        </Field>
      </div>
      <Field label="Picture" htmlFor={`${idPrefix}-image`} hint="Optional. PNG, JPEG, GIF or WebP, at most 200 KB. It is shown small, about 64 × 88 pixels, so use a tall or square image.">
        <input id={`${idPrefix}-image`} type="file" accept="image/png,image/jpeg,image/gif,image/webp" onChange={(e) => void choose(e)} className="block w-full text-sm text-text-muted" />
      </Field>
      {imageError && <p role="alert" className="text-sm text-danger">{imageError}</p>}
      {currentImageUrl && image === undefined && !removeImage && (
        <div className="flex items-center gap-3">
          <img src={currentImageUrl} alt="" width={64} height={88} className="h-[88px] w-16 rounded-control border border-border object-cover" />
          <button type="button" onClick={() => setRemoveImage(true)} className={smallButtonClass}>
            Remove picture
          </button>
        </div>
      )}
      {removeImage && <p className="text-sm text-text-muted">The picture is removed when you save.</p>}
      <label className="flex items-center gap-2 text-sm text-text">
        <input type="checkbox" checked={sponsor.enabled} onChange={(e) => setSponsor({ ...sponsor, enabled: e.target.checked })} />
        Show this card (inside its days)
      </label>
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

// "2026-10-22 to 2026-11-05", "from 2026-10-22", "until 2026-11-05" or "no dates set"
function schedule(sponsor: AdminSponsor): string {
  if (sponsor.startsOn && sponsor.endsOn) return `${sponsor.startsOn} to ${sponsor.endsOn}`;
  if (sponsor.startsOn) return `from ${sponsor.startsOn}`;
  if (sponsor.endsOn) return `until ${sponsor.endsOn}`;
  return 'no dates set';
}

interface AdminSponsorsProps {
  sponsors: AdminSponsor[];
  reload: () => void;
}

// The sponsor cards: one is shown at the end of the first page of the calendar, and sponsors that
// are running share the page loads.
export default function AdminSponsors({ sponsors, reload }: AdminSponsorsProps) {
  const changer = useChanger(reload);
  const [adding, setAdding] = useState(false);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [deletingId, setDeletingId] = useState<number | null>(null);

  return (
    <div>
      <p className="mb-4 text-sm text-text-muted">
        Paid cards, shown once at the end of the first page of the calendar with a &quot;Sponsored&quot; label. Counts only say how often a card was seen and
        clicked: nothing about the visitor is kept.
      </p>
      {adding ? (
        <div className="mb-4">
          <SponsorFormFields
            initial={EMPTY}
            currentImageUrl=""
            submitLabel="Add sponsor"
            idPrefix="new-sponsor"
            changer={changer}
            onSave={(sponsor) => changer.run(() => changeAdminSponsor('POST', undefined, sponsor)).then((ok) => (ok && setAdding(false), ok))}
            onCancel={() => setAdding(false)}
          />
        </div>
      ) : (
        <button type="button" onClick={() => setAdding(true)} className={`${smallButtonClass} mb-4`}>
          New sponsor
        </button>
      )}
      <ChangeStatus changer={changer} />

      {sponsors.length === 0 ? (
        <EmptyState title="No sponsors">Add the first one above.</EmptyState>
      ) : (
        <List>
          {sponsors.map((sponsor) => (
            <Row key={sponsor.id}>
              <div className="flex flex-wrap items-center justify-between gap-2">
                <span>
                  <span className="text-sm font-medium text-text">{sponsor.name}</span>{' '}
                  <Meta>
                    {sponsor.active ? 'showing now' : sponsor.enabled ? 'not showing today' : 'off'} · {schedule(sponsor)}
                  </Meta>
                  <br />
                  <Meta>
                    {count(sponsor.impressions, 'view')}, {count(sponsor.clicks, 'click')} · last 30 days: {count(sponsor.impressions30, 'view')},{' '}
                    {count(sponsor.clicks30, 'click')}
                  </Meta>
                </span>
                <span className="flex gap-2">
                  <button type="button" onClick={() => setEditingId(editingId === sponsor.id ? null : sponsor.id)} className={smallButtonClass}>
                    Edit
                  </button>
                  <button type="button" onClick={() => setDeletingId(deletingId === sponsor.id ? null : sponsor.id)} className={dangerButtonClass}>
                    Delete
                  </button>
                </span>
              </div>
              {editingId === sponsor.id && (
                <SponsorFormFields
                  initial={{ name: sponsor.name, text: sponsor.text, link: sponsor.link, startsOn: sponsor.startsOn, endsOn: sponsor.endsOn, enabled: sponsor.enabled }}
                  currentImageUrl={sponsor.imageUrl}
                  submitLabel="Save"
                  idPrefix={`sponsor-${sponsor.id}`}
                  changer={changer}
                  onSave={(next) => changer.run(() => changeAdminSponsor('PATCH', sponsor.id, next)).then((ok) => (ok && setEditingId(null), ok))}
                  onCancel={() => setEditingId(null)}
                />
              )}
              {deletingId === sponsor.id && (
                <div className="flex flex-wrap items-center gap-2 rounded-control border border-border bg-bg p-3">
                  <p className="flex-1 text-sm text-text-muted">
                    Delete <span className="font-medium text-text">{sponsor.name}</span> and its counts?
                  </p>
                  <button
                    type="button"
                    disabled={changer.busy}
                    onClick={() => void changer.run(() => changeAdminSponsor('DELETE', sponsor.id)).then((ok) => ok && setDeletingId(null))}
                    className={dangerButtonClass}
                  >
                    Delete sponsor
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
