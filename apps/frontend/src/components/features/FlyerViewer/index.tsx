import { useFlyerViewerStore } from '../../../store/flyerViewer';
import { useDialog } from '../../../hooks/useDialog';

// A flyer shown full screen. A tap anywhere, or Escape, closes it.
export default function FlyerViewer() {
  const flyer = useFlyerViewerStore((state) => state.flyer);
  if (!flyer) return null;
  return <Viewer url={flyer.url} title={flyer.title} />;
}

function Viewer({ url, title }: { url: string; title: string }) {
  const close = useFlyerViewerStore((state) => state.close);
  const ref = useDialog<HTMLDivElement>(close);

  return (
    <div
      ref={ref}
      role="dialog"
      aria-modal="true"
      aria-label={`Flyer for ${title}`}
      onClick={close}
      className="fixed inset-0 z-50 flex items-center justify-center bg-scrim-strong p-4"
    >
      <button
        type="button"
        data-autofocus
        onClick={close}
        aria-label="Close flyer"
        className="absolute right-3 top-3 flex h-11 w-11 cursor-pointer items-center justify-center rounded-control border border-border bg-surface text-text"
      >
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true">
          <path d="M6 6l12 12M18 6L6 18" />
        </svg>
      </button>
      <img src={url} alt={`Flyer for ${title}`} className="max-h-full max-w-full object-contain" />
    </div>
  );
}
