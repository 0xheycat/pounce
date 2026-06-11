# Pounce Capture (browser extension scaffold)

A Manifest V3 extension that hands browser downloads to the local Pounce engine.

> Status: **scaffold**. Context-menu + auto-capture wiring is in place; a popup UI, per-site rules, and an enable/disable toggle are great first contributions (see `docs/ROADMAP.md`).

## Load it (Chrome / Edge)
1. Make sure the Pounce engine is running (`http://127.0.0.1:7766`).
2. Go to `chrome://extensions`, enable **Developer mode**.
3. Click **Load unpacked** and select this `extension/` folder.

## What it does
- Adds a **“Download with Pounce 🐾”** right-click item for links, video, audio, and images.
- Optionally auto-captures any download the browser starts, cancels it, and forwards the URL to the engine.

## Notes
- Auto-capture defaults on; gate it behind a toggle stored in `chrome.storage.local` (`captureEnabled`).
- For Firefox, MV3 background needs `"scripts"` instead of `"service_worker"`; a cross-browser build is a planned enhancement.
