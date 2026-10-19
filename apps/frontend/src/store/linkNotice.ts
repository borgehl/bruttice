import { create } from 'zustand';

export const SKIP_LINK_NOTICE_KEY = 'skipLinkNotice';

// Storage can be missing or throw, so access is guarded. Without it the notice is just shown every time.
function safeStorage(): Storage | undefined {
  try {
    return globalThis.localStorage;
  } catch {
    return undefined;
  }
}

export function shouldSkipLinkNotice(storage: Pick<Storage, 'getItem'> | undefined = safeStorage()): boolean {
  try {
    return storage?.getItem(SKIP_LINK_NOTICE_KEY) === '1';
  } catch {
    return false;
  }
}

export function saveSkipLinkNotice(storage: Pick<Storage, 'setItem'> | undefined = safeStorage()): void {
  try {
    storage?.setItem(SKIP_LINK_NOTICE_KEY, '1');
  } catch {
    // The choice still applies until the page is reloaded.
  }
}

interface LinkNoticeState {
  // the link waiting for confirmation, null when the sheet is closed
  pendingUrl: string | null;
  ask: (url: string) => void;
  close: () => void;
}

export const useLinkNoticeStore = create<LinkNoticeState>((set) => ({
  pendingUrl: null,
  ask: (url) => set({ pendingUrl: url }),
  close: () => set({ pendingUrl: null }),
}));
