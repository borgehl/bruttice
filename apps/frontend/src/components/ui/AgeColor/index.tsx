export function getAgeColor(age: string): string {
  if (age === 'all-ages') return 'text-green-400';
  if (age === '21+') return 'text-red-400';
  if (age === '18+') return 'text-orange-400';
  if (age === '16+') return 'text-yellow-400';
  return 'text-text-dim';
}
