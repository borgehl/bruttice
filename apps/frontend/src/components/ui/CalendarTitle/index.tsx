interface CalendarTitleProps {
  title: string;
  link: string;
}

export default function CalendarTitle({ title, link }: CalendarTitleProps) {
  return (
    <h3 className="text-lg font-medium mb-1">
      <a
        href={link}
        target="_blank"
        rel="noopener noreferrer"
        className="text-text hover:text-primary"
      >
        {title}
      </a>
    </h3>
  );
}
