import { useState, useEffect, useRef, useCallback } from 'react';
import type { AdminEvent, Age, FilterOptions, Venue } from '../../../types/event';
import { adminUpdateEvent, fetchAdminAges, fetchFilterOptions, fetchVenues } from '../../../services/api';
import EditableCurrencyField from '../../ui/EditableCurrencyField';

interface EditableFieldProps {
  value: string;
  onChange: (value: string) => void;
  className?: string;
  multiline?: boolean;
  placeholder?: string;
}

function EditableField({ value, onChange, className = '', multiline = false, placeholder }: EditableFieldProps) {
  const [localValue, setLocalValue] = useState(value);
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    setLocalValue(value);
  }, [value]);

  const handleChange = (newValue: string) => {
    setLocalValue(newValue);

    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
    }

    timeoutRef.current = setTimeout(() => {
      if (newValue !== value) {
        onChange(newValue);
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

  if (multiline) {
    return (
      <textarea
        value={localValue}
        onChange={(e) => handleChange(e.target.value)}
        placeholder={placeholder}
        className={`bg-transparent border border-transparent hover:border-border focus:border-primary focus:outline-none rounded px-1 -mx-1 resize-none ${className}`}
        rows={2}
      />
    );
  }

  return (
    <input
      type="text"
      value={localValue}
      onChange={(e) => handleChange(e.target.value)}
      placeholder={placeholder}
      className={`bg-transparent border border-transparent hover:border-border focus:border-primary focus:outline-none rounded px-1 -mx-1 ${className}`}
    />
  );
}

interface EditableSelectProps {
  value: string;
  options: string[];
  onChange: (value: string) => void;
  className?: string;
}

function EditableSelect({ value, options, onChange, className = '' }: EditableSelectProps) {
  const handleChange = (newValue: string) => {
    if (newValue !== value) {
      onChange(newValue);
    }
  };

  return (
    <select
      value={value}
      onChange={(e) => handleChange(e.target.value)}
      className={`bg-transparent border border-transparent hover:border-border focus:border-primary focus:outline-none rounded px-1 -mx-1 cursor-pointer ${className}`}
    >
      {options.map((opt) => (
        <option key={opt} value={opt}>
          {opt}
        </option>
      ))}
    </select>
  );
}

let cachedOptions: { filterOptions: FilterOptions; venues: Venue[]; ages: Age[] } | null = null;

export type CardMode = 'approval' | 'upcoming';

export function AdminEventCard({
  event,
  mode = 'approval',
  onVote,
  onUpdate,
}: {
  event: AdminEvent;
  mode?: CardMode;
  onVote: (guid: string, vote: string) => void;
  onUpdate?: () => void;
}) {
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [venueNames, setVenueNames] = useState<string[]>([]);
  const [regionNames, setRegionNames] = useState<string[]>([]);
  const [ageNames, setAgeNames] = useState<string[]>([]);
  const [allTagNames, setAllTagNames] = useState<string[]>([]);
  const [currentTags, setCurrentTags] = useState<string[]>(event.tags);

  useEffect(() => {
    async function loadOptions() {
      if (cachedOptions) {
        setVenueNames(cachedOptions.venues.map((v) => v.name));
        setRegionNames(cachedOptions.filterOptions.regions.map((r) => r.name));
        setAgeNames(cachedOptions.ages.map((a) => a.name));
        setAllTagNames(cachedOptions.filterOptions.tags.map((t) => t.name).sort());
        return;
      }
      try {
        const [filterData, venueData, ageData] = await Promise.all([
          fetchFilterOptions(),
          fetchVenues(),
          fetchAdminAges(),
        ]);
        cachedOptions = { filterOptions: filterData, venues: venueData, ages: ageData };
        setVenueNames(venueData.map((v) => v.name));
        setRegionNames(filterData.regions.map((r) => r.name));
        setAgeNames(ageData.map((a) => a.name));
        setAllTagNames(filterData.tags.map((t) => t.name).sort());
      } catch {
        // Options will remain empty; fields stay as text
      }
    }
    loadOptions();
  }, []);

  const handleFieldUpdate = useCallback(async (field: 'title' | 'link' | 'coverCost' | 'venue' | 'region' | 'age' | 'date' | 'startTime' | 'doorsTime', value: string) => {
    setSaving(true);
    setError(null);

    try {
      await adminUpdateEvent({
        guid: event.guid,
        [field]: value,
      });
      onUpdate?.();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save');
    } finally {
      setSaving(false);
    }
  }, [event.guid, onUpdate]);

  const handleTagToggle = useCallback(async (tagName: string) => {
    const newTags = currentTags.includes(tagName)
      ? currentTags.filter((t) => t !== tagName)
      : [...currentTags, tagName];
    setCurrentTags(newTags);
    setSaving(true);
    setError(null);
    try {
      await adminUpdateEvent({
        guid: event.guid,
        tags: newTags,
      });
      onUpdate?.();
    } catch (err) {
      setCurrentTags(currentTags);
      setError(err instanceof Error ? err.message : 'Failed to save');
    } finally {
      setSaving(false);
    }
  }, [event.guid, currentTags, onUpdate]);

  return (
    <div className="p-4 bg-surface-elevated border border-border rounded-lg space-y-3">
      {error && (
        <div className="text-xs text-error bg-error-muted px-2 py-1 rounded">
          {error}
        </div>
      )}

      <div className="flex items-center justify-between">
        <div className="flex gap-4">
          <div>
            <label className="text-xs text-text-dim">Date</label>
            <input
              type="date"
              value={event.date}
              onChange={(e) => handleFieldUpdate('date', e.target.value)}
              className="block bg-transparent border border-transparent hover:border-border focus:border-primary focus:outline-none rounded px-1 -mx-1 text-sm font-medium text-text-muted cursor-pointer"
            />
          </div>
          <div>
            <label className="text-xs text-text-dim">Venue</label>
            {venueNames.length > 0 ? (
              <EditableSelect
                value={event.venue}
                options={venueNames}
                onChange={(value) => handleFieldUpdate('venue', value)}
                className="text-sm font-medium text-text-muted block"
              />
            ) : (
              <p className="text-sm font-medium text-text-muted">{event.venue}</p>
            )}
          </div>
        </div>
        {saving && <span className="text-xs text-text-dim">Saving...</span>}
      </div>

      <div>
        <EditableField
          value={event.title}
          onChange={(value) => handleFieldUpdate('title', value)}
          className="text-lg font-semibold text-primary w-full"
          placeholder="Event title"
        />
      </div>

      <div className="text-sm text-text-dim">
        <span className="text-text-muted">Link: </span>
        <EditableField
          value={event.link}
          onChange={(value) => handleFieldUpdate('link', value)}
          className="w-full break-all"
          placeholder="Event URL"
        />
      </div>

      <div className="flex flex-wrap gap-4 text-sm text-text-muted">
        <div>
          <label className="text-xs text-text-dim">Region</label>
          {regionNames.length > 0 ? (
            <EditableSelect
              value={event.region}
              options={regionNames}
              onChange={(value) => handleFieldUpdate('region', value)}
              className="block"
            />
          ) : (
            <span>{event.region}</span>
          )}
        </div>
        <div>
          <label className="text-xs text-text-dim">Age</label>
          {ageNames.length > 0 ? (
            <EditableSelect
              value={event.age}
              options={ageNames}
              onChange={(value) => handleFieldUpdate('age', value)}
              className="block"
            />
          ) : (
            <span>{event.age}</span>
          )}
        </div>
        <div>
          <label className="text-xs text-text-dim">Cost</label>
          <EditableCurrencyField
            value={event.coverCost}
            onChange={(value) => handleFieldUpdate('coverCost', value)}
            className="w-24"
          />
        </div>
        <div className="basis-full sm:basis-auto" />
        <div>
          <label className="text-xs text-text-dim">Start</label>
          <input
            type="time"
            value={event.startTime}
            onChange={(e) => handleFieldUpdate('startTime', e.target.value)}
            className="block bg-transparent border border-transparent hover:border-border focus:border-primary focus:outline-none rounded px-1 -mx-1 text-sm text-text-muted cursor-pointer"
          />
        </div>
        <div>
          <label className="text-xs text-text-dim">Doors</label>
          <input
            type="time"
            value={event.doorsTime ?? ''}
            onChange={(e) => handleFieldUpdate('doorsTime', e.target.value)}
            className="block bg-transparent border border-transparent hover:border-border focus:border-primary focus:outline-none rounded px-1 -mx-1 text-sm text-text-muted cursor-pointer"
          />
        </div>
      </div>

      {event.email && (
        <p className="text-sm text-text-dim">
          Submission from <span className="font-medium text-text-muted">{event.email}</span>
        </p>
      )}

      <div>
        <div className="flex items-center justify-between mb-1">
          <label className="text-xs text-text-dim">Tags</label>
          <span className="text-sm font-medium text-text-muted">Score: {event.score}</span>
        </div>
        <div className="flex flex-wrap gap-2">
          {allTagNames.length > 0 ? (
            allTagNames.map((tag) => (
              <button
                key={tag}
                type="button"
                onClick={() => handleTagToggle(tag)}
                className={`px-2 py-0.5 text-xs rounded-full transition-colors cursor-pointer ${
                  currentTags.includes(tag)
                    ? 'bg-primary text-white border border-primary'
                    : 'bg-surface border border-border text-text-dim hover:border-text-muted'
                }`}
              >
                {tag}
              </button>
            ))
          ) : (
            currentTags.map((tag) => (
              <span
                key={tag}
                className="px-2 py-0.5 text-xs bg-surface border border-border rounded-full text-text-muted"
              >
                {tag}
              </span>
            ))
          )}
        </div>
      </div>

      <div className="flex gap-2 ml-auto">
        <button
          onClick={() => onVote(event.guid, mode === 'approval' ? 'Approve' : 'Keep')}
          className="px-3 py-1 text-xs font-medium rounded transition-colors bg-success text-white hover:bg-success/80"
        >
          {mode === 'approval' ? 'Approve' : 'Keep'}
        </button>
        <button
          onClick={() => onVote(event.guid, mode === 'approval' ? 'Deny' : 'Remove')}
          className="px-3 py-1 text-xs font-medium rounded transition-colors bg-error text-white hover:bg-error/80"
        >
          {mode === 'approval' ? 'Deny' : 'Remove'}
        </button>
      </div>
    </div>
  );
}
