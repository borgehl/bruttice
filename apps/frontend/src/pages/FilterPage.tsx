import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { fetchFilterOptions, createFilter } from '../services/api';
import type { FilterOptions } from '../types/event';

export default function FilterPage() {
  const navigate = useNavigate();

  const [filterOptions, setFilterOptions] = useState<FilterOptions | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [selectedTags, setSelectedTags] = useState<string[]>([]);
  const [tagsInclude, setTagsInclude] = useState(true);
  const [selectedVenues, setSelectedVenues] = useState<string[]>([]);
  const [venuesInclude, setVenuesInclude] = useState(true);
  const [selectedAges, setSelectedAges] = useState<number[]>([]);
  const [agesLessThan, setAgesLessThan] = useState(true);

  const [creating, setCreating] = useState(false);

  useEffect(() => {
    async function loadOptions() {
      try {
        setLoading(true);
        const data = await fetchFilterOptions();
        setFilterOptions(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load filter options');
      } finally {
        setLoading(false);
      }
    }

    loadOptions();
  }, []);

  const handleCreateFilter = async () => {
    setCreating(true);
    setError(null);

    try {
      const result = await createFilter({
        tags: selectedTags,
        tagsInclude,
        ages: selectedAges,
        agesLessThan,
        venues: selectedVenues,
        venuesInclude,
      });
      navigate(`/?filterKey=${result.filterKey}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create filter');
    } finally {
      setCreating(false);
    }
  };

  const toggleTag = (tagName: string) => {
    setSelectedTags((prev) =>
      prev.includes(tagName) ? prev.filter((name) => name !== tagName) : [...prev, tagName]
    );
  };

  const toggleVenue = (venueName: string) => {
    setSelectedVenues((prev) =>
      prev.includes(venueName) ? prev.filter((v) => v !== venueName) : [...prev, venueName]
    );
  };

  const toggleAge = (ageValue: number) => {
    setSelectedAges((prev) =>
      prev.includes(ageValue) ? prev.filter((v) => v !== ageValue) : [...prev, ageValue]
    );
  };

  if (loading) {
    return (
      <div className="text-center py-12">
        <div className="inline-block animate-spin rounded-full h-8 w-8 border-4 border-primary border-t-transparent"></div>
        <p className="mt-4 text-text-muted">Loading filter options...</p>
      </div>
    );
  }

  if (error && !filterOptions) {
    return (
      <div className="p-4 bg-error-muted border border-error/30 rounded-lg">
        <p className="text-error">{error}</p>
      </div>
    );
  }

  return (
    <div className="max-w-3xl mx-auto">
      <h1 className="text-2xl font-bold mb-6 text-text">Create Custom Filter</h1>

      <div className="mb-6 p-4 bg-info-muted border border-info/30 rounded-lg">
        <p className="text-sm text-info">
          Create a custom filter to see only the events you're interested in. You can filter by tags,
          venues, and age restrictions. Once created, you'll get a shareable URL for your filtered calendar.
        </p>
      </div>

      {error && (
        <div className="mb-6 p-4 bg-error-muted border border-error/30 rounded-lg">
          <p className="text-error">{error}</p>
        </div>
      )}

      <div className="space-y-6">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {filterOptions?.tags && filterOptions.tags.length > 0 && (
            <section className="p-4 bg-surface border border-border rounded-lg">
              <div className="flex items-center justify-between mb-3">
                <h2 className="text-lg font-semibold text-text">Tags</h2>
                <select
                  onChange={(e) => setTagsInclude(e.target.value == "true")}
                  className="text-sm border border-border rounded px-2 py-1 bg-surface-elevated text-text"
                >
                  <option value="true">Include</option>
                  <option value="false">Exclude</option>
                </select>
              </div>
              <div className="grid grid-cols-2 gap-2">
                {filterOptions.tags.map((tag) => (
                  <button
                    key={tag.name}
                    type="button"
                    onClick={() => toggleTag(tag.name)}
                    className={`px-3 py-2 text-sm rounded border transition-colors text-left ${
                      selectedTags.includes(tag.name)
                        ? 'bg-primary text-white border-primary'
                        : 'bg-surface-elevated text-text-muted border-border hover:border-primary'
                    }`}
                  >
                    {tag.name}
                  </button>
                ))}
              </div>
            </section>
          )}

          {filterOptions?.ages && filterOptions.ages.length > 0 && (
            <section className="p-4 bg-surface border border-border rounded-lg">
              <div className="flex items-center justify-between mb-3">
                <h2 className="text-lg font-semibold text-text">Age Restrictions</h2>
                <select
                  onChange={(e) => setAgesLessThan(e.target.value == "true")}
                  className="text-sm border border-border rounded px-2 py-1 bg-surface-elevated text-text"
                >
                  <option value="true">&le; (max)</option>
                  <option value="false">&ge; (min)</option>
                </select>
              </div>
              <div className="grid grid-cols-2 gap-2">
                {filterOptions.ages.map((age) => (
                  <button
                    key={age.value}
                    type="button"
                    onClick={() => toggleAge(age.value)}
                    className={`px-3 py-2 text-sm rounded border transition-colors text-left ${
                      selectedAges.includes(age.value)
                        ? 'bg-primary text-white border-primary'
                        : 'bg-surface-elevated text-text-muted border-border hover:border-primary'
                    }`}
                  >
                    {age.name}
                  </button>
                ))}
              </div>
            </section>
          )}
        </div>

        {filterOptions?.venues && filterOptions.venues.length > 0 && (
          <section className="p-4 bg-surface border border-border rounded-lg">
            <div className="flex items-center justify-between mb-3">
              <h2 className="text-lg font-semibold text-text">Venues</h2>
              <select
                onChange={(e) => setVenuesInclude(e.target.value == "true")}
                className="text-sm border border-border rounded px-2 py-1 bg-surface-elevated text-text"
              >
                <option value="true">Include</option>
                <option value="false">Exclude</option>
              </select>
            </div>
            <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-2 max-h-64 overflow-y-auto">
              {filterOptions.venues.map((venue) => (
                <button
                  key={venue.slug}
                  type="button"
                  onClick={() => toggleVenue(venue.name)}
                  className={`px-3 py-2 text-sm rounded border transition-colors text-left truncate ${
                    selectedVenues.includes(venue.name)
                      ? 'bg-primary text-white border-primary'
                      : 'bg-surface-elevated text-text-muted border-border hover:border-primary'
                  }`}
                  title={venue.name}
                >
                  {venue.name}
                </button>
              ))}
            </div>
          </section>
        )}

        <button
          onClick={handleCreateFilter}
          disabled={creating || (selectedTags.length === 0 && selectedVenues.length === 0 && selectedAges.length === 0)}
          className="w-full px-4 py-3 bg-primary text-white font-medium rounded-md hover:bg-primary-hover disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {creating ? 'Creating Filter...' : 'Create Filter'}
        </button>
      </div>
    </div>
  );
}
