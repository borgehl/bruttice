import { getTagColor } from '../TagColor';
import { displayTag } from '../../../utils/tag';

interface CalendarTagsProps {
  tags: string[];
}

export default function CalendarTags({ tags }: CalendarTagsProps) {
  if (tags.length === 0) return null;

  return (
    <ul className="mt-1.5 flex flex-wrap gap-1.5">
      {tags.map((tag) => (
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
