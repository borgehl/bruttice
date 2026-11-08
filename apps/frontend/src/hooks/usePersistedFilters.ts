import { useCallback, useEffect, useRef } from 'react';
import { useSearchParams } from 'react-router-dom';
import { filtersFromParams, readFilters, restoreParams, writeFilters } from '../store/filters';
import { regionsKey } from '../utils/regionFilter';

// Keeps the calendar's filter key and regions across visits and restarts.
//
// When the calendar opens with neither in the address, the saved ones are put in (replacing the
// entry, so Back isn't affected). An address that has them, such as a shared link, is left alone.
// Every later change is saved, so clearing the filters clears what is saved.
//
// Returns keepsRegions(filterKey), for when a filter key has loaded: true once for the key that was
// restored. The regions saved with a key are the ones the visitor last had, and may differ from the
// key's own, so they are kept rather than replaced by the key's.
export function usePersistedFilters(): (filterKey: string) => boolean {
  const [searchParams, setSearchParams] = useSearchParams();
  const restoredKey = useRef<string | undefined>(undefined);
  const started = useRef(false);
  const skipSave = useRef(false);

  useEffect(() => {
    if (started.current) return;
    started.current = true;

    const next = restoreParams(searchParams, readFilters());
    if (!next) return;

    restoredKey.current = next.get('filterKey') || undefined;
    // the address changes on the next render, saving before that would overwrite what was just read
    skipSave.current = true;
    setSearchParams(next, { replace: true });
  }, [searchParams, setSearchParams]);

  const { filterKey, regions } = filtersFromParams(searchParams);
  const regionsId = regionsKey(regions);

  useEffect(() => {
    if (skipSave.current) {
      skipSave.current = false;
      return;
    }
    writeFilters({ filterKey, regions });
    // regionsId stands for regions, which is a new array on every render
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [filterKey, regionsId]);

  return useCallback((key: string) => {
    if (restoredKey.current !== key) return false;
    restoredKey.current = undefined;
    return true;
  }, []);
}
