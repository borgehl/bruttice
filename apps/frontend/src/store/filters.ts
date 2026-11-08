import { regionsFromParams, withRegions } from '../utils/regionFilter';

// The calendar's last filters, kept on the device so an installed app or a later visit opens on them.
// A filter key is a shared name for a set of tags, venues and ages, and regions are public names, so
// nothing in here identifies anyone. The start date is left out: a saved date goes stale.
export const FILTERS_KEY = 'calendar-filters';
export const FILTER_KEY_PARAM = 'filterKey';

export interface SavedFilters {
  filterKey?: string;
  // region names, every region when empty
  regions: string[];
}

export const NO_FILTERS: SavedFilters = { regions: [] };

// Anything that isn't what writeFilters saves is no filters.
export function parseFilters(raw: string | null | undefined): SavedFilters {
  if (!raw) return NO_FILTERS;
  try {
    const value: unknown = JSON.parse(raw);
    if (typeof value !== 'object' || value === null) return NO_FILTERS;
    const { filterKey, regions } = value as { filterKey?: unknown; regions?: unknown };
    return {
      filterKey: typeof filterKey === 'string' && filterKey !== '' ? filterKey : undefined,
      regions: Array.isArray(regions)
        ? regions.filter((name): name is string => typeof name === 'string' && name.trim() !== '')
        : [],
    };
  } catch {
    return NO_FILTERS;
  }
}

// Storage can be missing or throw (private windows, blocked site data), so every access is guarded.
export function readFilters(storage: Pick<Storage, 'getItem'> | undefined = safeStorage()): SavedFilters {
  try {
    return parseFilters(storage?.getItem(FILTERS_KEY));
  } catch {
    return NO_FILTERS;
  }
}

export function writeFilters(filters: SavedFilters, storage: Pick<Storage, 'setItem'> | undefined = safeStorage()): void {
  try {
    storage?.setItem(FILTERS_KEY, JSON.stringify(filters));
  } catch {
    // The filters still apply for this visit.
  }
}

function safeStorage(): Storage | undefined {
  try {
    return globalThis.localStorage;
  } catch {
    return undefined;
  }
}

// The filters in the address.
export function filtersFromParams(params: URLSearchParams): SavedFilters {
  return { filterKey: params.get(FILTER_KEY_PARAM) || undefined, regions: regionsFromParams(params) };
}

// The address with the saved filters put in, or null when there is nothing to put in. An address that
// already names a filter key or regions is a choice (a shared link, say) and is never changed.
export function restoreParams(params: URLSearchParams, saved: SavedFilters): URLSearchParams | null {
  const current = filtersFromParams(params);
  if (current.filterKey || current.regions.length > 0) return null;
  if (!saved.filterKey && saved.regions.length === 0) return null;

  const next = withRegions(params, saved.regions);
  if (saved.filterKey) next.set(FILTER_KEY_PARAM, saved.filterKey);
  return next;
}
