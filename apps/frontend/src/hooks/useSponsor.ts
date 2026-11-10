import { useEffect, useState } from 'react';
import { fetchSponsors } from '../services/api';
import type { Sponsor } from '../types/sponsor';
import { pickSponsor } from '../utils/sponsor';

// The sponsor for this page load, null while it loads or if no card is running. A failure is not
// shown: the calendar is the same without the card.
export function useSponsor(): Sponsor | null {
  const [sponsor, setSponsor] = useState<Sponsor | null>(null);

  useEffect(() => {
    let cancelled = false;
    fetchSponsors()
      .then((sponsors) => {
        if (!cancelled) setSponsor(pickSponsor(sponsors));
      })
      .catch(() => undefined);
    return () => {
      cancelled = true;
    };
  }, []);

  return sponsor;
}
