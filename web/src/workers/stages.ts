import type { Worker } from "./types";

// Loop strip stages (DESIGN_SYSTEM.md §8.9), derived from the same facts as status.
const STAGES = ["Prompt", "Worktree", "Preview", "Client", "Gate", "Live", "Watch"] as const;

export type Stage = (typeof STAGES)[number];
export type StageState = "done" | "running" | "pending" | "failed" | "skipped";

const CLIENT = STAGES.indexOf("Client");

function current(w: Worker): { index: number; state: StageState } {
  const at = (stage: Stage, state: StageState) => ({ index: STAGES.indexOf(stage), state });
  if (w.queued) return at("Worktree", "pending");
  switch (w.deployment.state) {
    case "rolled_back":
      return at("Watch", "failed");
    case "watching":
      return at("Watch", "running");
    case "live":
      return { index: STAGES.length, state: "done" };
    case "shipping":
      return at("Live", "running");
  }
  if (w.gate.state === "failed") return at("Gate", "failed");
  if (w.gate.state === "passed") return at("Live", "pending");
  if (w.clientReview && (w.shared || w.newFeedback > 0)) return at("Client", "running");
  if (w.previewLive) return at("Preview", "running");
  return at("Worktree", "running");
}

export function deriveStages(w: Worker): { stage: Stage; state: StageState }[] {
  const { index, state } = current(w);
  return STAGES.map((stage, i) => {
    if (i === CLIENT && !w.clientReview) return { stage, state: "skipped" };
    if (i < index) return { stage, state: "done" };
    if (i === index) return { stage, state };
    return { stage, state: "pending" };
  });
}

/** The stage a screen reader hears for the card; the mini strip itself is visual only. */
export function currentStageLabel(w: Worker): string {
  const stages = deriveStages(w);
  const active = stages.find((s) => s.state === "running" || s.state === "failed");
  if (active) return `${active.stage} ${active.state === "failed" ? "failed" : "in progress"}`;
  const next = stages.find((s) => s.state === "pending");
  return next ? `Next: ${next.stage}` : "All stages done";
}
