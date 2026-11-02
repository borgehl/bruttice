// The address of a feed, for all events or for a saved filter.
export function feedUrl(origin: string, feedType: string, filterKey: string | null): string {
  const url = `${origin}/api/feeds?feedType=${feedType}`;
  return filterKey ? `${url}&filterKey=${encodeURIComponent(filterKey)}` : url;
}
