import react from "@vitejs/plugin-react";
import { defineConfig } from "vitest/config";

export default defineConfig({
  plugins: [react()],
  // The taste-skill inspector expects the app on :5173 (DESIGN_RULES.md §1).
  server: { port: 5173, strictPort: true },
  test: { environment: "jsdom" },
});
