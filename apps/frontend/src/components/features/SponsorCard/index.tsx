import { useEffect, useRef, useState } from 'react';
import type { Sponsor } from '../../../types/sponsor';
import { sendSponsorImpression } from '../../../services/api';

interface SponsorCardProps {
  sponsor: Sponsor;
}

// A paid card that sits in the calendar like an event row, and says so. It is one still picture and
// two lines of text: nothing moves, opens or follows the visitor. The picture has the same box as a
// flyer, so the card is the same height whether or not it loads.
export default function SponsorCard({ sponsor }: SponsorCardProps) {
  const ref = useRef<HTMLElement>(null);
  const [imageFailed, setImageFailed] = useState(false);

  // Counted once, when at least half of the card has been on screen. Nothing about the visitor is sent.
  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    const observer = new IntersectionObserver(
      (entries) => {
        if (!entries.some((entry) => entry.isIntersecting)) return;
        observer.disconnect();
        sendSponsorImpression(sponsor.id);
      },
      { threshold: 0.5 },
    );
    observer.observe(el);
    return () => observer.disconnect();
  }, [sponsor.id]);

  const showImage = sponsor.imageUrl !== '' && !imageFailed;

  return (
    <aside ref={ref} aria-label={`Sponsored: ${sponsor.name}`} className="flex gap-3.5 border-b border-border bg-raised px-4 py-3.5">
      {/* no column of its own for the label: the text starts at the left edge, and the label sits at the
          bottom left, under the text, level with the bottom of the picture */}
      <div className="flex min-w-0 flex-1 flex-col gap-0.5">
        <h3 className="text-base font-semibold leading-snug text-pretty">
          <a href={sponsor.clickUrl} target="_blank" rel="sponsored noopener noreferrer" className="text-text hover:text-accent">
            {sponsor.name}
          </a>
        </h3>
        {sponsor.text && <p className="text-sm text-text-muted">{sponsor.text}</p>}
        <span className="font-data mt-auto pt-1.5 text-xs uppercase tracking-wide text-text-faint">Sponsored</span>
      </div>

      {showImage && (
        <a
          href={sponsor.clickUrl}
          target="_blank"
          rel="sponsored noopener noreferrer"
          tabIndex={-1}
          aria-hidden="true"
          className="h-[88px] w-16 shrink-0 overflow-hidden rounded-control border border-border bg-bg"
        >
          <img
            src={sponsor.imageUrl}
            alt=""
            width={64}
            height={88}
            loading="lazy"
            onError={() => setImageFailed(true)}
            className="h-full w-full object-cover"
          />
        </a>
      )}
    </aside>
  );
}
