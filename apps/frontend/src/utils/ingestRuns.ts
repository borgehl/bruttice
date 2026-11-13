import type { IngestSource } from '../types/event';

function plural(n: number, word: string): string {
  return `${n} ${word}${n === 1 ? '' : 's'}`;
}

// What the admin page's expandable part of a source says before it is opened, such as
// "2 errors · 1 skipped". Empty when there is nothing to expand: the page shows no expander then.
export function sourceDetailSummary(source: Pick<IngestSource, 'errors' | 'skipped'>): string {
  const parts: string[] = [];
  if (source.errors.length > 0) parts.push(plural(source.errors.length, 'error'));
  if (source.skipped.length > 0) parts.push(`${source.skipped.length} skipped`);
  return parts.join(' · ');
}
