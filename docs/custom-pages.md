# Custom pages

A deployment can add pages of its own, and replace the text of a few built-in ones, without changing the core. Pages are Markdown files in a directory the backend reads at startup.

```
pages/
  about.md          replaces the About introduction
  privacy-policy.md replaces the privacy policy text
  tk-night.md       a new page, at /p/tk-night
```

Set the directory with `pages-dir` in `config.yaml`. It defaults to `/app/pages`, and a missing directory just means no pages. Mount your files there (in Kubernetes, a ConfigMap built from the folder works, and so does a bind mount in Compose).

## A page

```
---
title: Pop! Booker
nav: true
order: 60
embed: https://app.example.com/calendar
---
Some **Markdown** text. [Links](https://example.com) to other sites open in a new tab.
```

The header is optional YAML between two `---` lines:

| Key | Description |
|---|---|
| `title` | Page title. Defaults to the file name, so `tk-night.md` is "Tk Night" |
| `nav` | `true` lists the page in the navigation: the header on a desktop, the More page on a phone. Default `false`, the page is then only reachable by its link |
| `order` | Position among the listed pages, lowest first. Ties go by title. Default 0 |
| `embed` | An `http(s)` page shown in a frame under the text, with an "Open in a new tab" link. Anything else is ignored |

The file name is the slug: lower case letters, digits and single dashes, then `.md`. A new page is at `/p/<slug>`, and `GET /api/pages` and `GET /api/pages/<slug>` serve the same content as JSON.

## Embedding with a script

Some embeds are a snippet of HTML with a script, not a page you can link to. Put a complete HTML document in a file with the same name as the page, with `.html` in place of `.md`:

```
pages/
  pop-booker.md     the page: title, text
  pop-booker.html   a full HTML document, shown under the text
```

It is shown in place of `embed`, in a frame that has no `allow-same-origin`. The script runs, but the document has an opaque origin: it can't read this site's cookies or storage or call its admin API, and it can't navigate the page. Links it opens in a new tab work as normal pages. Because the frame has its own document, style it there: set the `body` background and `margin: 0` or you will see white around the widget.

The frame starts 420px tall. To fit the content, have the document post its height:

```html
<script>
  new ResizeObserver(() => {
    parent.postMessage({ type: 'bruttice-embed-height', height: document.documentElement.scrollHeight }, '*');
  }).observe(document.documentElement);
</script>
```

Heights are kept between 120 and 2400px. A request the widget makes from the frame carries `Origin: null`, so the service it talks to must allow cross-origin requests. The `.html` file counts toward the 256 KB limit, is served as the page's `embedHtml` field, and is skipped with a logged problem if there is no `.md` of the same name.

Markdown is rendered with `react-markdown`. Raw HTML in a page is not rendered, and `javascript:` links are dropped.

## Replacing built-in text

A page whose slug is one of these replaces that page's own text, and is not listed in the navigation a second time:

| Slug | Replaces |
|---|---|
| `about` | The introduction on About. How to add an event and the contact link stay, they come from the site |
| `privacy-policy` | The whole privacy policy text. Check it still describes what your deployment collects |

## Mistakes

A file that can't be used (a bad name, a broken header, bigger than 256 KB) is skipped, and the backend logs why at startup (`page skipped: ...`). The other pages still load. Pages are read once at startup, so restart the backend after changing a file.
