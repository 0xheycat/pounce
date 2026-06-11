import type { Download, Settings } from "../types";

// In dev, Vite proxies /api to the engine. In production the engine serves
// the dashboard, so relative URLs are always correct.
const BASE = import.meta.env.VITE_ENGINE_URL ?? "";
const TOKEN_KEY = "pounce.token";

export function getToken(): string {
  return localStorage.getItem(TOKEN_KEY) ?? "";
}

export function setToken(token: string): void {
  if (token) localStorage.setItem(TOKEN_KEY, token);
  else localStorage.removeItem(TOKEN_KEY);
}

// captureTokenFromUrl reads a ?token=... query param (used by device pairing
// links / QR codes), stores it, then strips it from the address bar.
export function captureTokenFromUrl(): void {
  try {
    const params = new URLSearchParams(window.location.search);
    const t = params.get("token");
    if (!t) return;
    setToken(t);
    params.delete("token");
    const qs = params.toString();
    const clean =
      window.location.pathname + (qs ? `?${qs}` : "") + window.location.hash;
    window.history.replaceState({}, "", clean);
  } catch {
    // non-browser context or blocked history API; ignore
  }
}

function authHeaders(): Record<string, string> {
  const t = getToken();
  return t ? { Authorization: `Bearer ${t}` } : {};
}

async function req<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: { "Content-Type": "application/json", ...authHeaders() },
    ...init,
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error ?? `request failed: ${res.status}`);
  }
  return res.json() as Promise<T>;
}

export interface AddInput {
  url: string;
  dir?: string;
  connections?: number;
  speedLimit?: number;
  ytdlp?: boolean;
}

export const api = {
  list: () => req<Download[]>("/api/downloads"),

  add: (input: AddInput) =>
    req<Download>("/api/downloads", {
      method: "POST",
      body: JSON.stringify(input),
    }),

  pause: (id: string) => req(`/api/downloads/${id}/pause`, { method: "POST" }),
  resume: (id: string) =>
    req(`/api/downloads/${id}/resume`, { method: "POST" }),
  cancel: (id: string) =>
    req(`/api/downloads/${id}/cancel`, { method: "POST" }),
  remove: (id: string) => req(`/api/downloads/${id}`, { method: "DELETE" }),
  setSpeed: (id: string, limit: number) =>
    req(`/api/downloads/${id}/speed`, {
      method: "POST",
      body: JSON.stringify({ limit }),
    }),

  getSettings: () => req<Settings>("/api/settings"),
  saveSettings: (s: Settings) =>
    req<Settings>("/api/settings", { method: "PUT", body: JSON.stringify(s) }),
};

// subscribe opens the SSE stream and calls onDownload for every update. The
// token (if any) goes in the query string because EventSource can't set headers.
export function subscribe(onDownload: (d: Download) => void): () => void {
  const token = getToken();
  const url = `${BASE}/api/events${token ? `?token=${encodeURIComponent(token)}` : ""}`;
  const es = new EventSource(url);
  es.onmessage = (e) => {
    try {
      const msg = JSON.parse(e.data);
      if (msg.type === "download") onDownload(msg.data as Download);
    } catch {
      // ignore malformed frames
    }
  };
  return () => es.close();
}
