# Changelog

All notable changes to **Pounce** are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

> This file is generated automatically from [Conventional Commits](https://www.conventionalcommits.org/) by [`git-cliff`](https://git-cliff.org). Do not edit it by hand — run `scripts/release.sh vX.Y.Z`.

## [Unreleased]

### 🚀 Features
- Self-hostable download engine with permanent resume and per-segment progress.
- Multi-connection segmented downloads (HTTP Range), up to 32 streams.
- Live token-bucket speed limiting, adjustable per download.
- Optional bearer-token authentication for remote access.
- Persistent settings (themes, defaults, notifications) via `/api/settings`.
- yt-dlp source resolver for video/stream URLs.
- Full-3D, themeable web dashboard (5 themes) with desktop completion notifications.
- **Pounce Anywhere** remote access: `--remote` flag (binds all interfaces, auto-generates a token), LAN address auto-detection, and one-link/QR device pairing.
- Installable PWA dashboard (web app manifest + offline service worker) with `?token=` pairing capture.
- One-command Docker build (multi-stage Dockerfile + docker-compose) and a Makefile, so the project builds and runs without a local Go/Node toolchain.
- MV3 browser extension scaffold with link capture, popup, and configurable engine URL/token.

### 📚 Documentation
- Positioning guide, architecture, and roadmap.

## [0.1.0] - 2026-06-11

### 🚀 Features
- Initial public foundation: engine (download/resume/segment/throttle), REST + SSE API, and 3D dashboard.

[Unreleased]: https://github.com/0xheycat/pounce/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/0xheycat/pounce/releases/tag/v0.1.0
