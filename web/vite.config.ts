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
  server: { port: 5173, strictPort: true },
  test: { environment: "jsdom" },
});
