import type { WorkerStore } from "../workers/store";
import type { Worker } from "../workers/types";

// Fake fact stream for ADR 0001 validation 3: N fact changes per second into the store.

/** Small seeded PRNG so runs are repeatable. */
function mulberry32(seed: number) {
  return () => {
    seed = (seed + 0x6d2b79f5) | 0;
    let t = Math.imul(seed ^ (seed >>> 15), 1 | seed);
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

/** Moves a worker one plausible step along Prompt → … → Watch. */
export function advance(w: Worker, random: () => number): Worker {
  switch (w.deployment.state) {
    case "rolled_back":
      return { ...w, agentActivity: "working", gate: { state: "none" }, deployment: { state: "none" } };
    case "shipping":
      return { ...w, deployment: { state: "watching" } };
    case "watching":
      return random() < 0.1
        ? { ...w, deployment: { state: "rolled_back", reason: "error rate 3.2% within 2 min of deploy" } }
        : { ...w, deployment: { state: "live" } };
    case "live":
      // Recycle the card as a fresh task so the stream never runs dry.
      return {
        ...w,
        createdAt: Date.now(),
        agentActivity: "working",
        previewLive: false,
        shared: false,
        newFeedback: 0,
        gate: { state: "none" },
        deployment: { state: "none" },
        added: 0,
        removed: 0,
      };
  }
  if (w.agentActivity !== "idle") return { ...w, agentActivity: "idle", previewLive: true };
  if (w.gate.state === "passed") return { ...w, deployment: { state: "shipping" } };
  if (w.gate.state === "failed") return { ...w, agentActivity: "working", gate: { state: "none" } };
  if (w.newFeedback > 0) return { ...w, agentActivity: "working", newFeedback: 0 };
  if (!w.previewLive) return { ...w, previewLive: true };
  if (w.clientReview && !w.shared) return { ...w, shared: true };
  if (w.clientReview && random() < 0.4) return { ...w, newFeedback: 1 + Math.floor(random() * 3) };
  return random() < 0.2 ? { ...w, gate: { state: "failed", cause: "1 test failing" } } : { ...w, gate: { state: "passed" } };
}

function mutate(w: Worker, random: () => number): Worker {
  const roll = random();
  // Most real facts only touch the card (diff growth); a fifth move it along the loop.
  if (roll < 0.7) return { ...w, added: w.added + 1 + Math.floor(random() * 20), removed: w.removed + Math.floor(random() * 6) };
  if (roll < 0.8) return { ...w, agentActivity: w.agentActivity === "working" ? "needs_input" : "working" };
  return advance(w, random);
}

export function startSimulation(store: WorkerStore, perSecond: number) {
  const random = mulberry32(1);
  const start = performance.now();
  let sent = 0;
  const timer = setInterval(() => {
    const due = Math.floor(((performance.now() - start) * perSecond) / 1000) - sent;
    const ids = store.getOrder();
    for (let i = 0; i < due; i++) {
      const id = ids[Math.floor(random() * ids.length)];
      if (id) store.update(id, (w) => mutate(w, random));
    }
    sent += due;
  }, 10);
  return {
    count: () => sent,
    stop: () => clearInterval(timer),
  };
}
