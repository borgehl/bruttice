# Branding: your own icons and static files

The frontend image can serve your site's icons, web manifest or any other static file without a custom image. Mount a directory of files at `/usr/share/nginx/branding`. When the container starts, every file in it is copied over the file of the same name in the site, so it is served as `/<file name>`.

If the directory does not exist nothing happens, and the image is unchanged.

## The files the page asks for

The page (`apps/frontend/index.html`) links these, so these are the names to provide:

| File | Used for | Size |
|---|---|---|
| `favicon.ico` | Browser tab icon, older browsers | 16, 32 and 48 px in one file |
| `favicon-16x16.png` | Browser tab icon | 16 x 16 |
| `favicon-32x32.png` | Browser tab icon | 32 x 32 |
| `apple-touch-icon.png` | Icon when the site is added to an iPhone or iPad home screen | 180 x 180 |
| `site.webmanifest` | Web app manifest: name, colors and the icons below | JSON |
| `android-chrome-192x192.png`, `android-chrome-512x512.png` | The app icons Android and Chrome use, named in your `site.webmanifest` | 192 x 192, 512 x 512 |

A web manifest looks like this. The names in `icons` are the files you provide:

```json
{
  "name": "Do It Together Tulsa",
  "short_name": "DIT Tulsa",
  "icons": [
    { "src": "/android-chrome-192x192.png", "sizes": "192x192", "type": "image/png" },
    { "src": "/android-chrome-512x512.png", "sizes": "512x512", "type": "image/png" }
  ],
  "theme_color": "#12161c",
  "background_color": "#12161c",
  "display": "standalone"
}
```

The core ships no icons of its own yet: without a branding directory these links answer 404 and browsers show their default icon.

You can add any other file the same way (a logo, an `og:image`). Files go straight in the directory, subdirectories are not copied.

## `index.html` is not replaceable

`index.html` is built at start from `SITE_NAME` and `SITE_DESCRIPTION`, see the README. A file with that name in the branding directory is skipped, with a line in the container's log.

## Docker Compose

Bind-mount a directory of your files:

```yaml
services:
  bruttice-frontend:
    volumes:
      - ./branding:/usr/share/nginx/branding:ro
```

`make up-demo` does this with `demo/branding/`, which holds an example `site.webmanifest`.

## Kubernetes

Build a ConfigMap from the folder and mount it. Mount the whole ConfigMap as a directory at `/usr/share/nginx/branding`, not file by file into the site's root:

```sh
kubectl create configmap my-branding -n my-namespace --from-file=branding --dry-run=client -o yaml | kubectl apply -f -
```

```yaml
      containers:
        - name: frontend
          volumeMounts:
            - name: branding
              mountPath: /usr/share/nginx/branding
      volumes:
        - name: branding
          configMap:
            name: my-branding
            optional: true
```

A ConfigMap is limited to 1 MiB in total, which is plenty for icons. Icons are binary: they work in a ConfigMap built `--from-file`, which stores them as `binaryData`.

## Changing the files

- **Restart.** The copy is made when the container starts. After you change the files, start a new container: roll the Kubernetes deployment, or `docker compose up --force-recreate`. In Kubernetes, give the pod template an annotation that changes on every deploy (the deployment repo's backend does this with a run id) so a changed ConfigMap restarts it.
- **Removing a file** does not take it out of a running container, it keeps what it copied until it is recreated.
- **Browser cache.** The site's nginx caches `.png` and `.ico` files for a year as `immutable`. A browser that already has your old icon keeps showing it until its cache expires or is cleared, whatever you deploy, so settle on the icons before launch. `site.webmanifest` is not cached that long.
- **Checking.** The container logs one `branding: <file>` line per file it copied. `curl -I http://localhost:8080/favicon.ico` should answer `200` with an image type. A missing icon or manifest answers `404`. A file that is not a valid image, or the wrong size, is served as it is.
