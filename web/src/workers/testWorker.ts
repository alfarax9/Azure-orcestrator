import type { Worker } from "./types";

export function testWorker(overrides: Partial<Worker> = {}): Worker {
  return {
    id: "w1",
    title: "Fix checkout total rounding",
    project: "warung-sari / web",
    branch: "fix/checkout-total",
    agent: "Claude Code",
    createdAt: 0,
    clientReview: true,
    agentActivity: "idle",
    previewLive: false,
    shared: false,
    newFeedback: 0,
    gate: { state: "none" },
    deployment: { state: "none" },
    added: 0,
    removed: 0,
    ...overrides,
  };
}
