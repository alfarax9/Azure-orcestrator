import { expect, test } from "vitest";
import { parseMessage, toWorker, type ApiWorker } from "./wire";

const apiWorker: ApiWorker = {
  id: "w1",
  projectId: "p1",
  title: "Fix the header",
  branch: "astra/fix-the-header-3a9f2c",
  agent: "claude",
  createdAt: 1_700_000_000_000,
  clientReview: true,
  agentActivity: "needs_input",
  previewLive: false,
  shared: false,
  newFeedback: 0,
  gate: { state: "none" },
  deployment: { state: "none" },
  added: 2,
  removed: 1,
};

test("a worker from the daemon becomes a Board worker with the agent's name", () => {
  expect(toWorker(apiWorker, "acme / site")).toEqual({
    id: "w1",
    title: "Fix the header",
    project: "acme / site",
    branch: "astra/fix-the-header-3a9f2c",
    agent: "Claude Code",
    createdAt: 1_700_000_000_000,
    clientReview: true,
    agentActivity: "needs_input",
    previewLive: false,
    shared: false,
    newFeedback: 0,
    gate: { state: "none" },
    deployment: { state: "none" },
    added: 2,
    removed: 1,
  });
  expect(toWorker({ ...apiWorker, agent: "codex" }, "x").agent).toBe("Codex");
});

test("a failed gate or a rollback always carries a reason to show", () => {
  const failed = toWorker({ ...apiWorker, gate: { state: "failed" } }, "x");
  expect(failed.gate).toEqual({ state: "failed", cause: "Gate failed" });
  const named = toWorker({ ...apiWorker, gate: { state: "failed", cause: "2 tests" } }, "x");
  expect(named.gate).toEqual({ state: "failed", cause: "2 tests" });
  const back = toWorker({ ...apiWorker, deployment: { state: "rolled_back", reason: "error rate 4.1%" } }, "x");
  expect(back.deployment).toEqual({ state: "rolled_back", reason: "error rate 4.1%" });
  expect(toWorker({ ...apiWorker, deployment: { state: "rolled_back" } }, "x").deployment).toEqual({ state: "rolled_back", reason: "Rolled back" });
});

test("only messages this client understands are read", () => {
  const snapshot = { type: "snapshot", data: { workspaces: [], projects: [], workers: [apiWorker] } };
  expect(parseMessage(JSON.stringify(snapshot))?.type).toBe("snapshot");
  expect(parseMessage(JSON.stringify({ type: "worker", data: apiWorker }))?.type).toBe("worker");
  expect(parseMessage(JSON.stringify({ type: "project", data: { id: "p1" } }))?.type).toBe("project");
  expect(parseMessage(JSON.stringify({ type: "workspace", data: { id: "ws" } }))?.type).toBe("workspace");
  for (const bad of [
    "not json",
    "[]",
    "null",
    "{}",
    JSON.stringify({ type: "worker" }),
    JSON.stringify({ type: "worker", data: {} }),
    JSON.stringify({ type: "snapshot", data: { workspaces: [], projects: [] } }),
    JSON.stringify({ type: "snapshot", data: { workspaces: [], projects: [], workers: [{}] } }),
    JSON.stringify({ type: "somethingNew", data: { id: "x" } }),
  ]) {
    expect(parseMessage(bad), bad).toBeNull();
  }
});
