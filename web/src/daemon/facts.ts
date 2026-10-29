import type { WorkerStore } from "../workers/store";
import type { ApiProject, ApiWorker, ApiWorkspace, ServerMessage } from "./wire";
import { toWorker } from "./wire";

export interface ProjectOption {
  id: string;
  label: string;
}

export interface WorkspaceOption {
  slug: string;
  name: string;
}

export interface FactsState {
  /** True once the first snapshot arrived; before that the Board shows its loading state. */
  loaded: boolean;
  projects: ProjectOption[];
  workspaces: WorkspaceOption[];
}

/**
 * The daemon's facts as the Board sees them. Workers go to the worker store, which derives
 * their status; workspaces and projects stay here, since they only name things.
 */
export function createFacts(workers: WorkerStore) {
  let workspaces = new Map<string, ApiWorkspace>();
  let projects = new Map<string, ApiProject>();
  let apiWorkers = new Map<string, ApiWorker>();
  let state: FactsState = { loaded: false, projects: [], workspaces: [] };
  const listeners = new Set<() => void>();

  function label(projectId: string): string {
    const p = projects.get(projectId);
    if (!p) return "Unknown project";
    const ws = workspaces.get(p.workspaceId);
    return ws ? `${ws.slug} / ${p.name}` : p.name;
  }

  function options(): ProjectOption[] {
    return [...projects.values()].map((p) => ({ id: p.id, label: label(p.id) })).sort((a, b) => a.label.localeCompare(b.label));
  }

  function publish(loaded: boolean) {
    const named = [...workspaces.values()].map((w) => ({ slug: w.slug, name: w.name })).sort((a, b) => a.name.localeCompare(b.name));
    state = { loaded, projects: options(), workspaces: named };
    listeners.forEach((l) => l());
  }

  /** Names changed, so the cards of the affected projects change their label. */
  function relabel(affected: (w: ApiWorker) => boolean) {
    for (const w of apiWorkers.values()) if (affected(w)) workers.upsert(toWorker(w, label(w.projectId)));
  }

  return {
    subscribe(listener: () => void) {
      listeners.add(listener);
      return () => listeners.delete(listener);
    },
    getState: () => state,
    getProject: (id: string) => projects.get(id),

    apply(m: ServerMessage) {
      switch (m.type) {
        case "snapshot":
          workspaces = new Map(m.data.workspaces.map((w) => [w.id, w]));
          projects = new Map(m.data.projects.map((p) => [p.id, p]));
          apiWorkers = new Map(m.data.workers.map((w) => [w.id, w]));
          workers.replaceAll(m.data.workers.map((w) => toWorker(w, label(w.projectId))));
          publish(true);
          return;
        case "workspace":
          workspaces.set(m.data.id, m.data);
          relabel((w) => projects.get(w.projectId)?.workspaceId === m.data.id);
          break;
        case "project":
          projects.set(m.data.id, m.data);
          relabel((w) => w.projectId === m.data.id);
          break;
        case "worker":
          apiWorkers.set(m.data.id, m.data);
          workers.upsert(toWorker(m.data, label(m.data.projectId)));
          break;
      }
      publish(state.loaded);
    },
  };
}

export type Facts = ReturnType<typeof createFacts>;
