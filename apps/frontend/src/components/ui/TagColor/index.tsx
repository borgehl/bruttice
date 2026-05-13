const tagColors: Record<string, string> = {
  music: 'bg-purple-900/50 text-purple-300',
  comedy: 'bg-yellow-900/50 text-yellow-300',
  community: 'bg-green-900/50 text-green-300',
  literature: 'bg-sky-900/50 text-sky-300',
  diy: 'bg-orange-900/50 text-orange-300',
  free: 'bg-emerald-900/50 text-emerald-300',
  recurring: 'bg-neutral-700/50 text-neutral-300',
  art: 'bg-pink-900/50 text-pink-300',
  sports: 'bg-red-900/50 text-red-300',
  film: 'bg-indigo-900/50 text-indigo-300',
};

export function getTagColor(tag: string): string {
  return tagColors[tag.toLowerCase()] || 'bg-neutral-700/50 text-neutral-300';
}
