import { useEffect, type RefObject } from 'react';

// Closes an open popup on a press anywhere outside of `root`, and on Escape with the focus back on
// `returnFocus`. Nothing is listened to while it is closed.
export function useDismissable(
  open: boolean,
  onClose: () => void,
  root: RefObject<HTMLElement | null>,
  returnFocus: RefObject<HTMLElement | null>,
) {
  useEffect(() => {
    if (!open) return;

    const onPointerDown = (event: PointerEvent) => {
      if (!root.current?.contains(event.target as Node)) onClose();
    };
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key !== 'Escape') return;
      onClose();
      returnFocus.current?.focus();
    };

    document.addEventListener('pointerdown', onPointerDown);
    document.addEventListener('keydown', onKeyDown);
    return () => {
      document.removeEventListener('pointerdown', onPointerDown);
      document.removeEventListener('keydown', onKeyDown);
    };
  }, [open, onClose, root, returnFocus]);
}
