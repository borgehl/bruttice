// The message an inline embed posts to tell the page how tall it is:
//   parent.postMessage({ type: 'bruttice-embed-height', height: 600 }, '*')
export const EMBED_HEIGHT_MESSAGE = 'bruttice-embed-height';

export const DEFAULT_EMBED_HEIGHT = 420;
const MIN_EMBED_HEIGHT = 120;
const MAX_EMBED_HEIGHT = 2400;

// The height a message asks for, kept within sane bounds, or null if it is not an embed height message.
export function embedHeightFrom(data: unknown): number | null {
  if (typeof data !== 'object' || data === null) return null;
  const { type, height } = data as { type?: unknown; height?: unknown };
  if (type !== EMBED_HEIGHT_MESSAGE || typeof height !== 'number' || !Number.isFinite(height)) return null;
  return Math.min(MAX_EMBED_HEIGHT, Math.max(MIN_EMBED_HEIGHT, Math.ceil(height)));
}
