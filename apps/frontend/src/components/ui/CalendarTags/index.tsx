import { getTagColor } from '../TagColor';

interface CalendarTagsProps {
  tags: string[];
}

export default function CalendarTags({ tags }: CalendarTagsProps) {
  if (tags.length === 0) return null;

  return (
    <div className="flex flex-wrap gap-1">
      {tags.map((tag) => (
        <span
          key={tag}
          className={`px-2 py-0.5 text-xs rounded-full ${getTagColor(tag)}`}
        >
          {tag}
        </span>
      ))}
    </div>
  );
}
