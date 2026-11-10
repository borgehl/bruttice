import { describe, expect, it } from 'vitest';
import type { Sponsor } from '../types/sponsor';
import { imageProblem, MAX_SPONSOR_IMAGE_BYTES, pickSponsor, sponsorAnchor, type EventBox } from './sponsor';

const sponsor = (id: number): Sponsor => ({ id, name: `s${id}`, text: '', imageUrl: '', clickUrl: `/api/sponsors/${id}/click` });

describe('pickSponsor', () => {
  it('is null when no card is running', () => {
    expect(pickSponsor([])).toBeNull();
  });

  it('is the only sponsor', () => {
    expect(pickSponsor([sponsor(1)], () => 0.99)?.id).toBe(1);
  });

  it('spreads over several sponsors', () => {
    const all = [sponsor(1), sponsor(2), sponsor(3)];
    expect(pickSponsor(all, () => 0)?.id).toBe(1);
    expect(pickSponsor(all, () => 0.5)?.id).toBe(2);
    expect(pickSponsor(all, () => 0.999)?.id).toBe(3);
  });

  it('stays in range if the random number is exactly 1', () => {
    expect(pickSponsor([sponsor(1), sponsor(2)], () => 1)?.id).toBe(2);
  });
});

describe('sponsorAnchor', () => {
  // a scroll box 500 tall with rows 100 tall from the top, the card 118 tall
  const viewport = { top: 0, bottom: 500 };
  const rows = (count: number, height = 100, start = 0): EventBox[] =>
    Array.from({ length: count }, (_, i) => ({ guid: `e${i}`, top: start + i * height, bottom: start + (i + 1) * height }));

  it('is null when there are no rows', () => {
    expect(sponsorAnchor([], viewport, 118)).toBeNull();
  });

  it('is the last row that is fully in view with room for the card', () => {
    // rows end at 100..800: e0..e3 end by 400, and 400 + 118 = 518 is past the bottom, so e2 (300 + 118)
    expect(sponsorAnchor(rows(8), viewport, 118)).toBe('e2');
  });

  it('is the last row when the whole list fits with room to spare', () => {
    expect(sponsorAnchor(rows(2), viewport, 118)).toBe('e1');
  });

  it('skips a row that is cut off at the bottom', () => {
    // the sixth row runs from 500 to 600
    expect(sponsorAnchor(rows(8), viewport, 0)).toBe('e4');
  });

  it('falls back to the last fully visible row when none leaves room for the card', () => {
    // rows 240 tall: e0 ends at 240 (+118 fits), e1 ends at 480 (+118 does not), e2 is cut off
    expect(sponsorAnchor(rows(3, 240), viewport, 118)).toBe('e0');
    // a card taller than the box leaves no row with room, so the last full row is used
    expect(sponsorAnchor(rows(8), viewport, 600)).toBe('e4');
  });

  it('falls back to the first row when no row is fully in view', () => {
    expect(sponsorAnchor(rows(3, 600), viewport, 118)).toBe('e0');
  });

  it('ignores a row that starts above the scroll box', () => {
    expect(sponsorAnchor(rows(3, 100, -50), viewport, 118)).toBe('e2');
  });
});

describe('imageProblem', () => {
  it('takes the types the server takes', () => {
    for (const type of ['image/png', 'image/jpeg', 'image/gif', 'image/webp']) {
      expect(imageProblem({ size: 1000, type })).toBe('');
    }
  });

  it('refuses SVG and other files', () => {
    expect(imageProblem({ size: 1000, type: 'image/svg+xml' })).not.toBe('');
    expect(imageProblem({ size: 1000, type: 'text/html' })).not.toBe('');
  });

  it('refuses a picture over the limit', () => {
    expect(imageProblem({ size: MAX_SPONSOR_IMAGE_BYTES, type: 'image/png' })).toBe('');
    expect(imageProblem({ size: MAX_SPONSOR_IMAGE_BYTES + 1, type: 'image/png' })).toContain('too big');
  });
});
