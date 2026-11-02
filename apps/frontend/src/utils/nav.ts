// Is the nav item for `to` the current page? `also` lists other paths that belong to it.
export function isActive(pathname: string, to: string, also: string[] = []): boolean {
  if (to === '/') return pathname === '/' || also.includes(pathname);
  return [to, ...also].some((path) => pathname === path || pathname.startsWith(`${path}/`));
}

// The calendar keeps its date and filter when you come back to it, and Feeds shows the feeds of the
// filter you are looking at.
export function linkFor(to: string, search: string, filterKey: string | null) {
  if (to === '/') return { pathname: '/', search };
  if (to === '/feeds' && filterKey) return { pathname: '/feeds', search: `?filterKey=${encodeURIComponent(filterKey)}` };
  return to;
}
