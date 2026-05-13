import { getAgeColor } from '../AgeColor';

interface CalendarAgeProps {
  age: string;
}

export default function CalendarAge({ age }: CalendarAgeProps) {
  return (
    <span className={`font-medium ${getAgeColor(age)}`}>
      {age}
    </span>
  );
}
