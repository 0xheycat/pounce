// Theme palettes. Hex values are mirrored into CSS custom properties (consumed
// by Tailwind + index.css) and read directly by the three.js scene, because
// WebGL materials cannot resolve CSS variables.
export interface Theme {
  key: string;
  name: string;
  bg: string;
  card: string;
  accent: string;
  glow: string;
  core: string;
}

export const themes: Theme[] = [
  {
    key: "midnight",
    name: "Midnight",
    bg: "#070711",
    card: "#11121f",
    accent: "#7c5cff",
    glow: "#22d3ee",
    core: "#a78bfa",
  },
  {
    key: "aurora",
    name: "Aurora",
    bg: "#03130f",
    card: "#0c1f1a",
    accent: "#34d399",
    glow: "#22d3ee",
    core: "#6ee7b7",
  },
  {
    key: "sunset",
    name: "Sunset",
    bg: "#160a0c",
    card: "#23121a",
    accent: "#fb7185",
    glow: "#fbbf24",
    core: "#f97316",
  },
  {
    key: "matrix",
    name: "Matrix",
    bg: "#00120a",
    card: "#04210f",
    accent: "#22c55e",
    glow: "#4ade80",
    core: "#bbf7d0",
  },
  {
    key: "mono",
    name: "Mono",
    bg: "#0a0a0b",
    card: "#161618",
    accent: "#e5e7eb",
    glow: "#9ca3af",
    core: "#f3f4f6",
  },
];

export const defaultThemeKey = "midnight";

export function getTheme(key: string): Theme {
  return themes.find((t) => t.key === key) ?? themes[0];
}

// applyTheme writes the palette into CSS variables on <html>. Safe to call on
// every theme change and on startup.
export function applyTheme(key: string): Theme {
  const t = getTheme(key);
  if (typeof document !== "undefined") {
    const r = document.documentElement;
    r.style.setProperty("--pounce-bg", t.bg);
    r.style.setProperty("--pounce-card", t.card);
    r.style.setProperty("--pounce-accent", t.accent);
    r.style.setProperty("--pounce-glow", t.glow);
    r.style.setProperty("--pounce-core", t.core);
  }
  return t;
}
