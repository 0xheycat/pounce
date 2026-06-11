import { useEffect, useState } from "react";
import type { ReactNode } from "react";
import { useStore } from "../store";
import { getToken, setToken } from "../api/client";
import type { Settings } from "../types";
import { ThemeSwitcher } from "./ThemeSwitcher";

const inputClass =
  "mt-1 w-full rounded-lg bg-black/30 px-3 py-2 text-sm outline-none ring-1 ring-white/10 focus:ring-pounce-accent";

export function SettingsPanel() {
  const [open, setOpen] = useState(false);
  const settings = useStore((s) => s.settings);
  const saveSettings = useStore((s) => s.saveSettings);
  const [draft, setDraft] = useState<Settings | null>(settings);
  const [token, setTok] = useState(getToken());
  const [saved, setSaved] = useState(false);

  useEffect(() => setDraft(settings), [settings]);

  async function save() {
    setToken(token);
    if (draft) await saveSettings(draft);
    setSaved(true);
    setTimeout(() => setSaved(false), 1500);
  }

  if (!open) {
    return (
      <button
        type="button"
        onClick={() => setOpen(true)}
        className="rounded-lg bg-white/5 px-3 py-1.5 text-sm text-white/70 transition hover:bg-white/15"
      >
        ⚙ Settings
      </button>
    );
  }

  const d = draft;

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4"
      onClick={() => setOpen(false)}
    >
      <div
        className="glass w-full max-w-md rounded-2xl p-6"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="mb-5 flex items-center justify-between">
          <h2 className="text-lg font-semibold">Settings</h2>
          <button
            type="button"
            onClick={() => setOpen(false)}
            className="text-white/40 hover:text-white"
          >
            ✕
          </button>
        </div>

        <Field label="Theme">
          <ThemeSwitcher />
        </Field>

        {d && (
          <div className="mt-4 space-y-4">
            <Field label="Default save folder">
              <input
                value={d.defaultDir}
                onChange={(e) => setDraft({ ...d, defaultDir: e.target.value })}
                placeholder="~/Downloads/Pounce"
                className={inputClass}
              />
            </Field>
            <Field label={`Default connections: ${d.defaultConnections}`}>
              <input
                type="range"
                min={1}
                max={32}
                value={d.defaultConnections}
                onChange={(e) =>
                  setDraft({ ...d, defaultConnections: Number(e.target.value) })
                }
                className="w-full"
              />
            </Field>
            <Field label={`Max concurrent downloads: ${d.maxConcurrent}`}>
              <input
                type="range"
                min={1}
                max={10}
                value={d.maxConcurrent}
                onChange={(e) =>
                  setDraft({ ...d, maxConcurrent: Number(e.target.value) })
                }
                className="w-full"
              />
            </Field>
            <label className="flex items-center justify-between text-sm">
              <span className="text-white/60">
                Notify when a download finishes
              </span>
              <input
                type="checkbox"
                checked={d.notifyOnComplete}
                onChange={(e) =>
                  setDraft({ ...d, notifyOnComplete: e.target.checked })
                }
              />
            </label>
          </div>
        )}

        <div className="mt-4">
          <Field label="Engine auth token (optional)">
            <input
              value={token}
              onChange={(e) => setTok(e.target.value)}
              placeholder="leave empty if the engine has no token"
              className={inputClass}
            />
          </Field>
        </div>

        <button
          type="button"
          onClick={save}
          className="mt-6 w-full rounded-xl bg-pounce-accent py-3 font-semibold text-white transition hover:brightness-110"
        >
          {saved ? "Saved ✓" : "Save settings"}
        </button>
      </div>
    </div>
  );
}

function Field({ label, children }: { label: string; children: ReactNode }) {
  return (
    <label className="block text-sm">
      <span className="mb-1 block text-white/50">{label}</span>
      {children}
    </label>
  );
}
