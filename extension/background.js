// Pounce Capture — MV3 background service worker (scaffold).
//
// Goal: intercept downloads the browser is about to start, cancel them, and
// forward the URL to the local Pounce engine so it can download with resume
// + acceleration. This is a scaffold: wire-up is intentionally minimal so it
// is easy to extend. See docs/ROADMAP.md (v0.4 — Capture & remote).

const DEFAULT_ENGINE = "http://127.0.0.1:7766"
const CAPTURE_KEY = "captureEnabled"
const TOKEN_KEY = "authToken"
const ENGINE_KEY = "engineUrl"

async function isCaptureEnabled() {
  const v = await chrome.storage.local.get(CAPTURE_KEY)
  return v[CAPTURE_KEY] !== false // default on
}

// Engine URL + optional auth token are configured from the popup.
async function getConfig() {
  const v = await chrome.storage.local.get([TOKEN_KEY, ENGINE_KEY])
  return { engine: v[ENGINE_KEY] || DEFAULT_ENGINE, token: v[TOKEN_KEY] || "" }
}

async function sendToPounce(url) {
  try {
    const { engine, token } = await getConfig()
    const headers = { "Content-Type": "application/json" }
    if (token) headers["Authorization"] = "Bearer " + token
    const res = await fetch(engine + "/api/downloads", {
      method: "POST",
      headers,
      body: JSON.stringify({ url, connections: 8 }),
    })
    if (!res.ok) throw new Error("engine responded " + res.status)
    notify("Sent to Pounce 🐾", url)
    return true
  } catch (err) {
    notify("Pounce engine not reachable", String(err))
    return false
  }
}

function notify(title, message) {
  chrome.notifications?.create({
    type: "basic",
    iconUrl: "data:image/svg+xml,",
    title,
    message: message.slice(0, 180),
  })
}

// Right-click any link -> "Download with Pounce".
chrome.runtime.onInstalled.addListener(() => {
  chrome.contextMenus.create({
    id: "pounce-download",
    title: "Download with Pounce 🐾",
    contexts: ["link", "video", "audio", "image"],
  })
})

chrome.contextMenus.onClicked.addListener((info) => {
  const url = info.linkUrl || info.srcUrl
  if (url) void sendToPounce(url)
})

// Auto-capture: cancel the browser download and forward it instead.
chrome.downloads.onCreated.addListener(async (item) => {
  if (!(await isCaptureEnabled())) return
  if (!item.finalUrl && !item.url) return
  const ok = await sendToPounce(item.finalUrl || item.url)
  if (ok) {
    try {
      await chrome.downloads.cancel(item.id)
      await chrome.downloads.erase({ id: item.id })
    } catch {
      // best effort
    }
  }
})
