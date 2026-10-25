import { expect, test, vi } from "vitest";
import { createWorkerStore } from "./store";
import { testWorker } from "./testWorker";

function setup() {
  const flushes: (() => void)[] = [];
  const store = createWorkerStore(
    [testWorker({ id: "a" }), testWorker({ id: "b" }), testWorker({ id: "c" })],
    (flush) => flushes.push(flush),
  );
  return { store, flush: () => flushes.splice(0).forEach((f) => f()), scheduled: () => flushes.length };
}

test("many facts in one frame notify listeners once", () => {
  const { store, flush, scheduled } = setup();
  const listener = vi.fn();
  store.subscribe(listener);
  for (let i = 0; i < 100; i++) store.update("a", (w) => ({ ...w, added: w.added + 1 }));
  expect(scheduled()).toBe(1);
  flush();
  expect(listener).toHaveBeenCalledTimes(1);
  expect(store.getWorker("a")?.added).toBe(100);
});

test("statuses keep their identity until a status actually changes", () => {
  const { store } = setup();
  const before = store.getStatuses();
  store.update("a", (w) => ({ ...w, added: 5 }));
  expect(store.getStatuses()).toBe(before);
  store.update("a", (w) => ({ ...w, agentActivity: "working" }));
  expect(store.getStatuses()).not.toBe(before);
  expect(store.getStatuses().a).toBe("working");
});

test("reorder moves keys next to the target", () => {
  const { store } = setup();
  store.reorder(["c"], "a", "before");
  expect(store.getOrder()).toEqual(["c", "a", "b"]);
  store.reorder(["c"], "b", "after");
  expect(store.getOrder()).toEqual(["a", "b", "c"]);
});
