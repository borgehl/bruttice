import { create } from 'zustand';
import { loadNeonFonts } from '../styles/fonts';

export type Theme = 'info' | 'neon';

export const THEME_KEY = 'theme';

const META_COLOR: Record<Theme, string> = { info: '#12161c', neon: '#0b0714' };

// Anything that isn't a known theme is Info.
export function parseTheme(value: string | null | undefined): Theme {
  return value === 'neon' ? 'neon' : 'info';
}

// Storage can be missing or throw (private windows, blocked site data), so every access is guarded.
export function readTheme(storage: Pick<Storage, 'getItem'> | undefined = safeStorage()): Theme {
  try {
    return parseTheme(storage?.getItem(THEME_KEY));
  } catch {
    return 'info';
  }
}

export function writeTheme(theme: Theme, storage: Pick<Storage, 'setItem'> | undefined = safeStorage()): void {
  try {
    storage?.setItem(THEME_KEY, theme);
  } catch {
    // The choice still applies for this visit.
  }
}

function safeStorage(): Storage | undefined {
  try {
    return globalThis.localStorage;
  } catch {
    return undefined;
  }
}

function applyTheme(theme: Theme): void {
  document.documentElement.setAttribute('data-theme', theme);
  document.querySelector('meta[name="theme-color"]')?.setAttribute('content', META_COLOR[theme]);
  if (theme === 'neon') void loadNeonFonts();
}

interface ThemeState {
  theme: Theme;
  setTheme: (theme: Theme) => void;
}

// index.html has already set data-theme from storage before this runs.
export const useThemeStore = create<ThemeState>((set) => ({
  theme: typeof document === 'undefined' ? 'info' : parseTheme(document.documentElement.getAttribute('data-theme')),
  setTheme: (theme) => {
    applyTheme(theme);
    writeTheme(theme);
    set({ theme });
  },
}));

export function initTheme(): void {
  applyTheme(useThemeStore.getState().theme);
}
