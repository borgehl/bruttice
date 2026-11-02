export function getAgeColor(age: string): string {
  if (age === 'all-ages') return 'text-age-all';
  if (age === '21+') return 'text-age-21';
  if (age === '18+') return 'text-age-18';
  if (age === '16+') return 'text-age-16';
  return 'text-text-faint';
}
