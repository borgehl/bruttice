import { useState } from 'react';
import type { Venue } from '../../../types/event';
import { visibleVenues } from '../../../utils/venueFilter';

interface VenueListProps {
  // venues already narrowed by region
  venues: Venue[];
  selected: string[];
  onToggle: (name: string) => void;
}

// A searchable checkbox list. The search only narrows what is shown, it never changes the selection.
export default function VenueList({ venues, selected, onToggle }: VenueListProps) {
  const [query, setQuery] = useState('');
  const shown = visibleVenues(venues, [], query);

  return (
    <div>
      <label className="relative mb-2 flex">
        <span className="sr-only">Search venues</span>
        <svg
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-text-faint"
          aria-hidden="true"
        >
          <circle cx="11" cy="11" r="7" />
          <path d="M20 20l-4-4" />
        </svg>
        <input
          type="search"
          value={query}
          onChange={(event) => setQuery(event.target.value)}
          placeholder="Search venues"
          className="h-10 w-full rounded-control border border-border bg-raised pl-9 pr-3 text-sm text-text placeholder-text-faint focus:border-accent focus:outline-none"
        />
      </label>

      {shown.length === 0 ? (
        <div className="py-4 text-center text-sm text-text-faint">
          <p>{query.trim() ? `No venues match "${query.trim()}".` : 'No venues to show.'}</p>
          {query && (
            <button type="button" onClick={() => setQuery('')} className="mt-1 min-h-11 cursor-pointer text-accent hover:underline">
              Clear search
            </button>
          )}
        </div>
      ) : (
        <ul className="max-h-64 overflow-y-auto rounded-control border border-border">
          {shown.map((venue) => {
            const checked = selected.includes(venue.name);
            return (
              <li key={venue.slug} className="border-b border-border last:border-b-0">
                <label
                  className={`flex min-h-11 cursor-pointer items-center gap-3 px-3 text-sm ${
                    checked ? 'bg-accent/10 text-text' : 'text-text-muted'
                  }`}
                >
                  <input
                    type="checkbox"
                    checked={checked}
                    onChange={() => onToggle(venue.name)}
                    className="h-[18px] w-[18px] shrink-0 cursor-pointer accent-[var(--color-accent)]"
                  />
                  <span className="min-w-0 flex-1 truncate">{venue.name}</span>
                </label>
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}
