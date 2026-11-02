import { useState } from 'react';
import { useFlyerViewerStore } from '../../../store/flyerViewer';

interface FlyerThumbProps {
  url: string;
  title: string;
}

// A small flyer at the right of an event row. Tapping it shows the flyer full screen. If the image
// can't load nothing is drawn, so a dead link leaves no hole in the row.
export default function FlyerThumb({ url, title }: FlyerThumbProps) {
  const [failed, setFailed] = useState(false);
  const open = useFlyerViewerStore((state) => state.open);

  if (failed) return null;

  return (
    <button
      type="button"
      onClick={() => open(url, title)}
      aria-label={`View flyer for ${title}`}
      className="h-[88px] w-16 shrink-0 cursor-pointer overflow-hidden rounded-control border border-border bg-raised p-0"
    >
      <img
        src={url}
        alt=""
        loading="lazy"
        onError={() => setFailed(true)}
        className="h-full w-full object-cover"
      />
    </button>
  );
}
