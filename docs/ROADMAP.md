# Pounce Roadmap

Legend: ✅ done · 🚧 in progress · 🕜️ planned · 🐣 good first issue

## v0.1 — Foundation (current)
- ✅ Segmented multi-connection downloads (HTTP Range)
- ✅ Permanent resume across restarts (persisted per-segment progress)
- ✅ Pause / resume / cancel
- ✅ Live token-bucket speed limiting
- ✅ Choose save folder
- ✅ REST + SSE API
- ✅ Full 3D dashboard (orbs, stats, glassmorphism)

## v0.2 — More sources
- 🕜️ 🐣 **yt-dlp module** — detect & download video/stream URLs (`internal/sources/ytdlp.go` stub).
- 🕜️ **Torrent / magnet** via aria2 sidecar.
- 🕜️ 🐣 **Checksum verification** (MD5/SHA-256) after completion.

## v0.3 — Control & automation
- 🕜️ **Scheduler** + bandwidth profiles (e.g. full speed at night).
- 🕜️ Auto-categorize (video/docs/apps), auto-rename, dedup by hash.
- 🕜️ Desktop / Telegram / webhook notifications.

## v0.4 — Capture & remote
- ✅ **Remote access** ("Pounce Anywhere") — `--remote` mode, bearer-token auth, LAN auto-detect, one-link/QR device pairing.
- ✅ **Installable PWA** — add the dashboard to your phone's home screen; offline app shell.
- 🚧 **Browser extension** link capture (MV3 scaffold in `extension/`).

## v0.5 — Platform
- 🕜️ 🐣 **SQLite store backend** (drop-in for the JSON store).
- 🕜️ **Tauri desktop app** (double-click installer).
- 🕜️ Mirror / multi-source single-file downloads.
- 🕜️ WebSocket transport (two-way) alongside SSE.

## Good first issues 🐣
- Add a `--auth-token` flag and a bearer-token check middleware in `internal/api`.
- Implement `store.SqliteStore` satisfying the same methods as `store.Store`.
- Add SHA-256 verification with an optional expected hash on add.
- Persist a rolling speed history and render a live chart in the dashboard.
