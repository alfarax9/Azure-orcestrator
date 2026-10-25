import { expect, test } from "vitest";
import { currentStageLabel, deriveStages } from "./stages";
import { testWorker } from "./testWorker";

const states = (w: ReturnType<typeof testWorker>) => deriveStages(w).map((s) => s.state);

test("a fresh worker is running in its worktree", () => {
  expect(states(testWorker())).toEqual(["done", "running", "pending", "pending", "pending", "pending", "pending"]);
});

test("internal projects skip the Client stage", () => {
  expect(states(testWorker({ clientReview: false, gate: { state: "passed" } }))).toEqual([
    "done", "done", "done", "skipped", "done", "pending", "pending",
  ]);
});

test("a failed gate marks the Gate stage failed", () => {
  const w = testWorker({ gate: { state: "failed", cause: "2 tests" } });
  expect(states(w)[4]).toBe("failed");
  expect(currentStageLabel(w)).toBe("Gate failed");
});

test("a rollback fails the Watch stage after everything else is done", () => {
  expect(states(testWorker({ deployment: { state: "rolled_back", reason: "error rate" } }))).toEqual([
    "done", "done", "done", "done", "done", "done", "failed",
  ]);
});

test("a live deployment past its watch window has every stage done", () => {
  const w = testWorker({ deployment: { state: "live" } });
  expect(states(w).every((s) => s === "done")).toBe(true);
  expect(currentStageLabel(w)).toBe("All stages done");
});
