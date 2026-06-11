import { create } from "zustand";
import type { Download, Settings } from "./types";
import { api, subscribe, captureTokenFromUrl } from "./api/client";
import { applyTheme, defaultThemeKey } from "./themes";

const THEME_KEY = "pounce.theme";

interface State {
  downloads: Record<string, Download>;
  connected: boolean;
  theme: string;
  settings: Settings | null;
  ordered: () => Download[];
  init: () => Promise<void>;
  upsert: (d: Download) => void;
  setTheme: (key: string) => void;
  saveSettings: (next: Settings) => Promise<void>;
}

export const useStore = create<State>((set, get) => ({
  downloads: {},
  connected: false,
  theme: localStorage.getItem(THEME_KEY) ?? defaultThemeKey,
  settings: null,

  ordered: () =>
    Object.values(get().downloads).sort(
      (a, b) =>
        new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime(),
    ),

  init: async () => {
    // A device may have been paired via a ?token=... link; capture it first.
    captureTokenFromUrl();

    // Apply the locally-remembered theme immediately to avoid a flash.
    applyTheme(get().theme);

    try {
      const list = await api.list();
      const map: Record<string, Download> = {};
      for (const d of list) map[d.id] = d;
      set({ downloads: map });
    } catch {
      // engine may not be up yet; SSE will backfill
    }

    try {
      const settings = await api.getSettings();
      set({ settings });
      if (settings.theme) {
        set({ theme: settings.theme });
        applyTheme(settings.theme);
      }
    } catch {
      // keep the local theme if the engine has no settings yet
    }

    // Ask for notification permission up front so completion alerts work.
    if (
      typeof Notification !== "undefined" &&
      Notification.permission === "default"
    ) {
      void Notification.requestPermission().catch(() => undefined);
    }

    subscribe((d) => get().upsert(d));
    set({ connected: true });
  },

  upsert: (d) => {
    const prev = get().downloads[d.id];
    const notify = get().settings?.notifyOnComplete ?? true;
    if (
      notify &&
      d.status === "completed" &&
      prev &&
      prev.status !== "completed"
    ) {
      notifyComplete(d.filename);
    }
    if (d.status === "canceled") {
      set((s) => {
        const next = { ...s.downloads };
        delete next[d.id];
        return { downloads: next };
      });
      return;
    }
    set((s) => ({ downloads: { ...s.downloads, [d.id]: d } }));
  },

  setTheme: (key) => {
    localStorage.setItem(THEME_KEY, key);
    applyTheme(key);
    set({ theme: key });
    const cur = get().settings;
    if (cur)
      void api.saveSettings({ ...cur, theme: key }).catch(() => undefined);
  },

  saveSettings: async (next) => {
    const saved = await api.saveSettings(next);
    const theme = saved.theme || get().theme;
    localStorage.setItem(THEME_KEY, theme);
    applyTheme(theme);
    set({ settings: saved, theme });
  },
}));

function notifyComplete(name: string) {
  try {
    if (
      typeof Notification === "undefined" ||
      Notification.permission !== "granted"
    )
      return;
    new Notification("Pounce \u2014 download complete \ud83d\udc3e", {
      body: name,
    });
  } catch {
    // notifications are best-effort
  }
}
