import { expect, test } from "vitest";
import { connectionText } from "./ConnectionStatus";

const at = new Date(2026, 9, 29, 9, 14).getTime();

test("the status bar says the state of the connection in words, and how old the data is when it is not live", () => {
  expect(connectionText({ phase: "live", lastUpdate: at, retryAt: null })).toBe("Connected to the daemon");
  expect(connectionText({ phase: "connecting", lastUpdate: null, retryAt: null })).toBe("Connecting to the daemon…");
  expect(connectionText({ phase: "connecting", lastUpdate: at, retryAt: null })).toBe("Reconnecting · last updated 09:14");
  expect(connectionText({ phase: "offline", lastUpdate: null, retryAt: at })).toBe("Can't reach the daemon · retrying");
  expect(connectionText({ phase: "offline", lastUpdate: at, retryAt: at })).toBe("Offline · last updated 09:14 · retrying");
  expect(connectionText({ phase: "unauthorized", lastUpdate: at, retryAt: null })).toBe("This device is no longer paired");
});
