import { useState } from 'react';
import { Link } from 'react-router-dom';
import { postVisitorConsent } from '../../../services/api';

const CONSENT_KEY = 'tracking-consent';

function hasChosen(): boolean {
  try {
    return localStorage.getItem(CONSENT_KEY) !== null;
  } catch {
    // Without storage the choice can't be remembered, so ask once per visit.
    return false;
  }
}

function saveChoice(value: 'accepted' | 'declined'): void {
  try {
    localStorage.setItem(CONSENT_KEY, value);
  } catch {
    // The choice still applies for this visit.
  }
}

// Sits just above the bottom nav and pushes the page up instead of covering it.
export default function CookieBanner() {
  const [visible, setVisible] = useState(() => !hasChosen());

  if (!visible) return null;

  async function handleAccept() {
    try {
      await postVisitorConsent();
    } catch {
      // Non-fatal: consent endpoint failure doesn't block the UI choice.
    }
    saveChoice('accepted');
    setVisible(false);
  }

  function handleDecline() {
    saveChoice('declined');
    setVisible(false);
  }

  return (
    <section
      aria-label="Cookie consent"
      className="flex shrink-0 flex-col gap-3.5 border-t border-border bg-surface p-4 shadow-sheet"
    >
      <p className="text-sm leading-relaxed text-text">
        One analytics cookie counts unique visitors and which cities they come from. No personal info is
        collected or sold.{' '}
        <Link to="/privacy-policy" className="underline">
          Privacy policy
        </Link>
      </p>
      <div className="flex gap-2.5">
        <button
          type="button"
          onClick={handleDecline}
          className="font-label h-12 flex-1 cursor-pointer rounded-control border border-border text-sm font-semibold text-text-muted"
        >
          Decline
        </button>
        <button
          type="button"
          onClick={handleAccept}
          className="font-label chamfer h-12 flex-1 cursor-pointer rounded-control bg-accent text-sm font-semibold text-accent-ink"
        >
          Accept
        </button>
      </div>
    </section>
  );
}
