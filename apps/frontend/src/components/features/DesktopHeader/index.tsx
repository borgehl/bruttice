import { Link, NavLink, useLocation } from 'react-router-dom';
import ThemeSwitchCompact from '../ThemeSwitchCompact';
import { useAuthStore } from '../../../store/auth';
import { useSiteStore } from '../../../store/site';
import { navPages, usePagesStore } from '../../../store/pages';
import { isActive, linkFor } from '../../../utils/nav';

interface NavItem {
  to: string;
  label: string;
  also?: string[];
}

// The top of the site from tablet width up. Below that the bottom nav takes over.
export default function DesktopHeader() {
  const { pathname, search } = useLocation();
  const filterKey = new URLSearchParams(search).get('filterKey');
  const { isAuthenticated } = useAuthStore();
  const { site } = useSiteStore();
  const { pages } = usePagesStore();

  const items: NavItem[] = [
    { to: '/', label: 'Calendar', also: ['/calendar'] },
    { to: '/add-event', label: 'Add event' },
    { to: '/filter', label: 'Filters' },
    { to: '/feeds', label: 'Feeds' },
    { to: '/about', label: 'About' },
    ...navPages(pages).map((page) => ({ to: `/p/${page.slug}`, label: page.title })),
    ...(isAuthenticated ? [{ to: '/admin', label: 'Admin' }] : []),
  ];

  return (
    <header className="hidden shrink-0 border-b border-border bg-surface md:block">
      <div className="mx-auto max-w-4xl px-4 py-4">
        <div className="mb-3 flex items-start justify-between gap-4">
          <h1 className="font-label text-2xl font-bold text-accent in-data-[theme=neon]:glow-text">
            <Link to="/" className="no-underline hover:text-accent-hover">
              {site.name}
              {site.region && (
                <span className="font-medium text-text-muted">
                  <span className="in-data-[theme=neon]:hidden"> - </span>
                  <span className="hidden in-data-[theme=neon]:inline"> // </span>
                  {site.region}
                </span>
              )}
            </Link>
          </h1>
          <ThemeSwitchCompact />
        </div>

        <div className="mb-3 rounded-control border border-warning/10 bg-warning-muted px-3 py-1.5">
          <p className="text-sm text-warning">Always confirm details by checking event links.</p>
        </div>

        <nav aria-label="Main" className="flex flex-wrap gap-x-5 gap-y-1">
          {items.map((item) => {
            const active = isActive(pathname, item.to, item.also);
            return (
              <NavLink
                key={item.to}
                to={linkFor(item.to, search, filterKey)}
                aria-current={active ? 'page' : undefined}
                className={`font-label flex min-h-9 items-center border-b-2 text-sm ${
                  active
                    ? 'glow-text border-accent font-semibold text-accent'
                    : 'border-transparent text-text-muted hover:text-accent'
                }`}
              >
                {item.label}
              </NavLink>
            );
          })}
        </nav>
      </div>
    </header>
  );
}
