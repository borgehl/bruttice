import { useAfterDelay } from '../../../hooks/useAfterDelay';
import Spinner from '../Spinner';

interface LoadingBlockProps {
  label: string;
}

// A centered spinner for a view with no list shape. Mount it only while loading: it appears
// after a short delay, so a quick response shows nothing.
export default function LoadingBlock({ label }: LoadingBlockProps) {
  const show = useAfterDelay();
  if (!show) return null;

  return (
    <div role="status" className="py-12 text-center">
      <Spinner />
      <p className="mt-4 text-text-muted">{label}</p>
    </div>
  );
}
