// A failed request, with the status the server answered with.
export class ApiError extends Error {
  readonly status: number;

  constructor(status: number, message: string) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
  }
}

export interface LoadError {
  title: string;
  message: string;
}

// What to tell a visitor when something could not be loaded. Never the server's own text.
export function describeLoadError(err: unknown): LoadError {
  // fetch rejects with a TypeError when there is no connection at all
  if (err instanceof TypeError) {
    return { title: "Can't reach the server", message: 'Check your connection and try again.' };
  }
  if (err instanceof ApiError && err.status >= 500) {
    return { title: 'Something went wrong', message: 'This is on our side. Try again in a moment.' };
  }
  return { title: 'Something went wrong', message: 'It could not be loaded. Try again.' };
}

// For catch blocks: keeps the original error in the console and returns the copy to show.
export function toLoadError(err: unknown): LoadError {
  console.error(err);
  return describeLoadError(err);
}
