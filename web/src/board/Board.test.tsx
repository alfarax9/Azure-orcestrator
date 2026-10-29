import { act, fireEvent, render, screen, within } from "@testing-library/react";
import { expect, test, vi } from "vitest";
import { createWorkerStore, WorkerStoreContext } from "../workers/store";
import { testWorker } from "../workers/testWorker";
import type { Worker } from "../workers/types";
import { Board, type NewTaskAction } from "./Board";

function show(workers: Worker[], props: { loading?: boolean; newTask?: NewTaskAction; onOpenWorker?: (id: string) => void } = {}) {
  const store = createWorkerStore(workers, (flush) => flush());
  render(
    <WorkerStoreContext value={store}>
      <Board newTask={props.newTask ?? { onPress: vi.fn() }} loading={props.loading} onOpenWorker={props.onOpenWorker} />
    </WorkerStoreContext>,
  );
  return store;
}

test("before the first facts arrive the columns keep their place, with no counts and no claims", () => {
  show([], { loading: true });
  expect(screen.getByRole("heading", { level: 1 }).textContent?.trim()).toBe("Board");
  expect(screen.queryByText(/No workers yet/)).toBeNull();
  for (const name of ["Working", "Needs you", "Review", "Ready to ship", "Live"]) {
    expect(screen.getByRole("heading", { name: new RegExp(`^${name}`) }).textContent).toContain("—");
  }
  expect(screen.queryAllByRole("grid")).toHaveLength(0);
});

test("with no workers the Board says so and keeps its frame", () => {
  show([]);
  expect(screen.getByText("No workers yet. Start a task to put an agent to work.")).toBeTruthy();
  expect(screen.getByRole("heading", { level: 1 }).textContent).toContain("0 need you");
  expect(screen.getAllByRole("grid")).toHaveLength(5);
});

test("New task works, or explains why it does not", () => {
  const onPress = vi.fn();
  const { unmount } = render(
    <WorkerStoreContext value={createWorkerStore([])}>
      <Board newTask={{ onPress }} />
    </WorkerStoreContext>,
  );
  fireEvent.click(screen.getByRole("button", { name: "New task" }));
  expect(onPress).toHaveBeenCalledTimes(1);
  unmount();

  show([], { newTask: { onPress, disabledReason: "Offline: reconnect to the daemon to start a task." } });
  expect(screen.getByText("Offline: reconnect to the daemon to start a task.")).toBeTruthy();
  expect(screen.getByRole("button", { name: "New task" }).hasAttribute("disabled")).toBe(true);
});

test("a card opens its worker, and a worker whose facts change moves to the column its status belongs to", () => {
  const onOpenWorker = vi.fn();
  const store = show([testWorker({ id: "a", title: "Fix header", agentActivity: "working" })], { onOpenWorker });
  expect(within(screen.getByRole("grid", { name: "Working" })).getAllByRole("row")).toHaveLength(1);
  fireEvent.click(screen.getByRole("row", { name: /Fix header/ }));
  expect(onOpenWorker).toHaveBeenCalledWith("a");

  act(() => store.update("a", (w) => ({ ...w, agentActivity: "needs_input" })));
  expect(within(screen.getByRole("grid", { name: "Needs you" })).getAllByRole("row")).toHaveLength(1);
  expect(screen.getByRole("heading", { level: 1 }).textContent).toContain("1 needs you");
});
