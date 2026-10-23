import { useEffect, useRef, useState } from 'react';
import { DEFAULT_EMBED_HEIGHT, embedHeightFrom } from '../../../utils/embedHeight';

interface EmbedFrameProps {
  title: string;
  // A page of another site, with a link to open it on its own.
  url?: string;
  // A complete HTML document from the deployment, shown in place. It can run scripts, so the frame
  // has no allow-same-origin: it gets an opaque origin and can't reach this site's cookies, storage
  // or admin API. It may post its height to the page, see utils/embedHeight.
  html?: string;
}

// What the deployment embeds under a page's text.
export default function EmbedFrame({ title, url, html }: EmbedFrameProps) {
  const frame = useRef<HTMLIFrameElement>(null);
  const [height, setHeight] = useState(DEFAULT_EMBED_HEIGHT);

  useEffect(() => {
    if (!html) return;
    const onMessage = (event: MessageEvent) => {
      // only this frame, the opaque origin of a sandboxed document can't be compared
      if (event.source !== frame.current?.contentWindow) return;
      const next = embedHeightFrom(event.data);
      if (next !== null) setHeight(next);
    };
    window.addEventListener('message', onMessage);
    return () => window.removeEventListener('message', onMessage);
  }, [html]);

  if (html) {
    return (
      <section className="pt-2">
        <iframe
          ref={frame}
          srcDoc={html}
          title={title}
          loading="lazy"
          // popups may leave the sandbox, so a link the widget opens (tickets) works as a normal page
          sandbox="allow-scripts allow-popups allow-popups-to-escape-sandbox"
          style={{ height }}
          className="w-full rounded-control border border-border bg-surface"
        />
      </section>
    );
  }

  if (!url) return null;

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
