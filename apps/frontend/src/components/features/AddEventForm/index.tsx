import { useState, useEffect, type FormEvent } from 'react';
import { fetchFilterOptions, fetchSubmissionVenues } from '../../../services/api';
import type { FilterOptions, Venue } from '../../../types/event';
import DateInput from '../../ui/DateInput';
import EditableCurrencyField from '../../ui/EditableCurrencyField';

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
      } catch {
        const errorMsg = 'Failed to load options';
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
      {error && (
        <div className="mb-6 p-4 bg-error-muted border border-error/30 rounded-lg">
          <p className="text-error">{error}</p>
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-6">
        <div>
          <label htmlFor="title" className="block text-sm font-medium text-text mb-1">
            {isAdmin ? 'Title *' : 'Event Title *'}
          </label>
          <input
            type="text"
            id="title"
            required
            value={formData.title}
            onChange={(e) => setFormData({ ...formData, title: e.target.value })}
            className="w-full px-3 py-2 bg-surface border border-border rounded-md text-text focus:outline-none focus:ring-2 focus:ring-primary"
          />
        </div>

        <div>
          <label htmlFor="link" className="block text-sm font-medium text-text mb-1">
            {isAdmin ? 'URL *' : 'Event URL *'}
          </label>
          <input
            type="text"
            id="link"
            required
            placeholder="example.com/event"
            value={formData.link}
            onChange={(e) => setFormData({ ...formData, link: e.target.value })}
            className="w-full px-3 py-2 bg-surface border border-border rounded-md text-text placeholder-text-dim focus:outline-none focus:ring-2 focus:ring-primary"
          />
        </div>

        <div>
          <label htmlFor="venue" className="block text-sm font-medium text-text mb-1">
            Venue *
          </label>
          <select
            id="venue"
            required
            value={formData.venue}
            onChange={(e) => setFormData({ ...formData, venue: e.target.value })}
            disabled={venuesLoading}
            className="w-full px-3 py-2 bg-surface border border-border rounded-md text-text focus:outline-none focus:ring-2 focus:ring-primary"
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
            <label htmlFor="region" className="block text-sm font-medium text-text mb-1">
              Region *
            </label>
            <select
              id="region"
              required
              value={formData.region}
              onChange={(e) => setFormData({ ...formData, region: e.target.value })}
              className="w-full px-3 py-2 bg-surface border border-border rounded-md text-text focus:outline-none focus:ring-2 focus:ring-primary"
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
            <label htmlFor="date" className="block text-sm font-medium text-text mb-1">
              Date *
            </label>
            <DateInput
              id="date"
              required
              value={formData.date}
              onChange={(e) => setFormData({ ...formData, date: e.target.value })}
              className="w-full py-2 rounded-md focus:ring-2 focus:ring-primary"
            />
          </div>

          <div>
            <label htmlFor="age" className="block text-sm font-medium text-text mb-1">
              {isAdmin ? 'Age Limit *' : 'Age Restriction'}
            </label>
            <select
              id="age"
              required={isAdmin}
              value={formData.age}
              onChange={(e) => setFormData({ ...formData, age: e.target.value })}
              className="w-full px-3 py-2 bg-surface border border-border rounded-md text-text focus:outline-none focus:ring-2 focus:ring-primary"
            >
              {ages.map((age) => (
                <option key={age} value={age}>
                  {age}
                </option>
              ))}
            </select>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label htmlFor="doorsTime" className="block text-sm font-medium text-text mb-1">
              {isAdmin ? 'Doors' : 'Doors Open'}
            </label>
            <DateInput
              type="time"
              id="doorsTime"
              value={formData.doorsTime}
              onChange={(e) => setFormData({ ...formData, doorsTime: e.target.value })}
              className="w-full py-2 rounded-md focus:ring-2 focus:ring-primary"
            />
          </div>

          <div>
            <label htmlFor="startTime" className="block text-sm font-medium text-text mb-1">
              {isAdmin ? 'Start *' : 'Start Time *'}
            </label>
            <DateInput
              type="time"
              id="startTime"
              required
              value={formData.startTime}
              onChange={(e) => setFormData({ ...formData, startTime: e.target.value })}
              className="w-full py-2 rounded-md focus:ring-2 focus:ring-primary"
            />
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium text-text mb-1">
            {isAdmin ? 'Cost' : 'Cover / Cost'}
          </label>
          <EditableCurrencyField
            value={formData.coverCost}
            onChange={(value) => setFormData({ ...formData, coverCost: value })}
            variant="form"
          />
        </div>

        <div className="p-4 bg-surface border border-border rounded-lg">
          <label className="block text-lg font-semibold text-text mb-3">Tags</label>
          <div className="grid grid-cols-2 gap-2">
            {tags.map((tag) => (
              <button
                key={tag}
                type="button"
                onClick={() => toggleTag(tag)}
                className={`px-3 py-2 text-sm rounded border transition-colors text-left ${
                  formData.tags.includes(tag)
                    ? 'bg-primary text-white border-primary'
                    : 'bg-surface-elevated text-text-muted border-border hover:border-primary'
                }`}
              >
                {tag}
              </button>
            ))}
          </div>
        </div>

        {!isAdmin && (
          <div>
            <label htmlFor="email" className="block text-sm font-medium text-text mb-1">
              Your Email *
            </label>
            <input
              type="email"
              id="email"
              required
              placeholder="For verification purposes"
              value={formData.email}
              onChange={(e) => setFormData({ ...formData, email: e.target.value })}
              className="w-full px-3 py-2 bg-surface border border-border rounded-md text-text placeholder-text-dim focus:outline-none focus:ring-2 focus:ring-primary"
            />
            <p className="mt-1 text-xs text-text-dim">
              We'll send you a verification email. Your email won't be displayed publicly.
            </p>
          </div>
        )}

        <button
          type="submit"
          disabled={submitting}
          className="w-full px-4 py-3 bg-primary text-white font-medium rounded-md hover:bg-primary-hover disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {submitting ? 'Submitting...' : (isAdmin ? 'Submit' : 'Submit Event')}
        </button>
      </form>
    </>
  );
}
