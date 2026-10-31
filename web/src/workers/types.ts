// Facts the daemon stores about a worker (PRD §7). Status is never stored; see status.ts.

type AgentActivity = "working" | "needs_input" | "idle";

type GateFact = { state: "none" } | { state: "failed"; cause: string } | { state: "passed" };

type DeploymentFact =
  | { state: "none" }
  | { state: "shipping" }
  | { state: "watching" }
  | { state: "live" }
  | { state: "rolled_back"; reason: string };

export interface Worker {
  id: string;
  title: string;
  project: string;
  branch: string;
  agent: "Claude Code" | "Codex";
  createdAt: number;
  /** Waiting for a free slot because the server already runs as many workers as it may (F10-05). */
  queued?: boolean;
  /** False for internal projects: the Client stage is skipped. */
  clientReview: boolean;
  agentActivity: AgentActivity;
  previewLive: boolean;
  shared: boolean;
  newFeedback: number;
  gate: GateFact;
  deployment: DeploymentFact;
  added: number;
  removed: number;
}
