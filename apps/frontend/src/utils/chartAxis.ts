// Steps a count axis can take, times a power of ten: whole numbers that read cleanly.
const STEPS = [1, 2, 3, 4, 5, 6, 8, 10];

// The marks of a vertical count axis: four of them, evenly spaced from 0, whole numbers, the top one
// at or above `max`. [0, 3, 6, 9] for a highest value of 7. The bars are scaled to the last one.
export function axisTicks(max: number): number[] {
  const need = Math.max(1, max) / 3;

  let scale = 1;
  while (scale * STEPS[STEPS.length - 1] < need) scale *= 10;
  const step = STEPS.map((s) => s * scale).find((s) => s >= need) ?? scale * 10;

  return [0, step, step * 2, step * 3];
}

// The bars that get a date: the first and the last always, and between them every `step`th one
// counting back from the last, so the labels are spread evenly. About `target` labels, fewer for
// few bars. Counting back can leave a short gap before the first bar, then the label nearest to it
// gives way to the first bar's, so two never crowd.
export function labelIndexes(count: number, target = 6): number[] {
  if (count <= 0) return [];
  const step = Math.max(1, Math.ceil((count - 1) / (target - 1)));

  const indexes: number[] = [];
  for (let i = count - 1; i >= 0; i -= step) indexes.unshift(i);

  if (indexes[0] > 0) {
    if (indexes[0] < step / 2) indexes[0] = 0;
    else indexes.unshift(0);
  }
  return indexes;
}

const MONTHS = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];

// "2026-11-06" as "Nov 6". Read from the text, not through a Date: a date has no time zone to shift.
export function shortDate(date: string): string {
  const [, month, day] = date.split('-').map(Number);
  if (!month || !day || month > 12) return date;
  return `${MONTHS[month - 1]} ${day}`;
}
