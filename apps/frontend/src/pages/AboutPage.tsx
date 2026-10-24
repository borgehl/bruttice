import { Link } from 'react-router-dom';
import Page from '../components/ui/Page';
import Markdown from '../components/ui/Markdown';
import EmbedFrame from '../components/ui/EmbedFrame';
import { usePageOverride } from '../hooks/usePageOverride';
import { useSiteStore } from '../store/site';

// The introduction is the deployment's own about page if it has one. How to add an event and the
// contact link are the site's and stay either way.
export default function AboutPage() {
  const { site } = useSiteStore();
  const { page: own, ready } = usePageOverride('about');

  return (
    <Page title="About">
      <div className="space-y-4 leading-relaxed text-text-muted">
        {ready &&
          (own ? (
            <>
              <Markdown>{own.body}</Markdown>
              {(own.embed || own.embedHtml) && <EmbedFrame url={own.embed} html={own.embedHtml} title={own.title} />}
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

        {site.contactForm && (
          <>
            <h2 className="font-label pt-2 text-lg font-semibold text-text">Contact</h2>
            <p>
              Questions, corrections, suggestions? <Link to="/contact">Send us a message</Link>.
            </p>
          </>
        )}
      </div>
    </Page>
  );
}
