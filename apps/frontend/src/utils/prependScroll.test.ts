import { describe, expect, it } from 'vitest';
import { scrollTopAfterPrepend, type PrependScrollInput } from './prependScroll';

const base: PrependScrollInput = {
  initial: true,
  scrollTopBefore: 0,
  heightBefore: 1000,
  // 900 of earlier events were added, and the loading strip at the top went from 44 to 24
  heightAfter: 1880,
  sectionOffset: 924,
};

describe('scrollTopAfterPrepend', () => {
  it('starts the first day at the top on the first prepend, when nothing was scrolled', () => {
    expect(scrollTopAfterPrepend(base, 0)).toBe(924);
  });

  it('adds to the scroll the box has now', () => {
    expect(scrollTopAfterPrepend(base, 10)).toBe(934);
  });

  it('holds the view still when the visitor had already scrolled', () => {
    expect(scrollTopAfterPrepend({ ...base, scrollTopBefore: 300 }, 300)).toBe(300 + 880);
  });

  it('holds the view still on later prepends', () => {
    expect(scrollTopAfterPrepend({ ...base, initial: false }, 0)).toBe(880);
  });

  it('holds the view still when the first day cannot be found', () => {
    expect(scrollTopAfterPrepend({ ...base, sectionOffset: null }, 0)).toBe(880);
  });

  it('does not move when nothing was added', () => {
    expect(scrollTopAfterPrepend({ ...base, initial: false, heightAfter: 1000 }, 120)).toBe(120);
  });
});
