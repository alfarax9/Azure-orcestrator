import type { ITheme } from "@xterm/xterm";

// ANSI palette from DESIGN_SYSTEM.md §2.3. Program output may use every hue; the app's own UI may not.
const DARK = {
  black: "#3a3a3e", brightBlack: "#8a8c92",
  red: "#e5566f", brightRed: "#f07a8e",
  green: "#5fc39c", brightGreen: "#7fd6b2",
  yellow: "#d9a441", brightYellow: "#e8bd66",
  blue: "#5b9be0", brightBlue: "#82b4ec",
  magenta: "#c58ad8", brightMagenta: "#d6a6e4",
  cyan: "#4fb6c4", brightCyan: "#76cbd6",
  white: "#bdbdc0", brightWhite: "#ffffff",
};

const LIGHT = {
  black: "#131416", brightBlack: "#5d5f60",
  red: "#b4232a", brightRed: "#b4232a",
  green: "#1e7b4f", brightGreen: "#1e7b4f",
  yellow: "#8a5a00", brightYellow: "#8a5a00",
  blue: "#1d5fbf", brightBlue: "#1d5fbf",
  magenta: "#8b3fa8", brightMagenta: "#8b3fa8",
  cyan: "#0f6f7d", brightCyan: "#0f6f7d",
  white: "#5d5f60", brightWhite: "#131416",
};

/** The terminal's colours: ANSI from the palette, background, text and cursor from the tokens in use. */
export function terminalTheme(root: HTMLElement = document.documentElement): ITheme {
  const css = getComputedStyle(root);
  const token = (name: string, fallback: string) => css.getPropertyValue(name).trim() || fallback;
  const accent = token("--accent", "#4a8ad6");
  return {
    ...(root.dataset.theme === "light" ? LIGHT : DARK),
    background: token("--surface", "#131217"),
    foreground: token("--text", "#ffffff"),
    cursor: accent,
    cursorAccent: token("--surface", "#131217"),
    selectionBackground: `${accent}59`, // the accent at 35% opacity
  };
}

/** The terminal font: the monospace token, so a user's system monospace applies (DESIGN_SYSTEM.md §3). */
export function terminalFont(root: HTMLElement = document.documentElement): string {
  return getComputedStyle(root).getPropertyValue("--font-mono").trim() || "monospace";
}
