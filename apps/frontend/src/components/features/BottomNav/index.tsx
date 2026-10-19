import { NavLink, useLocation } from 'react-router-dom';

// Stroke paths on a 24 x 24 grid.
const icons = {
  calendar: 'M3 5h18v16H3zM3 10h18M8 3v4M16 3v4',
  add: 'M12 5v14M5 12h14',
  filters: 'M4 6h16M7 12h10M10 18h4',
  feeds: 'M5 12a7 7 0 0 1 7 7M5 5a14 14 0 0 1 14 14M6 19h.01',
  more: 'M5 12h.01M12 12h.01M19 12h.01',
};

const items = [
  { to: '/', label: 'Calendar', icon: icons.calendar, also: ['/calendar'] },
  { to: '/add-event', label: 'Add event', icon: icons.add, also: [] },
  { to: '/filter', label: 'Filters', icon: icons.filters, also: [] },
  { to: '/feeds', label: 'Feeds', icon: icons.feeds, also: [] },
  // More leads to everything that has no tab of its own.
  { to: '/more', label: 'More', icon: icons.more, also: ['/about', '/pop-booker', '/privacy-policy', '/admin'] },
];

function isActive(pathname: string, to: string, also: string[]): boolean {
  if (to === '/') return pathname === '/' || also.includes(pathname);
  return [to, ...also].some((path) => pathname === path || pathname.startsWith(`${path}/`));
}

// The calendar keeps its date and filter when you come back to it, and Feeds shows the feeds of the
// filter you are looking at.
function linkFor(to: string, search: string, filterKey: string | null) {
  if (to === '/') return { pathname: '/', search };
  if (to === '/feeds' && filterKey) return { pathname: '/feeds', search: `?filterKey=${encodeURIComponent(filterKey)}` };
  return to;
}

export default function BottomNav() {
  const { pathname, search } = useLocation();
  const filterKey = new URLSearchParams(search).get('filterKey');

  return (
    <nav
      aria-label="Main"
      className="flex shrink-0 border-t border-border bg-bg"
      style={{ height: 'var(--nav-height)' }}
    >
      {items.map((item) => {
        const active = isActive(pathname, item.to, item.also);
        return (
          <NavLink
            key={item.to}
            to={linkFor(item.to, search, filterKey)}
            aria-current={active ? 'page' : undefined}
            className={`flex flex-1 flex-col items-center justify-center gap-1 ${
              active ? 'glow-text text-accent' : 'text-text-muted hover:text-text'
            }`}
          >
            <svg
              width="22"
              height="22"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.8"
              strokeLinecap="round"
              strokeLinejoin="round"
              aria-hidden="true"
            >
              <path d={item.icon} />
            </svg>
            <span className="font-label whitespace-nowrap text-[11px] font-medium min-[360px]:text-xs">{item.label}</span>
          </NavLink>
        );
      })}
    </nav>
  );
}
