#!/usr/bin/env bash
# Validate that commit subjects on this branch follow Conventional Commits.
# Used by the "Commit Lint" workflow and runnable locally:  bash scripts/check-commits.sh
set -euo pipefail

pattern='^(feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\([a-z0-9_.-]+\))?!?: .+'
base_ref="${GITHUB_BASE_REF:-main}"

git fetch --quiet origin "$base_ref" 2>/dev/null || true
if git rev-parse --verify --quiet "origin/$base_ref" >/dev/null; then
  range="origin/$base_ref..HEAD"
else
  range=""
fi

if [ -n "$range" ]; then
  subjects=$(git log --format=%s "$range")
else
  subjects=$(git log --format=%s -n 20)
fi

fail=0
while IFS= read -r line; do
  [ -z "$line" ] && continue
  case "$line" in
    "Merge "*) continue ;;
  esac
  if printf '%s' "$line" | grep -Eq "$pattern"; then
    echo "  ok   $line"
  else
    echo "  BAD  $line"
    fail=1
  fi
done <<EOF
$subjects
EOF

if [ "$fail" -ne 0 ]; then
  echo
  echo "Commit messages must follow Conventional Commits:"
  echo "  https://www.conventionalcommits.org"
  echo "Example:  feat(engine): add checksum verification"
  exit 1
fi

echo
echo "All commit messages follow Conventional Commits."
