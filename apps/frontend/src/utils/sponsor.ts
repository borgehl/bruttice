import type { Sponsor } from '../types/sponsor';

// The largest picture the server takes, and the types it takes.
export const MAX_SPONSOR_IMAGE_BYTES = 200 * 1024;
const IMAGE_TYPES = ['image/png', 'image/jpeg', 'image/gif', 'image/webp'];

// One sponsor for a page load, at random so that several sponsors share the visits. null if no
// card is running.
export function pickSponsor(sponsors: Sponsor[], random: () => number = Math.random): Sponsor | null {
  if (sponsors.length === 0) return null;
  return sponsors[Math.min(sponsors.length - 1, Math.floor(random() * sponsors.length))];
}

// How tall the card is with a picture: the picture box and the row's padding and border. Room is kept
// for it when choosing where it goes, so the whole card is on screen and not just its top.
export const SPONSOR_CARD_HEIGHT = 118;

// Where an event row is on screen, in the same units as the scroll box.
export interface EventBox {
  guid: string;
  top: number;
  bottom: number;
}

// The event the card follows: the last one of the rows, in the order they are on the page, that is
// fully in the scroll box with room left under it for the card. Failing that, the last one that is
// fully in view, and failing that the first. null if there are no rows.
export function sponsorAnchor(boxes: EventBox[], viewport: { top: number; bottom: number }, cardHeight: number): string | null {
  let fitsWithCard: string | null = null;
  let fullyVisible: string | null = null;
  for (const box of boxes) {
    if (box.top < viewport.top) continue;
    if (box.bottom <= viewport.bottom) fullyVisible = box.guid;
    if (box.bottom + cardHeight <= viewport.bottom) fitsWithCard = box.guid;
  }
  return fitsWithCard ?? fullyVisible ?? boxes[0]?.guid ?? null;
}

// Why a picture can't be used, or '' if it can.
export function imageProblem(file: { size: number; type: string }): string {
  if (!IMAGE_TYPES.includes(file.type)) return 'The picture has to be a PNG, JPEG, GIF or WebP file.';
  if (file.size > MAX_SPONSOR_IMAGE_BYTES) return `The picture is too big (at most ${MAX_SPONSOR_IMAGE_BYTES / 1024} KB).`;
  return '';
}

// The file's bytes as base64, which is how a picture is sent.
export function readImage(file: Blob): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => {
      const result = String(reader.result);
      resolve(result.slice(result.indexOf(',') + 1));
    };
    reader.onerror = () => reject(reader.error ?? new Error('The picture could not be read'));
    reader.readAsDataURL(file);
  });
}
