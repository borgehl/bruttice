import type { Venue } from '../types/event';

// Venues the Filters page lists: those in any selected region (all of them when none is selected)
// whose name matches the search text, ignoring case.
export function visibleVenues(venues: Venue[], selectedRegions: string[], query = ''): Venue[] {
  const needle = query.trim().toLowerCase();
  return venues.filter(
    (venue) =>
      (selectedRegions.length === 0 || selectedRegions.includes(venue.region)) &&
      (needle === '' || venue.name.toLowerCase().includes(needle)),
  );
}

// Picking or dropping a region must not leave a hidden venue selected, so the saved filter never
// carries one the person can no longer see.
export function dropHiddenVenues(selectedVenues: string[], venues: Venue[], selectedRegions: string[]): string[] {
  const visible = new Set(visibleVenues(venues, selectedRegions).map((venue) => venue.name));
  return selectedVenues.filter((name) => visible.has(name));
}

export function toggle<T>(list: T[], value: T): T[] {
  return list.includes(value) ? list.filter((item) => item !== value) : [...list, value];
}

// "Showing venues in Downtown, Brookside" or the prompt to pick a region.
export function venueHint(selectedRegions: string[]): string {
  return selectedRegions.length > 0
    ? `Showing venues in ${selectedRegions.join(', ')}`
    : 'Pick a region to narrow this list';
}
