export interface PrependScrollInput {
  // this is the first time earlier events were put above the ones the page started with
  initial: boolean;
  // how far the scroll box was scrolled, and how tall its content was, just before the prepend
  scrollTopBefore: number;
  heightBefore: number;
  // how tall the content is now
  heightAfter: number;
  // How far the top of the day section of the event that was first before the prepend is from the top
  // of the scroll box now, or null if that section can't be found.
  sectionOffset: number | null;
}

// Where to scroll to once earlier events have been put above what was on screen.
//
// Normally the view holds still: the content is scrolled down by the height that was added, so the
// rows the visitor was looking at stay where they were.
//
// The first prepend is different. The page opens at the day the visitor asked for, and the earlier
// events load at once, because the "Loading earlier events" strip at the top is in view. Holding the
// view still would keep that strip's height at the top, which is then the tail of the previous day with
// its pinned day header over it: a header with no events. So when nothing has been scrolled yet, the
// view moves to the start of the day that was first, and the earlier days are one scroll away.
export function scrollTopAfterPrepend(input: PrependScrollInput, currentScrollTop: number): number {
  const { initial, scrollTopBefore, heightBefore, heightAfter, sectionOffset } = input;
  if (initial && scrollTopBefore === 0 && sectionOffset !== null) {
    return currentScrollTop + sectionOffset;
  }
  return currentScrollTop + (heightAfter - heightBefore);
}
