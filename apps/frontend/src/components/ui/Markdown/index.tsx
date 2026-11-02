import ReactMarkdown, { type Components } from 'react-markdown';

// Page text is Markdown. Raw HTML is not rendered, and links to javascript: are dropped by the
// library, so a page can't run script. Each element gets the site's text styles.
const components: Components = {
  h1: ({ children }) => <h2 className="font-label pt-2 text-xl font-bold text-text">{children}</h2>,
  h2: ({ children }) => <h2 className="font-label pt-2 text-lg font-semibold text-text">{children}</h2>,
  h3: ({ children }) => <h3 className="font-label pt-1 text-base font-semibold text-text">{children}</h3>,
  p: ({ children }) => <p>{children}</p>,
  ul: ({ children }) => <ul className="list-disc space-y-1 pl-5">{children}</ul>,
  ol: ({ children }) => <ol className="list-decimal space-y-1 pl-5">{children}</ol>,
  blockquote: ({ children }) => (
    <blockquote className="border-l-2 border-accent pl-4 text-text-muted">{children}</blockquote>
  ),
  hr: () => <hr className="border-border" />,
  strong: ({ children }) => <strong className="font-semibold text-text">{children}</strong>,
  code: ({ children }) => <code className="font-data rounded-[4px] bg-raised px-1 text-[0.9em]">{children}</code>,
  a: ({ href, children }) => {
    // a link to another site opens in a new tab, one on this site stays in the page
    const external = typeof href === 'string' && /^https?:\/\//.test(href);
    return (
      <a href={href} {...(external ? { target: '_blank', rel: 'noopener noreferrer' } : {})}>
        {children}
      </a>
    );
  },
};

export default function Markdown({ children }: { children: string }) {
  return (
    <div className="space-y-4 leading-relaxed text-text-muted">
      <ReactMarkdown components={components} skipHtml>
        {children}
      </ReactMarkdown>
    </div>
  );
}
