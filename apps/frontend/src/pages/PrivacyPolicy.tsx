import Page from '../components/ui/Page';

export default function PrivacyPolicy() {
  return (
    <Page title="Privacy policy">
    <div className="text-sm leading-relaxed text-text-muted">

      <section className="mb-6">
        <h2 className="font-label mb-2 text-base font-semibold text-text">Who we are</h2>
        <p>
          A community event calendar.
        </p>
      </section>

      <section className="mb-6">
        <h2 className="font-label mb-2 text-base font-semibold text-text">What we collect</h2>
        <p>When you accept the cookie banner, we collect:</p>
        <ul className="list-disc list-inside mt-2 space-y-1">
          <li>
            A randomly generated visitor ID (UUID) stored in a cookie — this is
            not linked to your name, email, or any other identifying information.
          </li>
          <li>The pages you visit on this site.</li>
          <li>
            A coarse location (city, state, country) derived from your IP address
            at the moment you accept the cookie. We do <strong>not</strong> store
            your raw IP address.
          </li>
        </ul>
        <p className="mt-2">
          If you decline the cookie banner, none of the above is collected and no
          cookie is set.
        </p>
      </section>

      <section className="mb-6">
        <h2 className="font-label mb-2 text-base font-semibold text-text">Why we collect it</h2>
        <p>
          We use this data solely for anonymous aggregate analytics: how many unique
          visitors we receive per day, which pages are popular, and the general
          geographic distribution of our audience. We may cite these aggregate,
          anonymous statistics in business contexts such as pitching advertisers.
          We do not use individual visitor data to target or profile anyone, and
          we do not sell or share personal data with third parties.
        </p>
        <p className="mt-2">
          We also maintain a simple daily page-request counter for visitors who have
          not consented to tracking. This counter records only the date, the page
          path, and a running total — no IP address, no cookie, and no information
          that could identify you. It is not personal data and exists solely so we
          can understand what fraction of our traffic is untracked.
        </p>
      </section>

      <section className="mb-6">
        <h2 className="font-label mb-2 text-base font-semibold text-text">Legal basis</h2>
        <p>
          Processing is based on your explicit consent (GDPR Article 6(1)(a)). You
          may withdraw consent at any time by clearing your cookies; the banner will
          reappear on your next visit and no further data will be collected.
        </p>
      </section>

      <section className="mb-6">
        <h2 className="font-label mb-2 text-base font-semibold text-text">Data retention</h2>
        <p>
          Access logs are retained for 12 months and then deleted. Your visitor ID
          cookie expires after 1 year.
        </p>
      </section>

      <section className="mb-6">
        <h2 className="font-label mb-2 text-base font-semibold text-text">Your rights</h2>
        <ul className="list-disc list-inside space-y-1">
          <li>
            <strong>Withdraw consent:</strong> Clear your cookies. The banner will
            reappear and no further data will be collected.
          </li>
          <li>
            <strong>Data portability:</strong> Not applicable — the data we hold
            is anonymous and not attributable to you individually.
          </li>
        </ul>
      </section>

      <section className="mb-6">
        <h2 className="font-label mb-2 text-base font-semibold text-text">Cookies we set</h2>
        <div className="overflow-x-auto">
          <table className="w-full text-sm border-collapse">
            <thead>
              <tr className="border-b border-border">
                <th className="text-left py-2 pr-4 font-semibold">Name</th>
                <th className="text-left py-2 pr-4 font-semibold">Purpose</th>
                <th className="text-left py-2 pr-4 font-semibold">Type</th>
                <th className="text-left py-2 pr-4 font-semibold">Duration</th>
                <th className="text-left py-2 font-semibold">HttpOnly</th>
              </tr>
            </thead>
            <tbody>
              <tr className="border-b border-border">
                <td className="py-2 pr-4 font-mono">visitor-id</td>
                <td className="py-2 pr-4">Counts unique visitors; links page views anonymously</td>
                <td className="py-2 pr-4">Analytics</td>
                <td className="py-2 pr-4">1 year</td>
                <td className="py-2">Yes</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <section className="mb-6">
        <h2 className="font-label mb-2 text-base font-semibold text-text">Third parties</h2>
        <p>
          At the moment you accept the cookie banner, we perform a single IP
          geolocation lookup via{' '}
          <a
            href="https://ipgeolocation.io"
            target="_blank"
            rel="noopener noreferrer"
            className="text-accent hover:text-accent-hover"
          >
            ipgeolocation.io
          </a>{' '}
          to determine your approximate city, state, and country. Only those three
          fields are stored by us; your raw IP address is not retained.
        </p>
      </section>
    </div>
    </Page>
  );
}
