import { useState } from 'react';
import { Link } from 'react-router-dom';

export default function FeedsPage() {
  const [copiedFeed, setCopiedFeed] = useState<string | null>(null);

  const baseUrl = window.location.origin;

  const copyFeedUrl = (feedType: string, url: string) => {
    navigator.clipboard.writeText(url);
    setCopiedFeed(feedType);
    setTimeout(() => setCopiedFeed(null), 2000);
  };

  const feeds = [
    {
      type: 'RSS',
      description: 'Standard RSS 2.0 feed, compatible with most feed readers.',
      url: `${baseUrl}/api/feeds?feedType=rss`,
    },
    {
      type: 'Atom',
      description: 'Atom 1.0 feed format.',
      url: `${baseUrl}/api/feeds?feedType=atom`,
    },
    {
      type: 'JSON',
      description: 'JSON feed format for programmatic access.',
      url: `${baseUrl}/api/feeds?feedType=json`,
    },
  ];

  return (
    <div className="max-w-2xl mx-auto">
      <h1 className="text-2xl font-bold mb-6 text-text">Event Feeds</h1>

      <div className="mb-6 p-4 bg-info-muted border border-info/30 rounded-lg">
        <p className="text-sm text-info">
          Subscribe to our event feeds to stay updated on upcoming events.
          You can use these feeds with your favorite feed reader or calendar app.
        </p>
      </div>

      <div className="space-y-4">
        {feeds.map((feed) => (
          <div key={feed.type} className="p-4 bg-surface border border-border rounded-lg">
            <h2 className="text-lg font-semibold mb-1 text-text">{feed.type} Feed</h2>
            <p className="text-sm text-text-muted mb-3">{feed.description}</p>
            <div className="flex items-center gap-2">
              <input
                type="text"
                readOnly
                value={feed.url}
                className="flex-1 px-3 py-2 text-sm bg-surface-elevated border border-border rounded text-text"
              />
              <button
                onClick={() => copyFeedUrl(feed.type, feed.url)}
                className="px-4 py-2 text-sm bg-primary text-white rounded hover:bg-primary-hover min-w-[70px]"
              >
                {copiedFeed === feed.type ? 'Copied!' : 'Copy'}
              </button>
            </div>
          </div>
        ))}
      </div>

      <div className="mt-8 p-4 bg-surface border border-border rounded-lg">
        <h2 className="text-lg font-semibold mb-2 text-text">Custom Filtered Feeds</h2>
        <p className="text-sm text-text-muted mb-3">
          Want to subscribe to a filtered feed? Create a custom filter first, then use the generated
          feed URL with your filter key.
        </p>
        <Link
          to="/filter"
          className="inline-block px-4 py-2 bg-primary text-white text-sm rounded hover:bg-primary-hover"
        >
          Create Custom Filter
        </Link>
      </div>
    </div>
  );
}
