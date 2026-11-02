interface SpinnerProps {
  size?: 'sm' | 'md';
}

export default function Spinner({ size = 'md' }: SpinnerProps) {
  const dimensions = size === 'sm' ? 'h-4 w-4 border-2' : 'h-8 w-8 border-4';

  return (
    <span
      className={`inline-block animate-spin rounded-full border-accent border-t-transparent motion-reduce:animate-pulse ${dimensions}`}
      aria-hidden="true"
    />
  );
}
