import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { fetchVenues } from '../services/api';
import type { Venue } from '../types/event';
import Page from '../components/ui/Page';
import Markdown from '../components/ui/Markdown';
import EmbedFrame from '../components/ui/EmbedFrame';
import { usePageOverride } from '../hooks/usePageOverride';
import { useSiteStore } from '../store/site';

// The introduction is the deployment's own about page if it has one. How to add an event, the
// contact address and the venues are the site's and stay either way.
export default function AboutPage() {
  const { site } = useSiteStore();
  const { page: own, ready } = usePageOverride('about');
  const [venues, setVenues] = useState<Venue[]>([]);

  useEffect(() => {
    async function loadVenues() {
      try {
        setVenues(await fetchVenues());
      } catch {
        // Venues are optional here.
      }
    }

    loadVenues();
  }, []);

  return (
    <Page title="About">
      <div className="space-y-4 leading-relaxed text-text-muted">
        {ready &&
          (own ? (
            <>
              <Markdown>{own.body}</Markdown>
              {own.embed && <EmbedFrame url={own.embed} title={own.title} />}
            </>
          ) : (
            <>
              <p>
                <span className="font-semibold text-text">{site.name}</span>
                {site.region && ` ${site.region}`}. A community-driven event calendar for{' '}
                {site.region || 'your area'}. We make it easy to find live music, comedy, art and community
                events around town.
              </p>
              <p>
                It&apos;s maintained by volunteers. Event information comes from venue websites and community
                submissions. We try to keep it accurate, but always confirm details on the original event link.
              </p>
            </>
          ))}

        <h2 className="font-label pt-2 text-lg font-semibold text-text">How to add an event</h2>
        <p>
          Anyone can submit an event on the <Link to="/add-event">Add event</Link> page. Verify by email,
          then an admin reviews and approves it.
        </p>

        {site.contactEmail && (
          <>
            <h2 className="font-label pt-2 text-lg font-semibold text-text">Contact</h2>
            <p>
              Questions, corrections, suggestions? <a href={`mailto:${site.contactEmail}`}>{site.contactEmail}</a>
            </p>
          </>
        )}

        {venues.length > 0 && (
          <>
            <h2 className="font-label flex items-baseline gap-2 pt-2 text-lg font-semibold text-text">
              Featured venues
              <span className="font-data text-xs font-normal text-text-faint">{venues.length} venues</span>
            </h2>
            <ul className="flex flex-wrap gap-2">
              {venues.map((venue) => (
                <li key={venue.slug}>
                  <a
                    href={venue.link}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="inline-flex min-h-9 items-center rounded-control border border-border px-3 text-sm"
                  >
                    {venue.name}
                  </a>
                </li>
              ))}
            </ul>
          </>
        )}
      </div>
    </Page>
  );
}
