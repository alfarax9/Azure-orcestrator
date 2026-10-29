// The device token the daemon gave this browser when it paired (ADR 0006). Kept in
// localStorage so a reload does not ask to pair again; it never goes into a URL or a log.
const KEY = "astra.device-token";

let memory: string | null = null;

export function loadToken(): string | null {
  try {
    return localStorage.getItem(KEY) ?? memory;
  } catch {
    return memory; // storage blocked (private window): the token lives until the tab closes
  }
}

export function saveToken(token: string): void {
  memory = token;
  try {
    localStorage.setItem(KEY, token);
  } catch {
    // The in-memory copy is enough for this tab.
  }
}

export function clearToken(): void {
  memory = null;
  try {
    localStorage.removeItem(KEY);
  } catch {
    // Nothing stored that could be removed.
  }
}
