import { describe, expect, test } from "vitest";
import { deriveStatus, STATUS, statusDetail } from "./status";
import { testWorker } from "./testWorker";

describe("deriveStatus", () => {
  test.each([
    ["agent writing", { agentActivity: "working" }, "working"],
    ["agent asks a question", { agentActivity: "needs_input", previewLive: true }, "needs_you"],
    ["agent idle, nothing running", {}, "needs_you"],
    ["preview up", { previewLive: true }, "preview_live"],
    ["link shared", { previewLive: true, shared: true }, "waiting_on_client"],
    ["new client comments", { previewLive: true, shared: true, newFeedback: 2 }, "client_feedback"],
    ["gate failed", { gate: { state: "failed", cause: "2 tests" } }, "gate_failed"],
    ["gate passed", { gate: { state: "passed" } }, "ready_to_ship"],
    ["deploying", { gate: { state: "passed" }, deployment: { state: "shipping" } }, "shipping"],
    ["watch window", { deployment: { state: "watching" } }, "live"],
    ["live", { deployment: { state: "live" } }, "live"],
    ["rolled back", { deployment: { state: "rolled_back", reason: "error rate 4.1%" } }, "rolled_back"],
  ] as const)("%s", (_, facts, expected) => {
    expect(deriveStatus(testWorker(facts))).toBe(expected);
  });

  test("a running agent outranks an older gate result", () => {
    expect(deriveStatus(testWorker({ agentActivity: "working", gate: { state: "failed", cause: "lint" } }))).toBe("working");
  });
});

describe("statusDetail", () => {
  test("names the cause next to the status", () => {
    expect(statusDetail(testWorker({ gate: { state: "failed", cause: "2 tests" } }), "gate_failed")).toBe("2 tests");
    expect(statusDetail(testWorker({ newFeedback: 1 }), "client_feedback")).toBe("1 new comment");
    expect(statusDetail(testWorker(), "working")).toBeNull();
  });
});

test("a queued worker is Queued, whatever else its facts say, and says what it waits for", () => {
  const w = testWorker({ queued: true, agentActivity: "working" });
  expect(deriveStatus(w)).toBe("queued");
  expect(STATUS.queued.column).toBe("working");
  expect(statusDetail(w, "queued")).toBe("Waiting for a free slot");
});
