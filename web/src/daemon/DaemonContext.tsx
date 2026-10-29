import { createContext, useContext, useSyncExternalStore } from "react";
import type { Api } from "./api";
import type { ConnectionState } from "./connection";
import type { Facts, FactsState } from "./facts";

export interface Daemon {
  api: Api;
  token: string;
  connection: ConnectionState;
  facts: Facts;
  /** Forgets this device's token; the next screen is the pairing screen. */
  signOut: () => void;
}

export const DaemonContext = createContext<Daemon | null>(null);

export function useDaemon(): Daemon {
  const daemon = useContext(DaemonContext);
  if (!daemon) throw new Error("DaemonContext is missing");
  return daemon;
}

/** Workspaces, projects and whether the first snapshot has arrived. */
export function useFacts(): FactsState {
  const { facts } = useDaemon();
  return useSyncExternalStore(facts.subscribe, facts.getState);
}
