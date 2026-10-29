import { afterEach, expect, test } from "vitest";
import { terminalFont, terminalTheme } from "./theme";

afterEach(() => {
  delete document.documentElement.dataset.theme;
  document.documentElement.style.cssText = "";
});

test("the dark theme uses the design's ANSI palette, and magenta is allowed in program output", () => {
  const theme = terminalTheme();
  expect(theme.red).toBe("#e5566f");
  expect(theme.magenta).toBe("#c58ad8");
  expect(theme.brightWhite).toBe("#ffffff");
});

test("the light theme swaps the palette", () => {
  document.documentElement.dataset.theme = "light";
  const theme = terminalTheme();
  expect(theme.black).toBe("#131416");
  expect(theme.green).toBe("#1e7b4f");
});

test("background, text and cursor come from the tokens in use, with the selection at 35% of the accent", () => {
  document.documentElement.style.setProperty("--surface", "#101010");
  document.documentElement.style.setProperty("--text", "#eeeeee");
  document.documentElement.style.setProperty("--accent", "#4a8ad6");
  const theme = terminalTheme();
  expect(theme.background).toBe("#101010");
  expect(theme.foreground).toBe("#eeeeee");
  expect(theme.cursor).toBe("#4a8ad6");
  expect(theme.selectionBackground).toBe("#4a8ad659");
});

test("the font is the monospace token, or a generic monospace when there is none", () => {
  expect(terminalFont()).toBe("monospace");
  document.documentElement.style.setProperty("--font-mono", "Menlo, monospace");
  expect(terminalFont()).toBe("Menlo, monospace");
});
