import optimizeLocales from "@react-aria/optimize-locales-plugin";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vitest/config";

export default defineConfig({
  plugins: [
    react(),
    // React Aria ships strings for ~30 locales; Astra has English and Indonesian (F11-06).
    { ...optimizeLocales.vite({ locales: ["en-US", "id-ID"] }), enforce: "pre" },
  ],
  // The taste-skill inspector expects the app on :5173 (DESIGN_RULES.md §1).
  // In development the daemon is reached through the dev server, so the page and the API share
  // an origin, as they do when the daemon serves the page itself (ADR 0006). The daemon must list
  // http://localhost:5173 in ASTRA_ORIGINS, because the proxy passes the browser's Origin on.
  server: {
    port: 5173,
    strictPort: true,
    proxy: { "/api": { target: "http://127.0.0.1:7437", ws: true } },
  },
  test: { environment: "jsdom", setupFiles: ["./src/test-setup.ts"] },
});
