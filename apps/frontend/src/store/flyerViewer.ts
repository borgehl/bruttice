import { create } from 'zustand';

interface FlyerViewerState {
  flyer: { url: string; title: string } | null;
  open: (url: string, title: string) => void;
  close: () => void;
}

export const useFlyerViewerStore = create<FlyerViewerState>((set) => ({
  flyer: null,
  open: (url, title) => set({ flyer: { url, title } }),
  close: () => set({ flyer: null }),
}));
