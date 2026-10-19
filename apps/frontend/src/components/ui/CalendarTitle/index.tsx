import type { MouseEvent } from 'react';
import { shouldSkipLinkNotice, useLinkNoticeStore } from '../../../store/linkNotice';

interface CalendarTitleProps {
  title: string;
  link: string;
}

export default function CalendarTitle({ title, link }: CalendarTitleProps) {
  const ask = useLinkNoticeStore((state) => state.ask);

  // The first tap shows the notice. Modified clicks (new tab, new window) go straight through.
  function onClick(event: MouseEvent<HTMLAnchorElement>) {
    if (event.metaKey || event.ctrlKey || event.shiftKey || event.altKey || event.button !== 0) return;
    if (shouldSkipLinkNotice()) return;
    event.preventDefault();
    ask(link);
  }

  return (
    <h3 className="text-base font-semibold leading-snug text-pretty">
      <a href={link} target="_blank" rel="noopener noreferrer" onClick={onClick} className="text-text hover:text-accent">
        {title}
      </a>
    </h3>
  );
}
