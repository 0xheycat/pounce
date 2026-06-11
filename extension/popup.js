// Popup UI: toggle auto-capture, configure the engine URL + auth token, and
// run a quick health check. Everything is stored in chrome.storage.local and
// read back by background.js.
const DEFAULT_ENGINE = "http://127.0.0.1:7766"

const els = {
  capture: document.getElementById("capture"),
  engine: document.getElementById("engine"),
  token: document.getElementById("token"),
  test: document.getElementById("test"),
  status: document.getElementById("status"),
}

chrome.storage.local.get(["captureEnabled", "engineUrl", "authToken"], (v) => {
  els.capture.checked = v.captureEnabled !== false
  els.engine.value = v.engineUrl || DEFAULT_ENGINE
  els.token.value = v.authToken || ""
})

function flash(message) {
  els.status.textContent = message
  setTimeout(() => {
    els.status.textContent = ""
  }, 1400)
}

function save() {
  chrome.storage.local.set(
    {
      captureEnabled: els.capture.checked,
      engineUrl: els.engine.value.trim() || DEFAULT_ENGINE,
      authToken: els.token.value.trim(),
    },
    () => flash("Saved \u2713"),
  )
}

els.capture.addEventListener("change", save)
els.engine.addEventListener("change", save)
els.token.addEventListener("change", save)

els.test.addEventListener("click", async () => {
  const engine = els.engine.value.trim() || DEFAULT_ENGINE
  els.status.textContent = "Testing…"
  try {
    const res = await fetch(engine + "/api/health")
    els.status.textContent = res.ok ? "Engine online \ud83d\udc3e" : "Engine error " + res.status
  } catch (err) {
    els.status.textContent = "Engine unreachable"
  }
})
