import { getAgeColor } from '../AgeColor';
import { displayAge } from '../../../utils/age';

interface CalendarAgeProps {
  age: string;
}

export default function CalendarAge({ age }: CalendarAgeProps) {
  return (
    <span className={`font-data ${getAgeColor(age)}`}>
      {displayAge(age)}
    </span>
  );
}
