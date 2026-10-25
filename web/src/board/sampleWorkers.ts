import type { Worker } from "../workers/types";
import rows from "./sample-workers.json";

type Row = Partial<Omit<Worker, "createdAt">> & Pick<Worker, "id" | "title" | "project" | "branch" | "agent"> & { minutesAgo: number };

/** Sample workers for the M0 Board prototype; the daemon replaces this in M1. */
export function sampleWorkers(now: number): Worker[] {
  return (rows as Row[]).map(({ minutesAgo, ...row }) => ({
    clientReview: true,
    agentActivity: "idle",
    previewLive: false,
    shared: false,
    newFeedback: 0,
    gate: { state: "none" },
    deployment: { state: "none" },
    added: 0,
    removed: 0,
    ...row,
    createdAt: now - minutesAgo * 60_000,
  }));
}
