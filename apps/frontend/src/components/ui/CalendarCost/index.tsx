interface CalendarCostProps {
  cost: string;
}

export function formatCost(cost: string): string {
  const num = Number(cost);
  if (!isNaN(num) && cost.trim() !== '') {
    return num === 0 ? 'Free' : `$${num.toFixed(2)}`;
  }
  return cost;
}

export default function CalendarCost({ cost }: CalendarCostProps) {
  if (!cost) return null;

  return <span className="font-data">{formatCost(cost)}</span>;
}
