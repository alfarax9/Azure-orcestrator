import { expect, test } from "vitest";
import { createWorkerStore } from "../workers/store";
import { createFacts } from "./facts";
import type { ApiProject, ApiWorker, ApiWorkspace } from "./wire";

const ws: ApiWorkspace = { id: "ws1", slug: "acme", name: "Warung Sari", createdAt: 1 };
const project: ApiProject = { id: "p1", workspaceId: "ws1", name: "site", defaultBranch: "main", createdAt: 2 };
const worker = (id: string, extra: Partial<ApiWorker> = {}): ApiWorker => ({
  id,
  projectId: "p1",
  title: `Task ${id}`,
  branch: `astra/${id}`,
  agent: "claude",
  createdAt: 100,
  clientReview: true,
  agentActivity: "idle",
  previewLive: false,
  shared: false,
  newFeedback: 0,
  gate: { state: "none" },
  deployment: { state: "none" },
  added: 0,
  removed: 0,
  ...extra,
});

function setup() {
  const store = createWorkerStore([], (flush) => flush());
  return { store, facts: createFacts(store) };
}

test("nothing is loaded until the first snapshot, which fills the Board", () => {
  const { store, facts } = setup();
  expect(facts.getState()).toEqual({ loaded: false, projects: [], workspaces: [] });
  facts.apply({ type: "snapshot", data: { workspaces: [ws], projects: [project], workers: [worker("a"), worker("b", { createdAt: 200 })] } });
  expect(facts.getState()).toEqual({
    loaded: true,
    projects: [{ id: "p1", label: "acme / site" }],
    workspaces: [{ slug: "acme", name: "Warung Sari" }],
  });
  expect(store.getOrder()).toEqual(["b", "a"]);
  expect(store.getWorker("a")?.project).toBe("acme / site");
  expect(store.getStatuses()).toEqual({ a: "needs_you", b: "needs_you" });
});

test("changes update, add and relabel workers", () => {
  const { store, facts } = setup();
  facts.apply({ type: "snapshot", data: { workspaces: [ws], projects: [project], workers: [worker("a")] } });

  facts.apply({ type: "worker", data: worker("a", { agentActivity: "working", added: 3 }) });
  expect(store.getStatuses().a).toBe("working");
  expect(store.getWorker("a")?.added).toBe(3);

  facts.apply({ type: "worker", data: worker("c", { createdAt: 300 }) });
  expect(store.getOrder()[0]).toBe("c");

  facts.apply({ type: "workspace", data: { ...ws, slug: "acme-co" } });
  expect(store.getWorker("a")?.project).toBe("acme-co / site");
  facts.apply({ type: "project", data: { ...project, name: "web" } });
  expect(store.getWorker("c")?.project).toBe("acme-co / web");
  expect(facts.getState().projects).toEqual([{ id: "p1", label: "acme-co / web" }]);
  expect(facts.getProject("p1")?.name).toBe("web");
});

test("a worker that arrives before its project is named once the project arrives", () => {
  const { store, facts } = setup();
  facts.apply({ type: "snapshot", data: { workspaces: [], projects: [], workers: [] } });
  facts.apply({ type: "worker", data: worker("a") });
  expect(store.getWorker("a")?.project).toBe("Unknown project");
  facts.apply({ type: "workspace", data: ws });
  facts.apply({ type: "project", data: project });
  expect(store.getWorker("a")?.project).toBe("acme / site");
});

test("a project without its workspace is shown by its own name", () => {
  const { facts } = setup();
  facts.apply({ type: "project", data: project });
  expect(facts.getState().projects).toEqual([{ id: "p1", label: "site" }]);
});

test("listeners hear about changes until they leave", () => {
  const { facts } = setup();
  let heard = 0;
  const leave = facts.subscribe(() => heard++);
  facts.apply({ type: "workspace", data: ws });
  leave();
  facts.apply({ type: "workspace", data: ws });
  expect(heard).toBe(1);
});
