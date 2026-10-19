// Class names are written out in full so Tailwind can see them. Each sets --tag, which `tag` uses.
const tagColors: Record<string, string> = {
  music: 'tag-music',
  comedy: 'tag-comedy',
  community: 'tag-community',
  literature: 'tag-literature',
  diy: 'tag-diy',
  free: 'tag-free',
  recurring: 'tag-recurring',
  art: 'tag-art',
  sports: 'tag-sports',
  film: 'tag-film',
};

export function getTagColor(tag: string): string {
  return `tag ${tagColors[tag.toLowerCase()] ?? 'tag-recurring'}`;
}
