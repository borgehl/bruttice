#!/bin/sh
# Runs when the production image starts (the nginx image runs every script in /docker-entrypoint.d).
#
# Puts the site's name and description into the page shell, so the browser tab, search engines and
# link previews show them. Those read the HTML as served and don't run JavaScript, so the name the
# app loads from /api/site is too late for them. The built page keeps its defaults (Bruttice) when
# SITE_NAME or SITE_DESCRIPTION is not set.
#
# The page is always rebuilt from the untouched copy, so a restart with other values starts clean.
set -eu

template=/usr/share/nginx/index.template.html
page=/usr/share/nginx/html/index.html

default_name='Bruttice'
default_description='A community event calendar.'

# a value is one line of text, shown as text in the page
html_escape() {
    printf '%s' "$1" | tr -d '\r\n' |
        sed -e 's/&/\&amp;/g' -e 's/</\&lt;/g' -e 's/>/\&gt;/g' -e 's/"/\&quot;/g' -e "s/'/\&#39;/g"
}

# and then a sed replacement, where \ & and the | delimiter are special
sed_escape() {
    printf '%s' "$1" | sed -e 's/[\\&|]/\\&/g'
}

name=$(sed_escape "$(html_escape "${SITE_NAME:-$default_name}")")
description=$(sed_escape "$(html_escape "${SITE_DESCRIPTION:-$default_description}")")

sed \
    -e "s|<title>$default_name</title>|<title>$name</title>|" \
    -e "s|content=\"$default_name\"|content=\"$name\"|g" \
    -e "s|content=\"$default_description\"|content=\"$description\"|g" \
    "$template" > "$page"
