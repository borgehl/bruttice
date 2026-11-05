#!/usr/bin/env bash
# Releases the next version of the whole repo: images, Go module and Python package.
#
#   scripts/bump-version.sh patch|minor|major [--dry-run]
#
# The version is the latest vX.Y.Z git tag plus one step. It creates a GitHub Release, and CI does
# the rest, see "Releasing" in the README. Nothing is changed locally.
#
#   patch  X.Y.Z -> X.Y.(Z+1)  a fix, nothing new
#   minor  X.Y.Z -> X.(Y+1).0  something new, or while the major is 0, a breaking change
#   major  X.Y.Z -> (X+1).0.0  a breaking change
set -euo pipefail

usage() {
	sed -n '2,12p' "$0" | sed 's/^# \{0,1\}//' >&2
	exit 2
}

part=${1:-}
dry_run=false
[ "${2:-}" = "--dry-run" ] && dry_run=true
[ $# -le 2 ] || usage
case "$part" in major | minor | patch) ;; *) usage ;; esac

cd "$(git rev-parse --show-toplevel)"

command -v gh >/dev/null || { echo "the GitHub CLI (gh) is needed" >&2; exit 1; }

branch=$(git rev-parse --abbrev-ref HEAD)
if [ "$branch" != "main" ]; then
	echo "release from main, you are on $branch" >&2
	exit 1
fi
if [ -n "$(git status --porcelain --untracked-files=no)" ]; then
	echo "commit or stash your changes first" >&2
	exit 1
fi

git fetch --quiet origin main --tags
if [ "$(git rev-parse HEAD)" != "$(git rev-parse origin/main)" ]; then
	echo "main is not the same as origin/main, push or pull first: a release is made from origin/main" >&2
	exit 1
fi

# plain releases only: the packages/go-pkg/v* tags and prereleases don't count
latest=$(git tag --list 'v[0-9]*' --sort=-v:refname | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' | head -n1 || true)
latest=${latest:-v0.0.0}

IFS=. read -r major minor patch <<<"${latest#v}"
case "$part" in
major) major=$((major + 1)) minor=0 patch=0 ;;
minor) minor=$((minor + 1)) patch=0 ;;
patch) patch=$((patch + 1)) ;;
esac
next="v$major.$minor.$patch"

echo "$latest -> $next ($(git rev-parse --short HEAD))"
git log --oneline "$([ "$latest" = v0.0.0 ] && echo "" || echo "$latest..")HEAD" | head -20

if $dry_run; then
	echo "dry run, nothing released"
	exit 0
fi

read -r -p "Release $next? [y/N] " answer
[ "$answer" = "y" ] || { echo "cancelled"; exit 1; }

gh release create "$next" --target "$(git rev-parse HEAD)" --title "$next" --generate-notes
echo "released $next, CI is publishing the images and tagging packages/go-pkg/$next"
