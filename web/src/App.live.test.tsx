import { act, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { afterEach, beforeEach, expect, test, vi } from "vitest";
import { App } from "./App";
import { clearToken, loadToken, saveToken } from "./daemon/token";
import type { ApiWorker } from "./daemon/wire";

// xterm.js needs a real layout engine; the terminal has its own tests.
vi.mock("./terminal/TerminalView", () => ({ TerminalView: () => <div>terminal</div> }));

class FakeWebSocket {
  static all: FakeWebSocket[] = [];
  onmessage: ((e: { data: unknown }) => void) | null = null;
  onclose: (() => void) | null = null;
  constructor(
    readonly url: string,
    readonly protocols: string[],
  ) {
    FakeWebSocket.all.push(this);
  }
  close() {}
  send(message: unknown) {
    this.onmessage?.({ data: JSON.stringify(message) });
  }
}

const worker = (id: string, extra: Partial<ApiWorker> = {}): ApiWorker => ({
  id,
  projectId: "p1",
  title: `Task ${id}`,
  branch: `astra/${id}`,
  agent: "claude",
  createdAt: Date.now(),
  clientReview: true,
  agentActivity: "working",
  previewLive: false,
  shared: false,
  newFeedback: 0,
  gate: { state: "none" },
  deployment: { state: "none" },
  added: 0,
  removed: 0,
  ...extra,
});

const facts = { workspaces: [{ id: "ws", slug: "acme", name: "Warung Sari", createdAt: 1 }], projects: [{ id: "p1", workspaceId: "ws", name: "site", defaultBranch: "main", createdAt: 1 }] };

let fetchCalls: { url: string; method: string }[] = [];
let answers: Record<string, Response> = {};

beforeEach(() => {
  FakeWebSocket.all = [];
  fetchCalls = [];
  answers = {};
  window.history.pushState({}, "", "/");
  clearToken();
  vi.stubGlobal("WebSocket", FakeWebSocket);
  vi.stubGlobal(
    "fetch",
    vi.fn(async (url: string, init?: RequestInit) => {
      fetchCalls.push({ url, method: init?.method ?? "GET" });
      return answers[`${init?.method ?? "GET"} ${url}`]?.clone() ?? new Response(JSON.stringify({ error: { code: "not_found", message: "x" } }), { status: 404 });
    }),
  );
});
afterEach(() => vi.unstubAllGlobals());

test("a browser that has not paired is asked to, and after pairing it shows the live Board", async () => {
  answers["POST /api/v1/pairing"] = new Response(JSON.stringify({ device: { id: "d" }, token: "device-token" }));
  render(<App />);
  expect(screen.getByRole("heading", { level: 1 }).textContent).toBe("Pair this device");

  fireEvent.change(screen.getByLabelText("Pairing token"), { target: { value: "one-time" } });
  fireEvent.click(screen.getByRole("button", { name: "Pair device" }));
  await screen.findByRole("heading", { level: 1, name: /Board/ });
  expect(loadToken()).toBe("device-token");
  expect(FakeWebSocket.all[0]?.url).toMatch(/^ws:\/\/.*\/api\/v1\/events$/);
  expect(FakeWebSocket.all[0]?.protocols).toEqual(["astra.v1", "bearer.device-token"]);
});

test("the Board is empty of claims until the snapshot, then follows every change", async () => {
  saveToken("tok");
  render(<App />);
  expect(screen.getByText("Connecting to the daemon…")).toBeTruthy();
  expect(screen.getByRole("heading", { level: 1 }).textContent?.trim()).toBe("Board");

  act(() => FakeWebSocket.all[0]?.send({ type: "snapshot", data: { ...facts, workers: [worker("a"), worker("b", { agentActivity: "needs_input" })] } }));
  expect(screen.getByText("Connected to the daemon")).toBeTruthy();
  expect(screen.getByRole("heading", { level: 1 }).textContent).toContain("1 needs you");
  expect(within(screen.getByRole("grid", { name: "Working" })).getAllByRole("row")).toHaveLength(1);
  expect(screen.getAllByText(/acme \/ site/, { selector: ".card-line" })).toHaveLength(2);

  act(() => FakeWebSocket.all[0]?.send({ type: "worker", data: worker("a", { agentActivity: "needs_input" }) }));
  await waitFor(() => expect(within(screen.getByRole("grid", { name: "Needs you" })).getAllByRole("row")).toHaveLength(2));
  expect(screen.getByRole("heading", { level: 1 }).textContent).toContain("2 need you");
});

test("opening a worker shows its screen, and the Board link comes back", async () => {
  saveToken("tok");
  render(<App />);
  act(() => FakeWebSocket.all[0]?.send({ type: "snapshot", data: { ...facts, workers: [worker("a")] } }));
  act(() => {
    window.location.hash = "#/workers/a";
    window.dispatchEvent(new HashChangeEvent("hashchange"));
  });
  expect((await screen.findByRole("heading", { level: 1 })).textContent).toBe("Task a");
  act(() => {
    window.location.hash = "#/";
    window.dispatchEvent(new HashChangeEvent("hashchange"));
  });
  expect((await screen.findByRole("heading", { level: 1 })).textContent).toContain("Board");
});

test("when the daemon drops, the data stays with its age and New task waits; a revoked token goes back to pairing", async () => {
  saveToken("tok");
  render(<App />);
  act(() => FakeWebSocket.all[0]?.send({ type: "snapshot", data: { ...facts, workers: [worker("a")] } }));

  answers["GET /api/v1/workspaces"] = new Response(null, { status: 502 });
  await act(async () => FakeWebSocket.all[0]?.onclose?.());
  expect(screen.getByText(/Offline · last updated/)).toBeTruthy();
  expect(screen.getByText("Task a")).toBeTruthy();
  expect(screen.getByRole("button", { name: "New task" }).hasAttribute("disabled")).toBe(true);

  answers["GET /api/v1/workspaces"] = new Response(JSON.stringify({ error: { code: "unauthorized", message: "x" } }), { status: 401 });
  vi.useFakeTimers();
  await act(async () => vi.advanceTimersByTimeAsync(1000));
  await act(async () => FakeWebSocket.all.at(-1)?.onclose?.());
  vi.useRealTimers();
  await waitFor(() => expect(screen.getByRole("heading", { level: 1 }).textContent).toBe("Pair this device"));
  expect(screen.getByText(/This device was disconnected from the daemon/)).toBeTruthy();
  expect(loadToken()).toBeNull();
});

test("starting a task opens its worker", async () => {
  saveToken("tok");
  answers["POST /api/v1/projects/p1/workers"] = new Response(JSON.stringify(worker("new1", { title: "Fix header" })), { status: 201 });
  answers["POST /api/v1/workers/new1/start"] = new Response(JSON.stringify(worker("new1")));
  render(<App />);
  act(() => FakeWebSocket.all[0]?.send({ type: "snapshot", data: { ...facts, workers: [] } }));
  fireEvent.click(screen.getByRole("button", { name: "New task" }));
  fireEvent.change(await screen.findByLabelText("What should the agent do?"), { target: { value: "Fix header" } });
  fireEvent.click(screen.getByRole("button", { name: "Start task" }));
  await waitFor(() => expect(window.location.hash).toBe("#/workers/new1"));
  expect(fetchCalls.map((c) => `${c.method} ${c.url}`).filter((c) => !c.includes("/checklist"))).toEqual(["POST /api/v1/projects/p1/workers", "POST /api/v1/workers/new1/start"]);
});
