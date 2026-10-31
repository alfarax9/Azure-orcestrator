// What the daemon sends (ADR 0006): facts, never statuses. internal/domain/*.go is the source.
import type { Worker } from "../workers/types";

export interface ApiWorkspace {
  id: string;
  slug: string;
  name: string;
  createdAt: number;
}

export interface ApiProject {
  id: string;
  workspaceId: string;
  name: string;
  repoUrl?: string;
  defaultBranch: string;
  stack?: string;
  createdAt: number;
}

export interface ApiWorker {
  id: string;
  projectId: string;
  title: string;
  branch: string;
  agent: string;
  createdAt: number;
  queued?: boolean;
  clientReview: boolean;
  agentActivity: Worker["agentActivity"];
  previewLive: boolean;
  shared: boolean;
  newFeedback: number;
  gate: { state: "none" | "failed" | "passed"; cause?: string };
  deployment: { state: "none" | "shipping" | "watching" | "live" | "rolled_back"; reason?: string };
  added: number;
  removed: number;
}

interface Snapshot {
  workspaces: ApiWorkspace[];
  projects: ApiProject[];
  workers: ApiWorker[];
}

/** One message on the events socket. */
export type ServerMessage =
  | { type: "snapshot"; data: Snapshot }
  | { type: "workspace"; data: ApiWorkspace }
  | { type: "project"; data: ApiProject }
  | { type: "worker"; data: ApiWorker };

function isObject(v: unknown): v is Record<string, unknown> {
  return typeof v === "object" && v !== null && !Array.isArray(v);
}

function hasId(v: unknown): boolean {
  return isObject(v) && typeof v.id === "string";
}

function isList(v: unknown): boolean {
  return Array.isArray(v) && v.every(hasId);
}

/** Reads one message, or null for anything that is not a message this client understands. */
export function parseMessage(text: string): ServerMessage | null {
  let raw: unknown;
  try {
    raw = JSON.parse(text);
  } catch {
    return null;
  }
  if (!isObject(raw) || typeof raw.type !== "string") return null;
  const data = raw.data;
  switch (raw.type) {
    case "snapshot":
      if (isObject(data) && isList(data.workspaces) && isList(data.projects) && isList(data.workers)) {
        return { type: "snapshot", data: data as unknown as Snapshot };
      }
      return null;
    case "workspace":
    case "project":
    case "worker":
      return hasId(data) ? ({ type: raw.type, data } as ServerMessage) : null;
  }
  return null;
}

const AGENT_LABELS: Record<string, Worker["agent"]> = { claude: "Claude Code", codex: "Codex" };

/** The Board's worker for the daemon's facts; `project` is the label shown on cards. */
export function toWorker(w: ApiWorker, project: string): Worker {
  return {
    id: w.id,
    title: w.title,
    project,
    branch: w.branch,
    agent: AGENT_LABELS[w.agent] ?? "Claude Code",
    createdAt: w.createdAt,
    queued: w.queued ?? false,
    clientReview: w.clientReview,
    agentActivity: w.agentActivity,
    previewLive: w.previewLive,
    shared: w.shared,
    newFeedback: w.newFeedback,
    gate: w.gate.state === "failed" ? { state: "failed", cause: w.gate.cause || "Gate failed" } : { state: w.gate.state },
    deployment:
      w.deployment.state === "rolled_back"
        ? { state: "rolled_back", reason: w.deployment.reason || "Rolled back" }
        : { state: w.deployment.state },
    added: w.added,
    removed: w.removed,
  };
}
