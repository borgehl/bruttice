import { useEffect, useState } from 'react';
import { fetchVenues } from '../services/api';
import type { Venue } from '../types/event';

export default function AboutPage() {
  const [venues, setVenues] = useState<Venue[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function loadVenues() {
      try {
        const data = await fetchVenues();
        setVenues(data);
      } catch {
        // Silently fail - venues are optional
      } finally {
        setLoading(false);
      }
    }

    loadVenues();
  }, []);

  return (
    <div className="max-w-2xl mx-auto">
      <h1 className="text-2xl font-bold mb-6 text-text">About</h1>

      <div className="space-y-4 text-text-muted">
        <p>
          A community-driven event calendar. Our goal is to make it easy for locals to discover
          live music, comedy, art, and community events happening around town.
        </p>

        <p>
          Event information is gathered from various sources including venue websites and community
          submissions. We do our best to keep information accurate, but always recommend confirming
          event details by clicking through to the original event link.
        </p>

        <h2 className="text-xl font-semibold mt-8 mb-4 text-text">How to Add an Event</h2>
        <p>
          Anyone can submit an event to be added to the calendar. Simply visit the{' '}
          <a href="/add-event" className="text-primary hover:text-primary-hover">
            Add Event
          </a>{' '}
          page and fill out the form. You'll receive an email to verify your submission, and then an
          admin will review and approve it.
        </p>

        {!loading && venues.length > 0 && (
          <>
            <h2 className="text-xl font-semibold mt-8 mb-4 text-text">Featured Venues</h2>
            <p>
              We track events from {venues.length} venues:
            </p>
            <div className="grid grid-cols-2 sm:grid-cols-3 gap-2 mt-4">
              {venues.map((venue) => (
                <a
                  key={venue.slug}
                  href={venue.link}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-sm text-primary hover:text-primary-hover"
                >
                  {venue.name}
                </a>
              ))}
            </div>
          </>
        )}

      </div>
    </div>
  );
}
