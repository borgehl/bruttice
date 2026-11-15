import { describe, expect, it } from 'vitest';
import { sourceDetailSummary } from './ingestRuns';

const skippedEvent = 'event 3 ("Twen"): not saved: an event with this guid was deleted';

describe('sourceDetailSummary', () => {
  it('is empty when there is nothing to expand', () => {
    expect(sourceDetailSummary({ errors: [], skipped: [] })).toBe('');
  });

  it('counts skipped events on their own, not as errors', () => {
    expect(sourceDetailSummary({ errors: [], skipped: [skippedEvent] })).toBe('1 skipped');
    expect(sourceDetailSummary({ errors: [], skipped: [skippedEvent, skippedEvent] })).toBe('2 skipped');
  });

  it('counts errors, with the plural', () => {
    expect(sourceDetailSummary({ errors: ['boom'], skipped: [] })).toBe('1 error');
    expect(sourceDetailSummary({ errors: ['a', 'b'], skipped: [] })).toBe('2 errors');
  });

  it('names both when a source has both, errors first', () => {
    expect(sourceDetailSummary({ errors: ['a', 'b'], skipped: [skippedEvent] })).toBe('2 errors · 1 skipped');
  });
});
