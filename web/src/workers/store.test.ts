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

test("status changes and explicit messages both reach the announcement", () => {
  const { store } = setup();
  store.update("a", (w) => ({ ...w, gate: { state: "passed" } }));
  expect(store.getAnnouncement()).toBe("Fix checkout total rounding is ready to ship");
  store.announce("Fix checkout total rounding moved to position 2 of 3 in Working");
  expect(store.getAnnouncement()).toBe("Fix checkout total rounding moved to position 2 of 3 in Working");
});

test("reorder moves keys next to the target", () => {
  const { store } = setup();
  store.reorder(["c"], "a", "before");
  expect(store.getOrder()).toEqual(["c", "a", "b"]);
  store.reorder(["c"], "b", "after");
  expect(store.getOrder()).toEqual(["a", "b", "c"]);
});

test("a snapshot keeps the order a person gave, puts new workers first, and drops the ones that are gone", () => {
  const { store } = setup();
  store.reorder(["c"], "a", "before"); // c, a, b
  store.replaceAll([testWorker({ id: "a" }), testWorker({ id: "b" }), testWorker({ id: "d", createdAt: 5 }), testWorker({ id: "e", createdAt: 9 })]);
  expect(store.getOrder()).toEqual(["e", "d", "a", "b"]);
  expect(store.getWorker("c")).toBeUndefined();
});

test("a snapshot leaves unchanged workers and statuses alone, and announces real status changes", () => {
  const { store } = setup();
  const a = store.getWorker("a");
  const statuses = store.getStatuses();
  store.replaceAll([testWorker({ id: "a" }), testWorker({ id: "b" }), testWorker({ id: "c" })]);
  expect(store.getWorker("a")).toBe(a);
  expect(store.getStatuses()).toBe(statuses);

  store.replaceAll([testWorker({ id: "a", agentActivity: "working" }), testWorker({ id: "b" }), testWorker({ id: "c", gate: { state: "passed" } })]);
  expect(store.getStatuses()).toMatchObject({ a: "working", c: "ready_to_ship" });
  expect(store.getAnnouncement()).toBe("Fix checkout total rounding is ready to ship");
});

test("upsert adds a new worker on top and replaces a known one only when it changed", () => {
  const { store, flush } = setup();
  const listener = vi.fn();
  store.subscribe(listener);
  store.upsert(testWorker({ id: "z" }));
  flush();
  expect(store.getOrder()[0]).toBe("z");
  store.upsert(testWorker({ id: "z" }));
  flush();
  expect(listener).toHaveBeenCalledTimes(1);
  store.upsert(testWorker({ id: "z", added: 4 }));
  flush();
  expect(listener).toHaveBeenCalledTimes(2);
  expect(store.getWorker("z")?.added).toBe(4);
});
