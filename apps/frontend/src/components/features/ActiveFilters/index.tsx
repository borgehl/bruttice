import { useState } from 'react';
import { Link } from 'react-router-dom';
import type { FilterDetails } from '../../../services/api';

interface ActiveFiltersProps {
  filterKey: string;
  filterDetails: FilterDetails;
}

export default function ActiveFilters({ filterKey, filterDetails }: ActiveFiltersProps) {
  const [copiedFeed, setCopiedFeed] = useState<string | null>(null);

  const copyFeedUrl = (feedType: string) => {
    const url = `${window.location.origin}/api/feeds?feedType=${feedType}&filterKey=${filterKey}`;
    navigator.clipboard.writeText(url);
    setCopiedFeed(feedType);
    setTimeout(() => setCopiedFeed(null), 2000);
  };

  return (
    <div className="mb-6 p-4 bg-surface-elevated border border-border rounded-lg">
      <div className="flex items-center justify-between mb-2">
        <h3 className="font-semibold text-primary">Active Filters</h3>
        <Link
          to="/"
          className="text-sm text-text-muted hover:text-primary underline"
        >
          Clear filters
        </Link>
      </div>
      <div className="flex flex-wrap gap-2 mb-4">
        {filterDetails.tags && filterDetails.tags.length > 0 && (
          <div className="flex items-center gap-1">
            <span className="text-xs text-text-dim">
              Tags {filterDetails.tagsInclude ? '(include)' : '(exclude)'}:
            </span>
            {filterDetails.tags.map((tag) => (
              <span
                key={tag}
                className="px-2 py-0.5 text-xs bg-primary/20 text-primary rounded-full"
              >
                {tag}
              </span>
            ))}
          </div>
        )}
        {filterDetails.venues && filterDetails.venues.length > 0 && (
          <div className="flex items-center gap-1">
            <span className="text-xs text-text-dim">
              Venues {filterDetails.venuesInclude ? '(include)' : '(exclude)'}:
            </span>
            {filterDetails.venues.map((venue) => (
              <span
                key={venue}
                className="px-2 py-0.5 text-xs bg-primary/20 text-primary rounded-full"
              >
                {venue}
              </span>
            ))}
          </div>
        )}
        {filterDetails.age !== undefined && filterDetails.age > 0 && (
          <div className="flex items-center gap-1">
            <span className="text-xs text-text-dim">
              Age {filterDetails.agesLessThan ? '<=' : '>='} {filterDetails.age}+
            </span>
          </div>
        )}
      </div>
      <div className="flex items-center gap-2">
        <span className="text-xs text-text-dim">Copy feed URL:</span>
        {['rss', 'atom', 'json'].map((feedType) => (
          <button
            key={feedType}
            onClick={() => copyFeedUrl(feedType)}
            className="px-3 py-1 text-xs bg-surface border border-border rounded hover:border-primary text-text-muted hover:text-primary transition-colors"
          >
            {copiedFeed === feedType ? 'Copied!' : feedType.toUpperCase()}
          </button>
        ))}
      </div>
    </div>
  );
}
