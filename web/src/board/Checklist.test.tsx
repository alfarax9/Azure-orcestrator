import { act, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, expect, test, vi } from "vitest";
import type { Checklist as ChecklistData } from "../daemon/api";
import { Checklist } from "./Checklist";

const partial: ChecklistData = {
  percent: 33,
  complete: false,
  items: [
    { id: "server", title: "Connect a server", detail: "This server runs the daemon you are talking to.", done: true, soon: false },
    { id: "domain", title: "Connect a domain", detail: "Previews and production need a domain. This arrives with previews.", done: false, soon: true },
    { id: "agent", title: "Sign in to an agent", detail: "Open a worker's terminal and run claude or codex.", done: false, soon: false },
    { id: "project", title: "Add your first project", detail: "Clone a repository.", done: false, soon: false },
  ],
};

beforeEach(() => vi.useFakeTimers());
afterEach(() => vi.useRealTimers());

function setup(checklist: () => Promise<ChecklistData>, props: { live?: boolean; facts?: number } = {}) {
  const onAddProject = vi.fn();
  const api = { checklist: vi.fn(checklist) };
  const view = render(<Checklist api={api} live={props.live ?? true} facts={props.facts ?? 0} onAddProject={onAddProject} />);
  return { api, onAddProject, view };
}

test("each step says its state in words, and the percentage is the daemon's", async () => {
  setup(() => Promise.resolve(partial));
  await act(async () => vi.advanceTimersByTimeAsync(0));
  expect(screen.getByRole("heading", { name: "Get started" })).toBeTruthy();
  expect(screen.getByText("33% done")).toBeTruthy();
  const items = screen.getAllByRole("listitem").map((li) => li.textContent);
  expect(items[0]).toMatch(/^DoneConnect a server/);
  expect(items[1]).toMatch(/^SoonConnect a domain/);
  expect(items[2]).toMatch(/^To doSign in to an agent/);
});

test("the project step offers to add one, and only while it is not done", async () => {
  const { onAddProject, view } = setup(() => Promise.resolve(partial));
  await act(async () => vi.advanceTimersByTimeAsync(0));
  fireEvent.click(screen.getByRole("button", { name: "Add a project" }));
  expect(onAddProject).toHaveBeenCalled();
  view.unmount();

  const done = { ...partial, items: partial.items.map((i) => (i.id === "project" ? { ...i, done: true } : i)) };
  setup(() => Promise.resolve(done));
  await act(async () => vi.advanceTimersByTimeAsync(0));
  expect(screen.queryByRole("button", { name: "Add a project" })).toBeNull();
});

test("it goes away once everything that can be done is done", async () => {
  setup(() => Promise.resolve({ ...partial, percent: 100, complete: true }));
  await act(async () => vi.advanceTimersByTimeAsync(0));
  expect(screen.queryByRole("heading", { name: "Get started" })).toBeNull();
});

test("it says nothing when the daemon cannot say, and keeps what it showed when a later read fails", async () => {
  const unreachable = setup(() => Promise.reject(new Error("offline")));
  await act(async () => vi.advanceTimersByTimeAsync(0));
  expect(screen.queryByRole("heading", { name: "Get started" })).toBeNull();
  unreachable.view.unmount();

  let calls = 0;
  const { api } = setup(() => (++calls === 1 ? Promise.resolve(partial) : Promise.reject(new Error("offline"))));
  await act(async () => vi.advanceTimersByTimeAsync(0));
  await act(async () => vi.advanceTimersByTimeAsync(15_000));
  expect(api.checklist).toHaveBeenCalledTimes(2);
  expect(screen.getByText("33% done")).toBeTruthy();
});

test("it reads again every so often, when the projects change, and not while the daemon is away", async () => {
  const { api, view } = setup(() => Promise.resolve(partial));
  await act(async () => vi.advanceTimersByTimeAsync(0));
  await act(async () => vi.advanceTimersByTimeAsync(15_000));
  expect(api.checklist).toHaveBeenCalledTimes(2);
  view.rerender(<Checklist api={api} live facts={1} onAddProject={vi.fn()} />);
  await act(async () => vi.advanceTimersByTimeAsync(0));
  expect(api.checklist).toHaveBeenCalledTimes(3);
  view.rerender(<Checklist api={api} live={false} facts={1} onAddProject={vi.fn()} />);
  await act(async () => vi.advanceTimersByTimeAsync(60_000));
  expect(api.checklist).toHaveBeenCalledTimes(3);
});
