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

Markdown is rendered with `react-markdown`. Raw HTML in a page is not rendered, and `javascript:` links are dropped.

## Replacing built-in text

A page whose slug is one of these replaces that page's own text, and is not listed in the navigation a second time:

| Slug | Replaces |
|---|---|
| `about` | The introduction on About. How to add an event, the contact address and the featured venues stay, they come from the site |
| `privacy-policy` | The whole privacy policy text. Check it still describes what your deployment collects |

## Mistakes

A file that can't be used (a bad name, a broken header, bigger than 256 KB) is skipped, and the backend logs why at startup (`page skipped: ...`). The other pages still load. Pages are read once at startup, so restart the backend after changing a file.
