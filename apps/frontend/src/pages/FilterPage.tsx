import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { fetchFilterOptions, createFilter } from '../services/api';
import type { FilterOptions } from '../types/event';
import Page from '../components/ui/Page';
import Chip from '../components/ui/Chip';
import LoadingBlock from '../components/ui/LoadingBlock';
import LoadFailed from '../components/ui/LoadFailed';
import { toLoadError, type LoadError } from '../utils/loadError';
import Segmented from '../components/ui/Segmented';
import FilterGroup from '../components/features/FilterGroup';
import VenueList from '../components/features/VenueList';
import { displayAge } from '../utils/age';
import { displayTag } from '../utils/tag';
import { dropHiddenVenues, toggle, venueHint, visibleVenues } from '../utils/venueFilter';

type Apply = 'include' | 'exclude';
type Bound = 'max' | 'min';

const includeOptions: { value: Apply; label: string }[] = [
  { value: 'include', label: 'Include' },
  { value: 'exclude', label: 'Exclude' },
];
const ageOptions: { value: Bound; label: string }[] = [
  { value: 'max', label: 'Max' },
  { value: 'min', label: 'Min' },
];

export default function FilterPage() {
  const navigate = useNavigate();

  const [filterOptions, setFilterOptions] = useState<FilterOptions | null>(null);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState<LoadError | null>(null);
  const [error, setError] = useState<string | null>(null);
  // bumped by Try again to load the options once more
  const [attempt, setAttempt] = useState(0);

  const [selectedTags, setSelectedTags] = useState<string[]>([]);
  const [tagsMode, setTagsMode] = useState<Apply>('include');
  const [selectedAges, setSelectedAges] = useState<number[]>([]);
  const [agesMode, setAgesMode] = useState<Bound>('max');
  // Regions only narrow the venue list. They are not saved with the filter.
  const [selectedRegions, setSelectedRegions] = useState<string[]>([]);
  const [selectedVenues, setSelectedVenues] = useState<string[]>([]);
  const [venuesMode, setVenuesMode] = useState<Apply>('include');

  const [creating, setCreating] = useState(false);

  useEffect(() => {
    async function loadOptions() {
      try {
        setFilterOptions(await fetchFilterOptions());
        setLoadError(null);
      } catch (err) {
        setLoadError(toLoadError(err));
      } finally {
        setLoading(false);
      }
    }

    loadOptions();
  }, [attempt]);

  const retryLoad = () => {
    setLoading(true);
    setAttempt((n) => n + 1);
  };

  const venues = filterOptions?.venues ?? [];
  const shownVenues = visibleVenues(venues, selectedRegions);
  const total = selectedTags.length + selectedAges.length + selectedRegions.length + selectedVenues.length;

  const toggleRegion = (name: string) => {
    const regions = toggle(selectedRegions, name);
    setSelectedRegions(regions);
    setSelectedVenues((current) => dropHiddenVenues(current, venues, regions));
  };

  const handleCreateFilter = async () => {
    setCreating(true);
    setError(null);

    try {
      const result = await createFilter({
        tags: selectedTags,
        tagsInclude: tagsMode === 'include',
        ages: selectedAges,
        agesLessThan: agesMode === 'max',
        venues: selectedVenues,
        venuesInclude: venuesMode === 'include',
      });
      navigate(`/?filterKey=${result.filterKey}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create filter');
    } finally {
      setCreating(false);
    }
  };

  if (loading) {
    return (
      <Page title="Filters">
        <LoadingBlock label="Loading filter options" />
      </Page>
    );
  }

  if (loadError && !filterOptions) {
    return (
      <Page title="Filters">
        <LoadFailed error={loadError} onRetry={retryLoad} />
      </Page>
    );
  }

  const disabled = creating || total === 0;

  return (
    <Page
      title="Filters"
      subtitle="Choose what to show. You get a link you can share or subscribe to."
      flush
      footer={
        <div>
          {error && <p className="mb-2 text-sm text-danger">{error}</p>}
          <button
            type="button"
            onClick={handleCreateFilter}
            disabled={disabled}
            className="font-label chamfer flex h-12 w-full cursor-pointer items-center justify-center gap-3 rounded-control bg-accent text-sm font-semibold text-accent-ink enabled:hover:bg-accent-hover disabled:cursor-not-allowed disabled:bg-raised disabled:text-text-faint"
          >
            {creating ? 'Creating filter...' : 'Create filter'}
            <span className="font-data text-xs font-normal opacity-80">{total} selected</span>
          </button>
        </div>
      }
    >
      {filterOptions && filterOptions.tags.length > 0 && (
        <FilterGroup
          title="Tags"
          count={selectedTags.length}
          mode={<Segmented label="Tags apply" value={tagsMode} options={includeOptions} onChange={setTagsMode} />}
        >
          <div className="flex flex-wrap gap-2">
            {filterOptions.tags.map((tag) => (
              <Chip
                key={tag.name}
                label={displayTag(tag.name)}
                selected={selectedTags.includes(tag.name)}
                onToggle={() => setSelectedTags((current) => toggle(current, tag.name))}
              />
            ))}
          </div>
        </FilterGroup>
      )}

      {filterOptions && filterOptions.ages.length > 0 && (
        <FilterGroup
          title="Age"
          count={selectedAges.length}
          mode={<Segmented label="Age applies" value={agesMode} options={ageOptions} onChange={setAgesMode} />}
        >
          <div className="flex flex-wrap gap-2">
            {filterOptions.ages.map((age) => (
              <Chip
                key={age.value}
                label={displayAge(age.name)}
                selected={selectedAges.includes(age.value)}
                onToggle={() => setSelectedAges((current) => toggle(current, age.value))}
              />
            ))}
          </div>
        </FilterGroup>
      )}

      {filterOptions && filterOptions.regions.length > 0 && (
        <FilterGroup title="Region" count={selectedRegions.length}>
          <div className="flex flex-wrap gap-2">
            {filterOptions.regions.map((region) => (
              <Chip
                key={region.name}
                label={region.name}
                selected={selectedRegions.includes(region.name)}
                onToggle={() => toggleRegion(region.name)}
              />
            ))}
          </div>
        </FilterGroup>
      )}

      {venues.length > 0 && (
        <FilterGroup
          title="Venues"
          count={selectedVenues.length}
          hint={venueHint(selectedRegions)}
          mode={<Segmented label="Venues apply" value={venuesMode} options={includeOptions} onChange={setVenuesMode} />}
        >
          <VenueList
            venues={shownVenues}
            selected={selectedVenues}
            onToggle={(name) => setSelectedVenues((current) => toggle(current, name))}
          />
        </FilterGroup>
      )}
    </Page>
  );
}
