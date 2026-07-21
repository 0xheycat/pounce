#!/usr/bin/env bash
# Local mirror of .github/workflows/ci.yml.
# Runs the exact same gates as CI, without needing GitHub Actions.
# Requirements: Go (>=1.23) and Node (>=20.19) installed locally.
#
# Usage:  bash scripts/check.sh
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "==> Engine (Go)"
cd "$root/engine"
echo "--> gofmt"
fmt_out="$(gofmt -l .)"
if [ -n "$fmt_out" ]; then
  echo "These files need gofmt:"
  echo "$fmt_out"
  echo "Run: (cd engine && gofmt -w ./...)"
  exit 1
fi
echo "--> go vet"
go vet ./...
echo "--> go test"
go test ./...
echo "--> go build"
go build ./...

echo "==> Dashboard (React)"
cd "$root/dashboard"
echo "--> npm ci"
npm ci
echo "--> type-check (npm run lint)"
npm run lint
echo "--> build (npm run build)"
npm run build
echo "--> audit production dependencies"
npm audit --omit=dev --audit-level=high

echo
echo "All checks passed. CI gates are green locally."
