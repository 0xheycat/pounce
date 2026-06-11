import { useEffect, useState } from "react";
import type { FormEvent } from "react";
import { api } from "../api/client";
import { useStore } from "../store";

export function AddBar() {
  const settings = useStore((s) => s.settings);
  const [url, setUrl] = useState("");
  const [dir, setDir] = useState("");
  const [connections, setConnections] = useState(8);
  const [speedKB, setSpeedKB] = useState(0); // 0 = unlimited, KB/s in the UI
  const [ytdlp, setYtdlp] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [open, setOpen] = useState(false);
  const [touched, setTouched] = useState(false);

  // Seed option defaults from saved settings until the user customizes them.
  useEffect(() => {
    if (!settings || touched) return;
    setConnections(settings.defaultConnections || 8);
    setSpeedKB(
      settings.defaultSpeedLimit > 0
        ? Math.round(settings.defaultSpeedLimit / 1024)
        : 0,
    );
    if (settings.defaultDir) setDir(settings.defaultDir);
  }, [settings, touched]);

  async function submit(e: FormEvent) {
    e.preventDefault();
    if (!url.trim()) return;
    setBusy(true);
    setError("");
    try {
      await api.add({
        url: url.trim(),
        dir: dir.trim() || undefined,
        connections,
        speedLimit: speedKB > 0 ? speedKB * 1024 : 0,
        ytdlp,
      });
      setUrl("");
    } catch (err) {
      setError(err instanceof Error ? err.message : "failed to add");
    } finally {
      setBusy(false);
    }
  }

  return (
    <form onSubmit={submit} className="glass mb-6 rounded-2xl p-4">
      <div className="flex gap-2">
        <input
          value={url}
          onChange={(e) => setUrl(e.target.value)}
          placeholder="Paste a download or video URL…"
          className="flex-1 rounded-xl bg-black/30 px-4 py-3 outline-none ring-1 ring-white/10 focus:ring-pounce-accent"
        />
        <button
          type="button"
          onClick={() => setOpen((v) => !v)}
          className="rounded-xl bg-white/5 px-4 text-sm text-white/60 hover:bg-white/10"
        >
          Options
        </button>
        <button
          type="submit"
          disabled={busy}
          className="rounded-xl bg-pounce-accent px-6 py-3 font-semibold text-white transition hover:brightness-110 disabled:opacity-50"
        >
          {busy ? "Adding…" : "Pounce"}
        </button>
      </div>

      {open && (
        <>
          <div className="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-3">
            <label className="text-sm">
              <span className="mb-1 block text-white/50">
                Save folder (optional)
              </span>
              <input
                value={dir}
                onChange={(e) => {
                  setTouched(true);
                  setDir(e.target.value);
                }}
                placeholder="~/Downloads/Pounce"
                className="w-full rounded-lg bg-black/30 px-3 py-2 ring-1 ring-white/10"
              />
            </label>
            <label className="text-sm">
              <span className="mb-1 block text-white/50">
                Connections: {connections}
              </span>
              <input
                type="range"
                min={1}
                max={32}
                value={connections}
                onChange={(e) => {
                  setTouched(true);
                  setConnections(Number(e.target.value));
                }}
                className="w-full"
              />
            </label>
            <label className="text-sm">
              <span className="mb-1 block text-white/50">
                Speed limit: {speedKB > 0 ? `${speedKB} KB/s` : "unlimited"}
              </span>
              <input
                type="range"
                min={0}
                max={10240}
                step={256}
                value={speedKB}
                onChange={(e) => {
                  setTouched(true);
                  setSpeedKB(Number(e.target.value));
                }}
                className="w-full"
              />
            </label>
          </div>
          <label className="mt-4 flex items-center gap-2 text-sm text-white/60">
            <input
              type="checkbox"
              checked={ytdlp}
              onChange={(e) => setYtdlp(e.target.checked)}
            />
            Resolve with yt-dlp (YouTube, Vimeo &amp; other video/stream sites)
          </label>
        </>
      )}

      {error && <p className="mt-3 text-sm text-red-400">{error}</p>}
    </form>
  );
}
