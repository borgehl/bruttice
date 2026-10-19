interface EmbedFrameProps {
  url: string;
  title: string;
}

// A page of another site, shown under a page's text, with a link to open it on its own.
export default function EmbedFrame({ url, title }: EmbedFrameProps) {
  return (
    <section className="pt-2">
      <iframe
        src={url}
        title={title}
        loading="lazy"
        sandbox="allow-scripts allow-same-origin allow-popups"
        className="h-[420px] w-full rounded-control border border-border bg-surface"
      />
      <a href={url} target="_blank" rel="noopener noreferrer" className="mt-2 inline-block text-sm">
        Open in a new tab
      </a>
    </section>
  );
}
