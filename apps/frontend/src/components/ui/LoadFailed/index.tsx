import type { LoadError } from '../../../utils/loadError';
import StatusScreen from '../StatusScreen';
import { primaryButtonClass } from '../Field';

interface LoadFailedProps {
  error: LoadError;
  onRetry: () => void;
}

// A view that could not be loaded at all: what happened, and a way to try again.
export default function LoadFailed({ error, onRetry }: LoadFailedProps) {
  return (
    <StatusScreen
      kind="error"
      title={error.title}
      action={
        <button type="button" onClick={onRetry} className={primaryButtonClass}>
          Try again
        </button>
      }
    >
      {error.message}
    </StatusScreen>
  );
}
