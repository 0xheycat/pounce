#!/usr/bin/env bash
# Run the Pounce engine and dashboard together for local development.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"

echo "🐾 Starting Pounce engine on http://127.0.0.1:7766 ..."
(cd "$ROOT/engine" && go run ./cmd/pounce) &
ENGINE_PID=$!

cleanup() { kill "$ENGINE_PID" 2>/dev/null || true; }
trap cleanup EXIT

echo "🐾 Starting dashboard on http://localhost:5173 ..."
cd "$ROOT/dashboard"
[ -d node_modules ] || npm install
npm run dev
