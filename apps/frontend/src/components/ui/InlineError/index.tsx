interface InlineErrorProps {
  message: string;
  onRetry?: () => void;
  className?: string;
}

// A failure beside content that is still on screen. Keeps what the visitor has.
export default function InlineError({ message, onRetry, className = '' }: InlineErrorProps) {
  return (
    <div role="alert" className={`rounded-control border border-danger/30 bg-danger-muted p-3 text-center ${className}`}>
      <p className="text-sm text-danger">{message}</p>
      {onRetry && (
        <button type="button" onClick={onRetry} className="mt-1 min-h-11 cursor-pointer text-sm text-accent hover:underline">
          Try again
        </button>
      )}
    </div>
  );
}
