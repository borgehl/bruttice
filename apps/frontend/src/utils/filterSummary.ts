import type { FilterDetails } from '../services/api';
import { displayTag } from './tag';

export function ageLabel(value: number): string {
  return value <= 1 ? 'All ages' : `${value}+`;
}

// What a saved filter shows, in words: "Music, Not Comedy, Max 18+, Not Cain's".
export function describeFilter(filter: FilterDetails): string[] {
  const parts: string[] = [];

  const tags = filter.tags ?? [];
  parts.push(...tags.map((tag) => (filter.tagsInclude ? displayTag(tag) : `Not ${displayTag(tag)}`)));

  if (filter.age !== undefined && filter.age > 0) {
    parts.push(filter.agesLessThan ? `Max ${ageLabel(filter.age)}` : `Min ${ageLabel(filter.age)}`);
  }

  const venues = filter.venues ?? [];
  parts.push(...venues.map((venue) => (filter.venuesInclude ? venue : `Not ${venue}`)));

  return parts;
}
