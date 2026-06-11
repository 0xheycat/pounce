import { defineConfig } from "vite"
import react from "@vitejs/plugin-react"

// During development the dashboard runs on :5173 and proxies API + SSE calls
// to the Go engine on :7766. In production the engine serves the built files
// directly, so same-origin requests just work.
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      "/api": {
        target: "http://127.0.0.1:7766",
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: "dist",
  },
})
