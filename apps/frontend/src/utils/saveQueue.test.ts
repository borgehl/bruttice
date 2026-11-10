import { describe, expect, it, vi } from 'vitest';
import { createSaveQueue } from './saveQueue';

// a task that finishes when its resolve is called, and says when it started
function controlled() {
  let resolve!: () => void;
  let reject!: (error: Error) => void;
  const started = vi.fn();
  const task = () =>
    new Promise<string>((res, rej) => {
      started();
      resolve = () => res('done');
      reject = rej;
    });
  return { task, started, resolve: () => resolve(), reject: (e: Error) => reject(e) };
}

// lets queued promise callbacks run
const flush = () => new Promise((resolve) => setTimeout(resolve, 0));

describe('createSaveQueue', () => {
  it('runs one task at a time, in the order they were queued', async () => {
    const queue = createSaveQueue();
    const first = controlled();
    const second = controlled();

    const a = queue.run(first.task);
    const b = queue.run(second.task);
    await flush();
    expect(first.started).toHaveBeenCalledTimes(1);
    expect(second.started).not.toHaveBeenCalled();

    first.resolve();
    await a;
    await flush();
    expect(second.started).toHaveBeenCalledTimes(1);

    second.resolve();
    await b;
  });

  it('gives back what the task gives back', async () => {
    const queue = createSaveQueue();
    await expect(queue.run(() => Promise.resolve(42))).resolves.toBe(42);
  });

  it('does not let a failed task stop the ones behind it', async () => {
    const queue = createSaveQueue();
    const failing = controlled();
    const next = vi.fn(() => Promise.resolve('ok'));

    const a = queue.run(failing.task);
    const b = queue.run(next);
    await flush();
    failing.reject(new Error('nope'));

    await expect(a).rejects.toThrow('nope');
    await expect(b).resolves.toBe('ok');
    expect(next).toHaveBeenCalledTimes(1);
  });

  it('says how many saves are waiting or running', async () => {
    const counts: number[] = [];
    const queue = createSaveQueue({ onPendingChange: (n) => counts.push(n) });
    const first = controlled();
    const second = controlled();

    const a = queue.run(first.task);
    const b = queue.run(second.task);
    await flush();
    first.resolve();
    await a;
    await flush();
    second.resolve();
    await b;

    expect(counts).toEqual([1, 2, 1, 0]);
  });

  it('is idle once when the queue empties, not between tasks', async () => {
    const onIdle = vi.fn();
    const queue = createSaveQueue({ onIdle });
    const first = controlled();
    const second = controlled();

    const a = queue.run(first.task);
    const b = queue.run(second.task);
    await flush();
    first.resolve();
    await a;
    await flush();
    expect(onIdle).not.toHaveBeenCalled();

    second.resolve();
    await b;
    expect(onIdle).toHaveBeenCalledTimes(1);
  });

  it('is idle again after a failed task, and for each time it empties', async () => {
    const onIdle = vi.fn();
    const queue = createSaveQueue({ onIdle });

    await expect(queue.run(() => Promise.reject(new Error('nope')))).rejects.toThrow('nope');
    expect(onIdle).toHaveBeenCalledTimes(1);

    await queue.run(() => Promise.resolve());
    expect(onIdle).toHaveBeenCalledTimes(2);
  });

  it('is up to date by the time the caller hears back', async () => {
    let pendingWhenDone = -1;
    const queue = createSaveQueue({ onPendingChange: (n) => (pendingWhenDone = n) });
    await queue.run(() => Promise.resolve());
    expect(pendingWhenDone).toBe(0);
  });
});
