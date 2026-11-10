export interface SaveQueueHandlers {
  // called with how many saves are waiting or running, every time that changes
  onPendingChange?: (pending: number) => void;
  // called once each time the queue empties
  onIdle?: () => void;
}

export interface SaveQueue {
  // Runs task after every task queued before it has finished, whether those worked or not, and
  // gives back what task gives back. A task that fails does not stop the ones behind it.
  run: <T>(task: () => Promise<T>) => Promise<T>;
}

// Saves of one thing, such as the fields of one event, go to the server one at a time and in the
// order they were made. Sent together they can finish in any order, and the last to finish is
// not always the last one made.
export function createSaveQueue({ onPendingChange, onIdle }: SaveQueueHandlers = {}): SaveQueue {
  // never rejects, so one failed task does not fail the chain
  let tail: Promise<unknown> = Promise.resolve();
  let pending = 0;

  const finish = () => {
    pending -= 1;
    onPendingChange?.(pending);
    if (pending === 0) onIdle?.();
  };

  return {
    run<T>(task: () => Promise<T>): Promise<T> {
      pending += 1;
      onPendingChange?.(pending);

      const result = tail.then(task);
      tail = result.catch(() => undefined);
      // registered before the caller's own handlers, so the queue is up to date when they run
      result.then(finish, finish);
      return result;
    },
  };
}
