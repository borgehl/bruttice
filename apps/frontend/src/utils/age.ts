// "all-ages" is a stored name, people read "All ages".
export function displayAge(name: string): string {
  return name === 'all-ages' ? 'All ages' : name;
}
