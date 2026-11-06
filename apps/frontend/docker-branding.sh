#!/bin/sh
# Runs when the production image starts (the nginx image runs every script in /docker-entrypoint.d).
#
# Lets a site put its own icons, web manifest or any other static file in the built site, without
# building its own image: every file in the branding directory is copied over the file of the same
# name in the site's root. The usual use is the icons the page links to (favicon.ico, the
# favicon-*.png, apple-touch-icon.png) and site.webmanifest, see docs/branding.md.
#
# The directory is /usr/share/nginx/branding, or BRANDING_DIR. With no such directory nothing
# happens, so a site that does not use this is unchanged.
#
# Only the files directly in the directory are used. A Kubernetes ConfigMap mounted there is a
# directory of symlinks plus hidden entries (..data and a timestamped directory), the glob below
# skips the hidden ones and cp -L copies what the symlinks point to.
#
# index.html is never replaced: docker-site-meta.sh builds it from SITE_NAME and SITE_DESCRIPTION.
#
# The copy is not undone when a file leaves the directory, the container keeps what it copied until
# it is recreated (a new pod in Kubernetes, `docker compose up --force-recreate`).
set -eu

branding=${BRANDING_DIR:-/usr/share/nginx/branding}
site=/usr/share/nginx/html

if [ ! -d "$branding" ]; then
    exit 0
fi

for file in "$branding"/*; do
    # a glob that matches nothing stays as written, and directories are not copied
    [ -f "$file" ] || continue

    name=$(basename "$file")
    if [ "$name" = "index.html" ]; then
        echo "branding: skipping index.html, the page is built from SITE_NAME and SITE_DESCRIPTION" >&2
        continue
    fi

    cp -L "$file" "$site/$name"
    echo "branding: $name"

    # the page links the default icon (favicon.svg) first, and a browser may prefer it to the
    # site's own icons, so the link goes once the site brings any of its own
    case "$name" in
        favicon.ico | favicon-*.png | apple-touch-icon.png)
            sed -i '/favicon\.svg/d' "$site/index.html"
            ;;
    esac
done
