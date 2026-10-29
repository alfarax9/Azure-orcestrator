import type { ConnectionState } from "../daemon/connection";
import { useStatuses } from "../workers/store";

function clock(at: number): string {
  return new Date(at).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit", hour12: false });
}

/** What the status bar says about the daemon: the connection first, never only in colour (DESIGN_RULES.md §10). */
export function connectionText(state: ConnectionState): string {
  switch (state.phase) {
    case "live":
      return "Connected to the daemon";
    case "connecting":
      return state.lastUpdate === null ? "Connecting to the daemon…" : `Reconnecting · last updated ${clock(state.lastUpdate)}`;
    case "offline":
      return state.lastUpdate === null ? "Can't reach the daemon · retrying" : `Offline · last updated ${clock(state.lastUpdate)} · retrying`;
    case "unauthorized":
      return "This device is no longer paired";
  }
}

export function ConnectionStatus({ state }: { state: ConnectionState }) {
  const statuses = useStatuses();
  const working = Object.values(statuses).filter((s) => s === "working").length;
  return (
    <>
      <span className="status-bar-brand">Astra</span>
      <span role="status">{connectionText(state)}</span>
      <span>{working} working</span>
    </>
  );
}
