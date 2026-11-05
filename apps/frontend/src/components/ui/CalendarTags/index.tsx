import { getTagColor } from '../TagColor';
import { displayTag, tagsToShow } from '../../../utils/tag';

interface CalendarTagsProps {
  tags: string[];
  // when the event was found to be cancelled, if it was
  cancelledAt?: string;
}

export default function CalendarTags({ tags, cancelledAt }: CalendarTagsProps) {
  const shown = tagsToShow(tags, cancelledAt);
  if (shown.length === 0) return null;

  return (
    <ul className="mt-1.5 flex flex-wrap gap-1.5">
      {shown.map((tag) => (
        <li
          key={tag}
          className={`flex h-[22px] items-center rounded-[4px] px-2 text-xs font-medium ${getTagColor(tag)}`}
        >
          {displayTag(tag)}
        </li>
      ))}
    </ul>
  );
}
