#!/usr/bin/env bash
# Cut a new Pounce release: regenerate the changelog from Conventional Commits,
# commit it, and create an annotated tag. Pushing the tag triggers the Release
# workflow which builds binaries and publishes the GitHub Release.
#
# Usage: scripts/release.sh v0.2.0
set -euo pipefail

VERSION="${1:-}"
if [ -z "$VERSION" ]; then
  echo "Usage: scripts/release.sh vX.Y.Z"
  exit 1
fi
case "$VERSION" in
  v[0-9]*) ;;
  *) echo "Version must start with 'v', e.g. v0.2.0"; exit 1 ;;
esac

if ! command -v git-cliff >/dev/null 2>&1; then
  echo "git-cliff is required. Install it:"
  echo "  cargo install git-cliff   # or see https://git-cliff.org/docs/installation"
  exit 1
fi

if [ -n "$(git status --porcelain)" ]; then
  echo "Working tree is dirty. Commit or stash changes first."
  exit 1
fi

echo "-> Regenerating CHANGELOG.md for $VERSION"
git-cliff --tag "$VERSION" -o CHANGELOG.md

git add CHANGELOG.md
git commit -m "chore(release): $VERSION"
git tag -a "$VERSION" -m "Pounce $VERSION"

echo
echo "Tagged $VERSION."
echo "Push it to trigger the release build:"
echo "  git push origin main --follow-tags"
