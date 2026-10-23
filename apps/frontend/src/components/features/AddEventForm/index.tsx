import { useState, useEffect, type FormEvent } from 'react';
import { fetchFilterOptions, fetchSubmissionVenues } from '../../../services/api';
import type { FilterOptions, Venue } from '../../../types/event';
import DateInput from '../../ui/DateInput';
import EditableCurrencyField from '../../ui/EditableCurrencyField';
import Chip from '../../ui/Chip';
import InlineError from '../../ui/InlineError';
import { toLoadError } from '../../../utils/loadError';
import { fieldClass, primaryButtonClass } from '../../ui/Field';
import { displayAge } from '../../../utils/age';
import { displayTag } from '../../../utils/tag';

const defaultTags = ['music', 'comedy', 'community', 'literature', 'diy', 'art', 'sports', 'film', 'free', 'recurring'];
const defaultAgeOptions = ['all-ages', '16+', '18+', '21+', 'unknown'];

export interface AddEventFormData {
  title: string;
  link: string;
  venue: string;
  region: string;
  age: string;
  date: string;
  startTime: string;
  doorsTime: string;
  coverCost: string;
  tags: string[];
  email?: string;
}

interface AddEventFormProps {
  isAdmin?: boolean;
  onSubmit: (data: AddEventFormData) => Promise<void>;
  onLoadError?: (error: string) => void;
}

export default function AddEventForm({ isAdmin = false, onSubmit, onLoadError }: AddEventFormProps) {
  const [options, setOptions] = useState<FilterOptions | null>(null);
  const [venues, setVenues] = useState<Venue[]>([]);
  const [venuesLoading, setVenuesLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [formData, setFormData] = useState<AddEventFormData>({
    title: '',
    link: '',
    venue: '',
    region: '',
    age: 'all-ages',
    date: '',
    startTime: '',
    doorsTime: '',
    coverCost: '',
    tags: [],
    email: '',
  });

  useEffect(() => {
    async function loadOptions() {
      try {
        if (isAdmin) {
          const [filterData, venueData] = await Promise.all([
            fetchFilterOptions(),
            fetchSubmissionVenues(),
          ]);
          setOptions(filterData);
          setVenues(venueData);
          if (venueData.length > 0) {
            setFormData((prev) => ({ ...prev, venue: venueData[0].name }));
          }
          if (filterData.regions.length > 0) {
            setFormData((prev) => ({ ...prev, region: filterData.regions[0].name }));
          }
          if (filterData.ages.length > 0) {
            setFormData((prev) => ({ ...prev, age: filterData.ages[0].name }));
          }
        } else {
          const venueData = await fetchSubmissionVenues();
          setVenues(venueData);
          if (venueData.length > 0) {
            setFormData((prev) => ({ ...prev, venue: venueData[0].name }));
          }
        }
      } catch (err) {
        const { title, message } = toLoadError(err);
        const errorMsg = `${title}. ${message}`;
        setError(errorMsg);
        onLoadError?.(errorMsg);
      } finally {
        setVenuesLoading(false);
      }
    }

    loadOptions();
  }, [isAdmin, onLoadError]);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    setError(null);

    try {
      await onSubmit(formData);
      // Reset form on success
      setFormData({
        title: '',
        link: '',
        venue: venues[0]?.name ?? '',
        region: isAdmin ? (options?.regions[0]?.name ?? '') : '',
        age: isAdmin ? (options?.ages[0]?.name ?? 'all-ages') : 'all-ages',
        date: '',
        startTime: '',
        doorsTime: '',
        coverCost: '',
        tags: [],
        email: '',
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to submit event');
    } finally {
      setSubmitting(false);
    }
  };

  const toggleTag = (tag: string) => {
    setFormData((prev) => ({
      ...prev,
      tags: prev.tags.includes(tag)
        ? prev.tags.filter((t) => t !== tag)
        : [...prev.tags, tag],
    }));
  };

  const tags = isAdmin && options?.tags ? options.tags.map((t) => t.name) : defaultTags;
  const ages = isAdmin && options?.ages ? options.ages.map((a) => a.name) : defaultAgeOptions;

  return (
    <>
      {error && <InlineError message={error} className="mb-5" />}

      <form onSubmit={handleSubmit} className="space-y-5">
        <div>
          <label htmlFor="title" className="mb-1.5 block text-[13px] font-medium text-text-muted">
            Event title *
          </label>
          <input
            type="text"
            id="title"
            required
            value={formData.title}
            onChange={(e) => setFormData({ ...formData, title: e.target.value })}
            className={fieldClass}
          />
        </div>

        <div>
          <label htmlFor="link" className="mb-1.5 block text-[13px] font-medium text-text-muted">
            Event URL *
          </label>
          <input
            type="text"
            id="link"
            required
            placeholder="example.com/event"
            value={formData.link}
            onChange={(e) => setFormData({ ...formData, link: e.target.value })}
            className={fieldClass}
          />
        </div>

        <div>
          <label htmlFor="venue" className="mb-1.5 block text-[13px] font-medium text-text-muted">
            Venue *
          </label>
          <select
            id="venue"
            required
            value={formData.venue}
            onChange={(e) => setFormData({ ...formData, venue: e.target.value })}
            disabled={venuesLoading}
            className={fieldClass}
          >
            {venuesLoading ? (
              <option>Loading venues...</option>
            ) : (
              venues.map((venue) => (
                <option key={venue.slug} value={venue.name}>
                  {venue.name}
                </option>
              ))
            )}
          </select>
        </div>

        {isAdmin && (
          <div>
            <label htmlFor="region" className="mb-1.5 block text-[13px] font-medium text-text-muted">
              Region *
            </label>
            <select
              id="region"
              required
              value={formData.region}
              onChange={(e) => setFormData({ ...formData, region: e.target.value })}
              className={fieldClass}
            >
              {options?.regions.map((region) => (
                <option key={region.id} value={region.name}>
                  {region.name}
                </option>
              ))}
            </select>
          </div>
        )}

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label htmlFor="date" className="mb-1.5 block text-[13px] font-medium text-text-muted">
              Date *
            </label>
            <DateInput
              id="date"
              required
              value={formData.date}
              onChange={(e) => setFormData({ ...formData, date: e.target.value })}
              className={`${fieldClass} px-3`}
            />
          </div>

          <div>
            <label htmlFor="age" className="mb-1.5 block text-[13px] font-medium text-text-muted">
              Age
            </label>
            <select
              id="age"
              required={isAdmin}
              value={formData.age}
              onChange={(e) => setFormData({ ...formData, age: e.target.value })}
              className={fieldClass}
            >
              {ages.map((age) => (
                <option key={age} value={age}>
                  {displayAge(age)}
                </option>
              ))}
            </select>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label htmlFor="doorsTime" className="mb-1.5 block text-[13px] font-medium text-text-muted">
              Doors open
            </label>
            <DateInput
              type="time"
              id="doorsTime"
              value={formData.doorsTime}
              onChange={(e) => setFormData({ ...formData, doorsTime: e.target.value })}
              className={`${fieldClass} px-3`}
            />
          </div>

          <div>
            <label htmlFor="startTime" className="mb-1.5 block text-[13px] font-medium text-text-muted">
              Start time *
            </label>
            <DateInput
              type="time"
              id="startTime"
              required
              value={formData.startTime}
              onChange={(e) => setFormData({ ...formData, startTime: e.target.value })}
              className={`${fieldClass} px-3`}
            />
          </div>
        </div>

        <div>
          <label className="mb-1.5 block text-[13px] font-medium text-text-muted">
            Cover / cost
          </label>
          <EditableCurrencyField
            value={formData.coverCost}
            onChange={(value) => setFormData({ ...formData, coverCost: value })}
            variant="form"
          />
        </div>

        <div>
          <span className="mb-2 block text-[13px] font-medium text-text-muted">Tags</span>
          <div className="flex flex-wrap gap-2">
            {tags.map((tag) => (
              <Chip key={tag} label={displayTag(tag)} selected={formData.tags.includes(tag)} onToggle={() => toggleTag(tag)} />
            ))}
          </div>
        </div>

        {!isAdmin && (
          <div>
            <label htmlFor="email" className="mb-1.5 block text-[13px] font-medium text-text-muted">
              Your email *
            </label>
            <input
              type="email"
              id="email"
              required
              placeholder="For verification purposes"
              value={formData.email}
              onChange={(e) => setFormData({ ...formData, email: e.target.value })}
              className={fieldClass}
            />
            <p className="mt-1.5 text-xs text-text-faint">
              We&apos;ll email a verification link. Your email isn&apos;t shown publicly.
            </p>
          </div>
        )}

        <button
          type="submit"
          disabled={submitting}
          className={primaryButtonClass}
        >
          {submitting ? 'Submitting...' : isAdmin ? 'Submit' : 'Submit event'}
        </button>
      </form>
    </>
  );
}
