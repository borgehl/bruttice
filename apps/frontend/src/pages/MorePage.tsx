import { Link } from 'react-router-dom';
import Page from '../components/ui/Page';
import ThemeSwitch from '../components/features/ThemeSwitch';
import { useAuthStore } from '../store/auth';
import { useSiteStore } from '../store/site';
import { navPages, usePagesStore } from '../store/pages';

interface Row {
  to: string;
  label: string;
  sub: string;
}

export default function MorePage() {
  const { isAuthenticated } = useAuthStore();
  const { site } = useSiteStore();
  const { pages } = usePagesStore();

  const rows: Row[] = [
    { to: '/about', label: 'About', sub: 'Who runs this, venues' },
    ...navPages(pages).map((page) => ({ to: `/p/${page.slug}`, label: page.title, sub: '' })),
    ...(site.contactForm ? [{ to: '/contact', label: 'Contact', sub: 'Questions, corrections, suggestions' }] : []),
    { to: '/privacy-policy', label: 'Privacy policy', sub: 'Analytics cookie' },
    isAuthenticated
      ? { to: '/admin', label: 'Admin', sub: 'Events and ingestion' }
      : { to: '/admin/login', label: 'Admin login', sub: 'Organizers only' },
  ];

  return (
    <Page title="More">
      <ThemeSwitch />

      <ul>
        {rows.map((row) => (
          <li key={row.to} className="border-b border-border">
            <Link to={row.to} className="flex items-center justify-between gap-4 py-3.5 text-text hover:text-text">
              <span>
                <span className="block font-medium">{row.label}</span>
                {row.sub && <span className="block text-sm text-text-faint">{row.sub}</span>}
              </span>
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true">
                <path d="M9 6l6 6-6 6" />
              </svg>
            </Link>
          </li>
        ))}
      </ul>

      <p className="mt-6 text-sm leading-relaxed text-text-muted">{site.description}</p>
    </Page>
  );
}
