import { useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import Page from '../components/ui/Page';
import { ghostButtonClass } from '../components/ui/Field';
import { feedUrl } from '../utils/feeds';

const feeds = [
  { type: 'RSS', key: 'rss', description: 'Standard RSS 2.0, works with most feed readers.' },
  { type: 'Atom', key: 'atom', description: 'Atom 1.0 feed format.' },
  { type: 'JSON', key: 'json', description: 'JSON feed for programmatic access.' },
];

export default function FeedsPage() {
  const [searchParams] = useSearchParams();
  const filterKey = searchParams.get('filterKey');
  const [copiedFeed, setCopiedFeed] = useState<string | null>(null);

  const copyFeedUrl = async (feedType: string, url: string) => {
    try {
      await navigator.clipboard.writeText(url);
      setCopiedFeed(feedType);
      setTimeout(() => setCopiedFeed(null), 2000);
    } catch {
      // Clipboard access can be refused. The URL is on screen to copy by hand.
    }
  };

  return (
    <Page
      title="Feeds"
      subtitle={
        filterKey
          ? 'These feeds follow the filter you are looking at.'
          : 'Subscribe in your feed reader or calendar app.'
      }
    >
      <ul className="space-y-3">
        {feeds.map((feed) => {
          const url = feedUrl(window.location.origin, feed.key, filterKey);
          return (
            <li key={feed.key} className="rounded-control border border-border bg-surface p-4">
              <div className="mb-1 flex items-center gap-2">
                <h2 className="font-label text-base font-semibold text-text">{feed.type}</h2>
                <span className="font-data text-xs text-text-faint">.{feed.key}</span>
              </div>
              <p className="mb-3 text-sm text-text-muted">{feed.description}</p>
              <div className="flex items-center gap-2">
                <input
                  type="text"
                  readOnly
                  aria-label={`${feed.type} feed URL`}
                  value={url}
                  onFocus={(event) => event.currentTarget.select()}
                  className="font-data h-10 min-w-0 flex-1 rounded-control border border-border bg-raised px-3 text-xs text-text"
                />
                <button
                  type="button"
                  onClick={() => copyFeedUrl(feed.key, url)}
                  className="font-label h-10 min-w-[72px] cursor-pointer rounded-control border border-accent px-3 text-sm font-semibold text-accent hover:text-accent-hover"
                >
                  {copiedFeed === feed.key ? 'Copied' : 'Copy'}
                </button>
              </div>
            </li>
          );
        })}
      </ul>

      <section className="mt-6">
        <h2 className="font-label mb-1 text-base font-semibold text-text">Filtered feeds</h2>
        <p className="mb-3 text-sm text-text-muted">
          Want only some events? Create a filter, then open Feeds from the calendar to copy its feed link.
        </p>
        <Link to="/filter" className={`${ghostButtonClass} hover:text-accent-hover`}>
          Create custom filter
        </Link>
      </section>
    </Page>
  );
}
