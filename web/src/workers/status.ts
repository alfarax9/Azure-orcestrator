import type { Worker } from "./types";

// Status vocabulary from DESIGN_SYSTEM.md §10, derived from facts on read (F3-06).
export type WorkerStatus =
  | "queued"
  | "working"
  | "needs_you"
  | "preview_live"
  | "waiting_on_client"
  | "client_feedback"
  | "gate_failed"
  | "ready_to_ship"
  | "shipping"
  | "live"
  | "rolled_back";

export type BoardColumn = "working" | "needs_you" | "review" | "ready_to_ship" | "live";

export type StatusMark = "pulse-accent" | "dot-positive" | "pill-warning" | "pill-neutral" | "pill-negative" | "pill-positive";

export const STATUS: Record<WorkerStatus, { label: string; mark: StatusMark; column: BoardColumn }> = {
  queued: { label: "Queued", mark: "pill-neutral", column: "working" },
  working: { label: "Working", mark: "pulse-accent", column: "working" },
  needs_you: { label: "Needs you", mark: "pill-warning", column: "needs_you" },
  preview_live: { label: "Preview live", mark: "pill-neutral", column: "review" },
  waiting_on_client: { label: "Waiting on client", mark: "pill-neutral", column: "review" },
  // Feedback, failed gates and rollbacks wait on the developer, so they sit in Needs you.
  client_feedback: { label: "Client feedback", mark: "pill-warning", column: "needs_you" },
  gate_failed: { label: "Gate failed", mark: "pill-negative", column: "needs_you" },
  ready_to_ship: { label: "Ready to ship", mark: "pill-positive", column: "ready_to_ship" },
  shipping: { label: "Shipping", mark: "pulse-accent", column: "ready_to_ship" },
  live: { label: "Live", mark: "dot-positive", column: "live" },
  rolled_back: { label: "Rolled back", mark: "pill-warning", column: "needs_you" },
};

// Board column order (DESIGN_SYSTEM.md §9.2): Needs you is always second.
export const COLUMNS: { id: BoardColumn; label: string }[] = [
  { id: "working", label: "Working" },
  { id: "needs_you", label: "Needs you" },
  { id: "review", label: "Review" },
  { id: "ready_to_ship", label: "Ready to ship" },
  { id: "live", label: "Live" },
];

/** Later stages win: a live deployment outranks an old gate result, and so on. */
export function deriveStatus(w: Worker): WorkerStatus {
  if (w.queued) return "queued"; // not started: nothing else is true of it yet
  switch (w.deployment.state) {
    case "rolled_back":
      return "rolled_back";
    case "shipping":
      return "shipping";
    case "watching":
    case "live":
      return "live";
  }
  if (w.agentActivity === "needs_input") return "needs_you";
  if (w.agentActivity === "working") return "working";
  if (w.gate.state === "failed") return "gate_failed";
  if (w.gate.state === "passed") return "ready_to_ship";
  if (w.newFeedback > 0) return "client_feedback";
  if (w.shared) return "waiting_on_client";
  if (w.previewLive) return "preview_live";
  // The agent finished its turn and nothing runs yet: it waits for your next message.
  return "needs_you";
}

// Status changes worth a screen reader announcement (DESIGN_RULES.md §6).
const ANNOUNCE: Partial<Record<WorkerStatus, string>> = {
  needs_you: "needs you",
  client_feedback: "has new client feedback",
  gate_failed: "failed the gate",
  ready_to_ship: "is ready to ship",
  live: "is live",
  rolled_back: "was rolled back",
};

export function announcement(w: Worker, status: WorkerStatus): string | null {
  const phrase = ANNOUNCE[status];
  return phrase ? `${w.title} ${phrase}` : null;
}

/** Short cause shown next to statuses that need evidence (DESIGN_RULES.md §4). */
export function statusDetail(w: Worker, status: WorkerStatus): string | null {
  if (status === "gate_failed" && w.gate.state === "failed") return w.gate.cause;
  if (status === "rolled_back" && w.deployment.state === "rolled_back") return w.deployment.reason;
  if (status === "queued") return "Waiting for a free slot";
  if (status === "client_feedback") return w.newFeedback === 1 ? "1 new comment" : `${w.newFeedback} new comments`;
  return null;
}
