// The calendar's region picker keeps its choice in the address, as one `regions` parameter per
// region (?regions=tulsa&regions=okc), so a link carries it. No parameter means every region.
export const REGIONS_PARAM = 'regions';

export const ALL_REGIONS_LABEL = 'All Regions';

// The regions in the address: names without surrounding space, none empty, none twice.
export function regionsFromParams(params: URLSearchParams): string[] {
  const regions: string[] = [];
  for (const value of params.getAll(REGIONS_PARAM)) {
    const name = value.trim();
    if (name !== '' && !regions.includes(name)) regions.push(name);
  }
  return regions;
}

// A copy of the address parameters with the regions replaced. An empty list is every region, so it
// leaves no parameter at all.
export function withRegions(params: URLSearchParams, regions: string[]): URLSearchParams {
  const next = new URLSearchParams(params);
  next.delete(REGIONS_PARAM);
  for (const region of regions) next.append(REGIONS_PARAM, region);
  return next;
}

// What the picker's button says: "All Regions", the region's name, or how many.
//
// Picking every region is the same as picking none, so it reads "All Regions" too. A name in the
// address that the site doesn't have is not counted: it matches nothing, and it would only make
// the button promise a region that is not there.
export function regionLabel(selected: string[], available: string[]): string {
  const known = selected.filter((name) => available.includes(name));
  if (known.length === 0 || known.length === available.length) return ALL_REGIONS_LABEL;
  if (known.length === 1) return known[0];
  return `${known.length} regions`;
}

// A key that is the same for the same regions, whatever order they were picked in. Mount the
// calendar with it so a change starts a fresh scroll.
export function regionsKey(regions: string[]): string {
  return [...regions].sort().join('\u001f');
}
