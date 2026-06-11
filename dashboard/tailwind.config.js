/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        // Backed by CSS variables so the theme switcher restyles everything live.
        pounce: {
          bg: "var(--pounce-bg)",
          card: "var(--pounce-card)",
          accent: "var(--pounce-accent)",
          glow: "var(--pounce-glow)",
          core: "var(--pounce-core)",
        },
      },
      fontFamily: {
        sans: ["Inter", "system-ui", "sans-serif"],
      },
    },
  },
  plugins: [],
}
