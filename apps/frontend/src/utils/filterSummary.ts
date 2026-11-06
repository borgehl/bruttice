import type { FilterDetails } from '../services/api';
import { displayTag } from './tag';

export function ageLabel(value: number): string {
  return value <= 1 ? 'All ages' : `${value}+`;
}

export interface FilterGroup {
  title: string;
  parts: string[];
}

// What a saved filter shows, in words and by kind: Tags "Music, Not Comedy", Age "Max 18+",
// Regions "okc", Venues "Not Cain's". A kind the filter doesn't use is left out.
export function describeFilterGroups(filter: FilterDetails): FilterGroup[] {
  const groups: FilterGroup[] = [];

  const tags = filter.tags ?? [];
  groups.push({
    title: 'Tags',
    parts: tags.map((tag) => (filter.tagsInclude ? displayTag(tag) : `Not ${displayTag(tag)}`)),
  });

  groups.push({
    title: 'Age',
    parts:
      filter.age !== undefined && filter.age > 0
        ? [filter.agesLessThan ? `Max ${ageLabel(filter.age)}` : `Min ${ageLabel(filter.age)}`]
        : [],
  });

  groups.push({ title: 'Regions', parts: filter.regions ?? [] });

  const venues = filter.venues ?? [];
  groups.push({ title: 'Venues', parts: venues.map((venue) => (filter.venuesInclude ? venue : `Not ${venue}`)) });

  return groups.filter((group) => group.parts.length > 0);
}

// The same in one list: "Music, Not Comedy, Max 18+, okc, Not Cain's".
export function describeFilter(filter: FilterDetails): string[] {
  return describeFilterGroups(filter).flatMap((group) => group.parts);
}
