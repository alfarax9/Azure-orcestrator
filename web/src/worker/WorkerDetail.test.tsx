import { render, screen, within } from "@testing-library/react";
import { expect, test, vi } from "vitest";
import { DaemonContext, type Daemon } from "../daemon/DaemonContext";
import { createFacts } from "../daemon/facts";
import { createWorkerStore, WorkerStoreContext } from "../workers/store";
import { testWorker } from "../workers/testWorker";
import type { Worker } from "../workers/types";
import { LoopStrip } from "./LoopStrip";
import { WorkerDetail } from "./WorkerDetail";

vi.mock("../terminal/TerminalView", () => ({
  TerminalView: ({ workerId, title, token }: { workerId: string; title: string; token: string }) => (
    <section aria-label={`Terminal for ${title}`}>
      terminal of {workerId} with {token}
    </section>
  ),
}));

function show(workers: Worker[], loaded: boolean, id = "w1") {
  const store = createWorkerStore([], (flush) => flush());
  const facts = createFacts(store);
  if (loaded) facts.apply({ type: "snapshot", data: { workspaces: [], projects: [], workers: [] } });
  workers.forEach((w) => store.add(w));
  const daemon: Daemon = {
    api: { pair: vi.fn(), probe: vi.fn(), createWorker: vi.fn(), startWorker: vi.fn(), interruptWorker: vi.fn(), checklist: vi.fn(), createWorkspace: vi.fn(), createProject: vi.fn() },
    token: "tok",
    connection: { phase: "live", lastUpdate: 1, retryAt: null },
    facts,
    signOut: vi.fn(),
  };
  render(
    <DaemonContext value={daemon}>
      <WorkerStoreContext value={store}>
        <WorkerDetail id={id} />
      </WorkerStoreContext>
    </DaemonContext>,
  );
}

test("the worker's screen answers where the task stands and what comes next", async () => {
  show([testWorker({ agentActivity: "needs_input", added: 12, removed: 3 })], true);
  expect(screen.getByRole("heading", { level: 1 }).textContent).toBe("Fix checkout total rounding");
  expect(screen.getByText("Needs you")).toBeTruthy();
  expect(screen.getByText("The agent is waiting for your answer. Reply in the terminal.")).toBeTruthy();
  expect(screen.getByText("fix/checkout-total")).toBeTruthy();
  expect(screen.getByText("+12 −3")).toBeTruthy();
  expect(screen.getByRole("link", { name: "Board" }).getAttribute("href")).toBe("#/");
  expect(screen.getByRole("tab", { name: "Terminal" })).toBeTruthy();
  expect((await screen.findByRole("region", { name: "Terminal for Fix checkout total rounding" })).textContent).toBe("terminal of w1 with tok");
});

test("an idle agent is waiting for the next message, and a working one needs no sentence", () => {
  show([testWorker({ agentActivity: "idle" })], true);
  expect(screen.getByText("The agent is waiting for your next message. Type it in the terminal.")).toBeTruthy();
});

test("a working agent needs no sentence", () => {
  show([testWorker({ agentActivity: "working" })], true);
  expect(screen.queryByText(/The agent is waiting/)).toBeNull();
  expect(screen.getByText("Working")).toBeTruthy();
});

test("a failed gate shows its cause next to the status", () => {
  show([testWorker({ gate: { state: "failed", cause: "2 tests" } })], true);
  expect(screen.getByText("Gate failed")).toBeTruthy();
  expect(screen.getAllByText(/2 tests/).length).toBeGreaterThan(0);
});

test("before the facts arrive the title has a place; after, an unknown worker says so", () => {
  show([], false);
  expect(screen.queryByRole("heading", { level: 1 })).toBeNull();
  expect(screen.getByRole("link", { name: "Board" })).toBeTruthy();
});

test("an unknown worker says so once the facts are in", () => {
  show([], true, "gone");
  expect(screen.getByRole("heading", { level: 1 }).textContent).toBe("Worker not found");
});

test("the loop strip says each stage's state in words, and why one failed", () => {
  render(<LoopStrip worker={testWorker({ clientReview: false, gate: { state: "failed", cause: "2 tests" } })} />);
  const list = screen.getByRole("list", { name: "Loop" });
  const stages = within(list).getAllByRole("listitem");
  expect(stages.map((s) => s.textContent)).toEqual([
    "Prompt (done)",
    "Worktree (done)",
    "Preview (done)",
    "Client (skipped)",
    "Gate: 2 tests (failed)",
    "Live (not started)",
    "Watch (not started)",
  ]);
  expect(stages.filter((s) => s.getAttribute("aria-current") === "step")).toHaveLength(0);
});

test("the running stage is marked as the current step", () => {
  render(<LoopStrip worker={testWorker({ previewLive: true })} />);
  const current = screen.getAllByRole("listitem").filter((s) => s.getAttribute("aria-current") === "step");
  expect(current.map((s) => s.textContent)).toEqual(["Preview (in progress)"]);
});

test("a rolled back deployment shows its reason on the Watch stage", () => {
  render(<LoopStrip worker={testWorker({ deployment: { state: "rolled_back", reason: "error rate 4.1%" } })} />);
  expect(screen.getByText(/Watch: error rate 4.1%/)).toBeTruthy();
});
