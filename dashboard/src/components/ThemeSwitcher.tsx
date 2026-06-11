import { useStore } from "../store";
import { themes } from "../themes";

// A row of color swatches; clicking one applies + persists the theme.
export function ThemeSwitcher() {
  const theme = useStore((s) => s.theme);
  const setTheme = useStore((s) => s.setTheme);

  return (
    <div className="flex items-center gap-1.5">
      {themes.map((t) => {
        const active = theme === t.key;
        const swatch = {
          background: `linear-gradient(135deg, ${t.accent}, ${t.glow})`,
        };
        return (
          <button
            key={t.key}
            type="button"
            title={t.name}
            aria-label={`Theme: ${t.name}`}
            onClick={() => setTheme(t.key)}
            style={swatch}
            className={`h-5 w-5 rounded-full transition ${
              active
                ? "ring-2 ring-white scale-110"
                : "ring-1 ring-white/20 hover:scale-105"
            }`}
          />
        );
      })}
    </div>
  );
}
