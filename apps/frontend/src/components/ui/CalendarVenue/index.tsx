interface CalendarVenueProps {
  name: string;
  link: string;
}

export default function CalendarVenue({ name, link }: CalendarVenueProps) {
  return (
    <a
      href={link}
      target="_blank"
      rel="noopener noreferrer"
      className="text-sm text-text-muted hover:text-accent"
    >
      {name}
    </a>
  );
}
