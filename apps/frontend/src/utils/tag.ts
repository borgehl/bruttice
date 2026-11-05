// Tags are stored lower case, people read "Music" and "DIY".
export function displayTag(name: string): string {
  if (name.toLowerCase() === 'diy') return 'DIY';
  return name.charAt(0).toUpperCase() + name.slice(1);
}

// The tags to show on an event: a cancelled event leads with a "cancelled" mark.
export const CANCELLED_TAG = 'cancelled';

export function tagsToShow(tags: string[], cancelledAt?: string): string[] {
  const rest = tags.filter((tag) => tag.toLowerCase() !== CANCELLED_TAG);
  return cancelledAt ? [CANCELLED_TAG, ...rest] : rest;
}
