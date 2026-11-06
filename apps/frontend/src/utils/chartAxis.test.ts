import { describe, it, expect } from 'vitest'
import { axisTicks, labelIndexes, shortDate } from './chartAxis'

describe('axisTicks', () => {
  it('has four marks, from zero, evenly spaced', () => {
    for (const max of [0, 1, 2, 7, 10, 33, 100, 999, 12345]) {
      const ticks = axisTicks(max)
      expect(ticks).toHaveLength(4)
      expect(ticks[0]).toBe(0)
      expect(ticks[2] - ticks[1]).toBe(ticks[1] - ticks[0])
      expect(ticks[3] - ticks[2]).toBe(ticks[1] - ticks[0])
    }
  })

  it('reaches the highest value without going far past it', () => {
    expect(axisTicks(7)).toEqual([0, 3, 6, 9])
    expect(axisTicks(9)).toEqual([0, 3, 6, 9])
    expect(axisTicks(10)).toEqual([0, 4, 8, 12])
    expect(axisTicks(100)).toEqual([0, 40, 80, 120])
  })

  it('never has a mark below the highest value', () => {
    for (let max = 0; max < 2000; max++) {
      const ticks = axisTicks(max)
      expect(ticks[3]).toBeGreaterThanOrEqual(max)
    }
  })

  it('is whole numbers, and a small axis for no visitors', () => {
    expect(axisTicks(0)).toEqual([0, 1, 2, 3])
    for (const tick of axisTicks(47)) expect(Number.isInteger(tick)).toBe(true)
  })
})

describe('labelIndexes', () => {
  it('always names the newest bar', () => {
    for (const count of [1, 2, 7, 30, 90, 365]) {
      const indexes = labelIndexes(count)
      expect(indexes[indexes.length - 1]).toBe(count - 1)
    }
  })

  it('always names the first bar', () => {
    for (let count = 1; count <= 400; count++) {
      expect(labelIndexes(count)[0]).toBe(0)
    }
  })

  it('spreads about six labels over a long range', () => {
    expect(labelIndexes(30)).toEqual([0, 5, 11, 17, 23, 29])
    expect(labelIndexes(90)).toEqual([0, 17, 35, 53, 71, 89])
  })

  it('lets the label next to the first bar give way when it would crowd it', () => {
    // counting back from 31 in steps of 7 ends at 3, three bars from the first
    expect(labelIndexes(32)).toEqual([0, 10, 17, 24, 31])
  })

  it('names every other bar for a week, never more than the target plus one for a year', () => {
    expect(labelIndexes(7)).toEqual([0, 2, 4, 6])
    expect(labelIndexes(365).length).toBeLessThanOrEqual(7)
  })

  it('is in order, with no bar twice', () => {
    for (let count = 1; count <= 400; count++) {
      const indexes = labelIndexes(count)
      expect(new Set(indexes).size).toBe(indexes.length)
      expect([...indexes].sort((a, b) => a - b)).toEqual(indexes)
    }
  })

  it('has nothing to name without bars', () => {
    expect(labelIndexes(0)).toEqual([])
  })
})

describe('shortDate', () => {
  it('reads month and day from the text', () => {
    expect(shortDate('2026-11-06')).toBe('Nov 6')
    expect(shortDate('2026-01-31')).toBe('Jan 31')
  })

  it('leaves what is not a date as it is', () => {
    expect(shortDate('soon')).toBe('soon')
  })
})
