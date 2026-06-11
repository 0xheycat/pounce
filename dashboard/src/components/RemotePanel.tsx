import { useState } from "react";
import QRCode from "react-qr-code";
import { getToken, setToken } from "../api/client";

// RemotePanel powers the "Pounce Anywhere" killer feature: scan a QR with your
// phone to open the dashboard pre-authenticated with the engine token.
export function RemotePanel() {
  const [open, setOpen] = useState(false);
  const [token, setTok] = useState(getToken());
  const [copied, setCopied] = useState(false);

  const origin = typeof window !== "undefined" ? window.location.origin : "";
  const host = typeof window !== "undefined" ? window.location.host : "";
  const pairUrl = token
    ? `${origin}/?token=${encodeURIComponent(token)}`
    : origin;

  function persist(value: string) {
    setTok(value);
    setToken(value);
  }

  async function copy() {
    try {
      await navigator.clipboard.writeText(pairUrl);
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    } catch {
      // clipboard may be blocked on plain HTTP; the link is shown below anyway
    }
  }

  if (!open) {
    return (
      <button
        type="button"
        onClick={() => setOpen(true)}
        className="rounded-lg bg-white/5 px-3 py-1.5 text-sm text-white/70 transition hover:bg-white/15"
      >
        📱 Pair device
      </button>
    );
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4"
      onClick={() => setOpen(false)}
    >
      <div
        className="glass w-full max-w-md rounded-2xl p-6"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="mb-1 flex items-center justify-between">
          <h2 className="text-lg font-semibold">📱 Pounce Anywhere</h2>
          <button
            type="button"
            onClick={() => setOpen(false)}
            className="text-white/40 hover:text-white"
          >
            ✕
          </button>
        </div>
        <p className="mb-5 text-sm text-white/50">
          Scan to open Pounce on your phone. Both devices must reach this engine
          — same network, or through your tunnel / reverse proxy.
        </p>

        <div className="flex flex-col items-center gap-4">
          <div className="rounded-xl bg-white p-3">
            <QRCode value={pairUrl} size={176} />
          </div>
          <code className="w-full truncate rounded-lg bg-black/30 px-3 py-2 text-center text-xs text-white/60">
            {pairUrl}
          </code>
          <button
            type="button"
            onClick={copy}
            className="w-full rounded-xl bg-pounce-accent py-2.5 font-semibold text-white transition hover:brightness-110"
          >
            {copied ? "Copied ✓" : "Copy pairing link"}
          </button>
        </div>

        <label className="mt-5 block text-sm">
          <span className="mb-1 block text-white/50">
            Engine token (match the engine's --auth-token / --remote token)
          </span>
          <input
            type="text"
            value={token}
            onChange={(e) => persist(e.target.value)}
            placeholder="paste the token printed by the engine"
            className="w-full rounded-lg bg-black/30 px-3 py-2 text-sm outline-none ring-1 ring-white/10 focus:ring-pounce-accent"
          />
        </label>

        <p className="mt-4 text-xs text-white/40">
          Reachable at <span className="text-white/60">{host}</span>. For access
          beyond your network, put Pounce behind HTTPS — see docs/REMOTE.md.
        </p>
      </div>
    </div>
  );
}
