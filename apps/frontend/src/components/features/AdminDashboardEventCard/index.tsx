import { useState, useEffect, useRef, useCallback } from 'react';
import type { AdminEvent, Age, FilterOptions, Venue } from '../../../types/event';
import { adminUpdateEvent, fetchAdminAges, fetchFilterOptions, fetchVenues, type AdminUpdateEventData } from '../../../services/api';
import { createSaveQueue } from '../../../utils/saveQueue';
import EditableCurrencyField from '../../ui/EditableCurrencyField';
import { getTagColor } from '../../ui/TagColor';
import { displayTag } from '../../../utils/tag';

interface EditableFieldProps {
  value: string;
  // saves the new value. The promise, if there is one, ends when the save has.
  onChange: (value: string) => void | Promise<void>;
  className?: string;
  multiline?: boolean;
  placeholder?: string;
}

function EditableField({ value, onChange, className = '', multiline = false, placeholder }: EditableFieldProps) {
  const [localValue, setLocalValue] = useState(value);
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  // Counts the edits, so a save that ends can tell if the visitor has typed something newer since.
  const editRef = useRef(0);
  // From the first keystroke until its save has ended and nothing newer was typed. Then the field is
  // the visitor's, and the value from the server must not replace what they are typing.
  const [editing, setEditing] = useState(false);

  const [prevValue, setPrevValue] = useState(value);
  if (value !== prevValue) {
    setPrevValue(value);
    if (!editing) setLocalValue(value);
  }

  const handleChange = (newValue: string) => {
    setLocalValue(newValue);
    setEditing(true);
    const edit = ++editRef.current;

    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
    }

    timeoutRef.current = setTimeout(async () => {
      if (newValue !== value) {
        await onChange(newValue);
      }
      // The refetch that follows a save comes after this, and it is taken then: what the server
      // made of the value, a link tidied up, replaces the text once the visitor has stopped.
      if (edit === editRef.current) setEditing(false);
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
        className={`bg-transparent border border-transparent hover:border-border focus:border-accent focus:outline-none rounded px-1 -mx-1 resize-none ${className}`}
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
      className={`bg-transparent border border-transparent hover:border-border focus:border-accent focus:outline-none rounded px-1 -mx-1 ${className}`}
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
      className={`bg-transparent border border-transparent hover:border-border focus:border-accent focus:outline-none rounded px-1 -mx-1 cursor-pointer ${className}`}
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
  // saves that are waiting or running. "Saving..." shows while there is any.
  const [pendingSaves, setPendingSaves] = useState(0);
  const saving = pendingSaves > 0;
  const [error, setError] = useState<string | null>(null);
  const [venueNames, setVenueNames] = useState<string[]>([]);
  const [regionNames, setRegionNames] = useState<string[]>([]);
  const [ageNames, setAgeNames] = useState<string[]>([]);
  const [allTagNames, setAllTagNames] = useState<string[]>([]);
  const [currentTags, setCurrentTags] = useState<string[]>(event.tags);
  // the tags as the latest toggle left them, which a quick second toggle must start from
  const tagsRef = useRef<string[]>(event.tags);

  // The latest onUpdate, for the queue, which is made once.
  const onUpdateRef = useRef(onUpdate);
  useEffect(() => {
    onUpdateRef.current = onUpdate;
  }, [onUpdate]);
  // whether a save worked since the dashboard was last refreshed
  const savedRef = useRef(false);

  // how many times the queue has emptied
  const [idleCount, setIdleCount] = useState(0);

  // One queue for the card: its saves go to the server one at a time, in the order they were made.
  const [saveQueue] = useState(() =>
    createSaveQueue({
      onPendingChange: setPendingSaves,
      onIdle: () => setIdleCount((count) => count + 1),
    }),
  );

  // The dashboard is refreshed once when the saves are all done, not after each of them.
  useEffect(() => {
    if (idleCount === 0 || !savedRef.current) return;
    savedRef.current = false;
    onUpdateRef.current?.();
  }, [idleCount]);

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

  // Queues one change of the event. Says whether it was saved; a failure is shown on the card and
  // does not stop the changes queued behind it.
  const save = useCallback(
    (change: Omit<AdminUpdateEventData, 'guid'>): Promise<boolean> =>
      saveQueue.run(async () => {
        setError(null);
        try {
          await adminUpdateEvent({ guid: event.guid, ...change });
          savedRef.current = true;
          return true;
        } catch (err) {
          setError(err instanceof Error ? err.message : 'Failed to save');
          return false;
        }
      }),
    [saveQueue, event.guid],
  );

  const handleFieldUpdate = useCallback(
    async (field: 'title' | 'link' | 'coverCost' | 'flyerUrl' | 'venue' | 'region' | 'age' | 'date' | 'startTime' | 'doorsTime', value: string) => {
      await save({ [field]: value });
    },
    [save],
  );

  const handleTagToggle = useCallback(
    async (tagName: string) => {
      const before = tagsRef.current;
      const after = before.includes(tagName) ? before.filter((t) => t !== tagName) : [...before, tagName];
      tagsRef.current = after;
      setCurrentTags(after);

      if (!(await save({ tags: after })) && tagsRef.current === after) {
        // put it back, unless a later toggle has already built on this one
        tagsRef.current = before;
        setCurrentTags(before);
      }
    },
    [save],
  );

  return (
    <div className="space-y-3 rounded-control border border-border bg-surface p-4">
      {error && (
        <div className="text-xs text-danger bg-danger-muted px-2 py-1 rounded">
          {error}
        </div>
      )}

      {event.cancelledAt && (
        <div>
          <span className={`${getTagColor('cancelled')} inline-flex h-[22px] items-center rounded-[4px] px-2 text-xs font-medium`}>
            {displayTag('cancelled')}
          </span>
        </div>
      )}

      <div className="flex items-center justify-between">
        <div className="flex gap-4">
          <div>
            <label className="text-xs text-text-faint">Date</label>
            <input
              type="date"
              value={event.date}
              onChange={(e) => handleFieldUpdate('date', e.target.value)}
              className="block bg-transparent border border-transparent hover:border-border focus:border-accent focus:outline-none rounded px-1 -mx-1 text-sm font-medium text-text-muted cursor-pointer"
            />
          </div>
          <div>
            <label className="text-xs text-text-faint">Venue</label>
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
        {saving && <span className="text-xs text-text-faint">Saving...</span>}
      </div>

      <div>
        <EditableField
          value={event.title}
          onChange={(value) => handleFieldUpdate('title', value)}
          className="text-lg font-semibold text-accent w-full"
          placeholder="Event title"
        />
      </div>

      <div className="text-sm text-text-faint">
        <span className="text-text-muted">Link: </span>
        <EditableField
          value={event.link}
          onChange={(value) => handleFieldUpdate('link', value)}
          className="w-full break-all"
          placeholder="Event URL"
        />
      </div>

      <div className="text-sm text-text-faint">
        <span className="text-text-muted">Flyer: </span>
        <EditableField
          value={event.flyerUrl ?? ''}
          onChange={(value) => handleFieldUpdate('flyerUrl', value)}
          className="w-full break-all"
          placeholder="Flyer image URL (optional)"
        />
      </div>

      <div className="flex flex-wrap gap-4 text-sm text-text-muted">
        <div>
          <label className="text-xs text-text-faint">Region</label>
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
          <label className="text-xs text-text-faint">Age</label>
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
          <label className="text-xs text-text-faint">Cost</label>
          <EditableCurrencyField
            value={event.coverCost}
            onChange={(value) => handleFieldUpdate('coverCost', value)}
            className="w-24"
          />
        </div>
        <div className="basis-full sm:basis-auto" />
        <div>
          <label className="text-xs text-text-faint">Start</label>
          <input
            type="time"
            value={event.startTime}
            onChange={(e) => handleFieldUpdate('startTime', e.target.value)}
            className="block bg-transparent border border-transparent hover:border-border focus:border-accent focus:outline-none rounded px-1 -mx-1 text-sm text-text-muted cursor-pointer"
          />
        </div>
        <div>
          <label className="text-xs text-text-faint">Doors</label>
          <input
            type="time"
            value={event.doorsTime ?? ''}
            onChange={(e) => handleFieldUpdate('doorsTime', e.target.value)}
            className="block bg-transparent border border-transparent hover:border-border focus:border-accent focus:outline-none rounded px-1 -mx-1 text-sm text-text-muted cursor-pointer"
          />
        </div>
      </div>

      {event.email && (
        <p className="text-sm text-text-faint">
          Submission from <span className="font-medium text-text-muted">{event.email}</span>
        </p>
      )}

      <div>
        <div className="flex items-center justify-between mb-1">
          <label className="text-xs text-text-faint">Tags</label>
          <span className="text-sm font-medium text-text-muted">Score: {event.score}</span>
        </div>
        <div className="flex flex-wrap gap-2">
          {allTagNames.length > 0 ? (
            allTagNames.map((tag) => (
              <button
                key={tag}
                type="button"
                onClick={() => handleTagToggle(tag)}
                className={`min-h-8 cursor-pointer rounded-control border px-2.5 text-xs font-medium ${
                  currentTags.includes(tag)
                    ? 'border-accent bg-accent/15 text-accent'
                    : 'border-border text-text-muted hover:border-text-faint'
                }`}
              >
                {displayTag(tag)}
              </button>
            ))
          ) : (
            currentTags.map((tag) => (
              <span
                key={tag}
                className={`flex h-[22px] items-center rounded-[4px] px-2 text-xs font-medium ${getTagColor(tag)}`}
              >
                {displayTag(tag)}
              </span>
            ))
          )}
        </div>
      </div>

      <div className="flex gap-2.5">
        <button
          type="button"
          onClick={() => onVote(event.guid, mode === 'approval' ? 'Approve' : 'Keep')}
          className="font-label h-11 flex-1 cursor-pointer rounded-control bg-success text-sm font-semibold text-bg hover:opacity-90"
        >
          {mode === 'approval' ? 'Approve' : 'Keep'}
        </button>
        <button
          type="button"
          onClick={() => onVote(event.guid, mode === 'approval' ? 'Deny' : 'Remove')}
          className="font-label h-11 flex-1 cursor-pointer rounded-control border border-danger text-sm font-semibold text-danger hover:bg-danger-muted"
        >
          {mode === 'approval' ? 'Deny' : 'Remove'}
        </button>
      </div>
    </div>
  );
}
