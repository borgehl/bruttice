import { Navigate } from 'react-router-dom';
import Page from '../components/ui/Page';
import { useSiteStore } from '../store/site';

// Only deployments that turn it on in their config have this page.
export default function PopBookerPage() {
  const { site, loaded } = useSiteStore();

  if (loaded && !site.features.popBooker) return <Navigate to="/more" replace />;

  return (
    <Page title="Pop! Booker" subtitle="DIY and independent booking platform">
      <div className="space-y-4 leading-relaxed text-text-muted">
        <p>
          Pop! Booker was born from Brutus wanting a better way to book and organize shows. Messages are
          scattered across email, text and IG DMs, jumbled up with password resets and marketing. Contacts
          live in phone numbers, handles and an out-of-sync spreadsheet. There had to be a better way to
          organize around an event.
        </p>
        <p>
          The guiding principle: it has to be useful if you&apos;re the only person on it. It&apos;s not
          another social platform promising features once everyone joins. It&apos;s an organizing tool worth
          using today.
        </p>
        <p>
          Brutus uses it for the monthly <span className="font-semibold text-text">T0y0tuh Kn1ghts</span>{' '}
          showcase at <span className="font-semibold text-text">Whittier Bar</span>. It&apos;s early, so
          expect fast iteration. Free to sign up. Pitch an event to the TK team, or DM on IG if you&apos;re
          curious.
        </p>

        {site.popBookerUrl && (
          <section className="pt-2">
            <h2 className="font-label mb-1 text-lg font-semibold text-text">TK calendar</h2>
            <p className="mb-3 text-sm">
              Linked to Pop! Booker. Many export options, and the API is public so a website dev can style it.
            </p>
            <iframe
              src={site.popBookerUrl}
              title="TK calendar"
              loading="lazy"
              sandbox="allow-scripts allow-same-origin allow-popups"
              className="h-[420px] w-full rounded-control border border-border bg-surface"
            />
            <a
              href={site.popBookerUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="mt-2 inline-block text-sm"
            >
              Open in a new tab
            </a>
          </section>
        )}
      </div>
    </Page>
  );
}
