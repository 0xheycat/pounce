# Pounce Architecture

Pounce is split into two cleanly separated parts: a **headless engine** (Go) and a **dashboard** (React + React Three Fiber). They communicate over a small HTTP API.

## Components

### Engine (`engine/`)
A dependency-free Go daemon. Packages:

| Package | Responsibility |
|---------|----------------|
| `internal/model` | Core types: `Download`, `Segment`, `Status`. |
| `internal/store` | Persists downloads as JSON (atomic writes). Pluggable; SQLite is a planned backend. |
| `internal/ratelimit` | Token-bucket limiter (bytes/sec), live-adjustable. |
| `internal/download` | The engine: probing, segmented transfers, resume, queueing, progress. |
| `internal/api` | REST + SSE HTTP layer, optional static file serving, SSE `Hub`. |
| `cmd/pounce` | Entrypoint / flag parsing / wiring. |

### Dashboard (`dashboard/`)
React + Vite + Tailwind + React Three Fiber + Zustand.
- `api/client.ts` — REST calls + `EventSource` SSE subscription.
- `store.ts` — Zustand store, kept in sync by SSE.
- `scene/` — the 3D world (orbiting orbs, central core, starfield).
- `components/` — add bar, stats, list, cards.

## How resume works
1. On **add**, the engine probes the URL (HEAD, then a 1-byte ranged GET) to learn the size and whether the server supports `Accept-Ranges`.
2. If resumable, the file is split into N **segments**, each a byte range.
3. Each segment downloads over its own connection and writes to the target file with `WriteAt` at absolute offsets (one shared file handle).
4. Every segment's `downloaded` counter is persisted to `~/.pounce/meta/<id>.json`.
5. On **pause** the context is cancelled; partial bytes and counters remain.
6. On **resume** (or after a restart) each segment requests `Range: bytes=(start+downloaded)-end` and continues.
7. When all segments complete, the `.pdownload` file is atomically renamed to the final filename.

Servers without range support fall back to a single stream; those downloads restart on resume (documented limitation, same as IDM).

## Realtime updates
The engine pushes a JSON frame per download change to all SSE clients via a fan-out `Hub`. Slow clients drop frames instead of blocking the engine. The dashboard upserts each frame into the Zustand store, which re-renders both the list and the 3D scene.

## Why SSE instead of WebSockets?
SSE needs no third-party library and is a perfect fit for one-way server→client streaming, keeping the engine 100% standard library. A WebSocket transport (for two-way control) is a welcome future contribution.

## Security notes
- The engine binds to `127.0.0.1` by default. Exposing it to a network should be gated behind the planned auth layer.
- CORS is permissive to support the dev proxy; tighten it for remote deployments.
