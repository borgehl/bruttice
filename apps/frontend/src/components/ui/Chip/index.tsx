interface ChipProps {
  label: string;
  selected: boolean;
  onToggle: () => void;
}

// A 40px toggle for tags, ages and regions. Selected is a soft tint of the accent, not a solid fill.
export default function Chip({ label, selected, onToggle }: ChipProps) {
  return (
    <button
      type="button"
      aria-pressed={selected}
      onClick={onToggle}
      className={`min-h-10 cursor-pointer rounded-control border px-3.5 text-sm font-medium ${
        selected
          ? 'border-accent bg-accent/15 text-accent'
          : 'border-border text-text-muted hover:border-text-faint hover:text-text'
      }`}
    >
      {label}
    </button>
  );
}
