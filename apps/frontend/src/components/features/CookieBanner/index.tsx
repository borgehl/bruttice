import { useState, useRef, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { postVisitorConsent } from '../../../services/api';

export default function CookieBanner() {
  const [visible, setVisible] = useState(() => {
    return localStorage.getItem('tracking-consent') === null;
  });
  const bannerRef = useRef<HTMLDivElement>(null);
  const [bannerHeight, setBannerHeight] = useState(0);

  useEffect(() => {
    if (!visible || !bannerRef.current) return;
    const el = bannerRef.current;
    const update = () => setBannerHeight(el.offsetHeight);
    update();
    const observer = new ResizeObserver(update);
    observer.observe(el);
    return () => observer.disconnect();
  }, [visible]);

  if (!visible) return null;

  async function handleAccept() {
    try {
      await postVisitorConsent();
    } catch {
      // Non-fatal: consent endpoint failure doesn't block the UI choice.
    }
    localStorage.setItem('tracking-consent', 'accepted');
    setVisible(false);
  }

  function handleDecline() {
    localStorage.setItem('tracking-consent', 'declined');
    setVisible(false);
  }

  return (
    <>
      {/* Spacer prevents the fixed banner from covering page content */}
      <div style={{ height: bannerHeight }} aria-hidden="true" />
      <div
        ref={bannerRef}
        className="fixed bottom-0 left-0 right-0 z-50 bg-surface border-t border-border shadow-lg"
      >
        <div className="max-w-4xl mx-auto px-4 py-4 flex flex-col sm:flex-row gap-4 items-start sm:items-center justify-between">
          <p className="text-sm text-text-muted flex-1">
            We use a single analytics cookie to count unique visitors and understand which
            cities our audience comes from. No personal information is collected or sold.{' '}
            <Link to="/privacy-policy" className="text-primary hover:text-primary-hover underline">
              Privacy Policy
            </Link>
          </p>
          <div className="flex gap-3 shrink-0">
            <button
              onClick={handleDecline}
              className="px-4 py-2 text-sm border border-border rounded hover:bg-surface text-text-muted"
            >
              Decline
            </button>
            <button
              onClick={handleAccept}
              className="px-4 py-2 text-sm border border-border rounded hover:bg-surface text-text-muted"
            >
              Accept
            </button>
          </div>
        </div>
      </div>
    </>
  );
}
