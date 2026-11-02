// Tags are stored lower case, people read "Music" and "DIY".
export function displayTag(name: string): string {
  if (name.toLowerCase() === 'diy') return 'DIY';
  return name.charAt(0).toUpperCase() + name.slice(1);
}
