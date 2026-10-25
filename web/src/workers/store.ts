import { createContext, useContext, useSyncExternalStore } from "react";
import { announcement, deriveStatus, type WorkerStatus } from "./status";
import type { Worker } from "./types";

type Listener = () => void;
type Schedule = (flush: () => void) => void;

// Many facts can arrive within one frame; listeners hear about them once per frame.
const nextFrame: Schedule = (flush) =>
  typeof requestAnimationFrame === "function" ? requestAnimationFrame(flush) : setTimeout(flush, 16);

export function createWorkerStore(initial: Worker[], schedule: Schedule = nextFrame) {
  let workers: Record<string, Worker> = {};
  let statuses: Record<string, WorkerStatus> = {};
  for (const w of initial) {
    workers[w.id] = w;
    statuses[w.id] = deriveStatus(w);
  }
  let order = initial.map((w) => w.id);
  let lastAnnouncement = "";
  const listeners = new Set<Listener>();
  let pending = false;

  function changed() {
    if (pending) return;
    pending = true;
    schedule(() => {
      pending = false;
      listeners.forEach((l) => l());
    });
  }

  function setWorker(next: Worker) {
    workers = { ...workers, [next.id]: next };
    const status = deriveStatus(next);
    // Keep the statuses object stable when no status moved, so the Board skips the render.
    if (statuses[next.id] !== status) {
      if (statuses[next.id]) lastAnnouncement = announcement(next, status) ?? lastAnnouncement;
      statuses = { ...statuses, [next.id]: status };
    }
    changed();
  }

  return {
    subscribe(listener: Listener) {
      listeners.add(listener);
      return () => listeners.delete(listener);
    },
    getWorker: (id: string) => workers[id],
    getStatuses: () => statuses,
    getOrder: () => order,
    getAnnouncement: () => lastAnnouncement,
    update(id: string, fn: (w: Worker) => Worker) {
      const w = workers[id];
      if (w) setWorker(fn(w));
    },
    add(w: Worker) {
      order = [w.id, ...order];
      setWorker(w);
    },
    /** Moves keys next to target, keeping their relative order (drag and drop reorder). */
    reorder(keys: string[], target: string, position: "before" | "after") {
      const moving = new Set(keys);
      const rest = order.filter((id) => !moving.has(id));
      const at = rest.indexOf(target);
      if (at === -1) return;
      rest.splice(position === "before" ? at : at + 1, 0, ...order.filter((id) => moving.has(id)));
      order = rest;
      changed();
    },
  };
}

export type WorkerStore = ReturnType<typeof createWorkerStore>;

export const WorkerStoreContext = createContext<WorkerStore | null>(null);

export function useWorkerStore(): WorkerStore {
  const store = useContext(WorkerStoreContext);
  if (!store) throw new Error("WorkerStoreContext is missing");
  return store;
}

export function useWorker(id: string): Worker | undefined {
  const store = useWorkerStore();
  return useSyncExternalStore(store.subscribe, () => store.getWorker(id));
}

export function useStatuses(): Record<string, WorkerStatus> {
  const store = useWorkerStore();
  return useSyncExternalStore(store.subscribe, store.getStatuses);
}

export function useOrder(): string[] {
  const store = useWorkerStore();
  return useSyncExternalStore(store.subscribe, store.getOrder);
}

export function useAnnouncement(): string {
  const store = useWorkerStore();
  return useSyncExternalStore(store.subscribe, store.getAnnouncement);
}
