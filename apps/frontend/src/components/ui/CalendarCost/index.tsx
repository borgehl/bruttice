interface CalendarCostProps {
  cost: string;
}

function formatCost(cost: string): string {
  const num = Number(cost);
  if (!isNaN(num) && cost.trim() !== '') {
    return `$${num.toFixed(2)}`;
  }
  return cost;
}

export default function CalendarCost({ cost }: CalendarCostProps) {
  if (!cost) return null;

  return (
    <span className="text-text-muted">{formatCost(cost)}</span>
  );
}
