import type { ApiWorker } from "./wire";

export class ApiError extends Error {
  readonly status: number;
  readonly code: string;
  constructor(status: number, code: string, message: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
  }
}

interface NewWorker {
  title: string;
  agent: "claude" | "codex";
  clientReview: boolean;
}

interface ChecklistItem {
  id: string;
  title: string;
  detail: string;
  done: boolean;
  soon: boolean;
}

export interface Checklist {
  items: ChecklistItem[];
  percent: number;
  complete: boolean;
}

export interface Api {
  pair(pairingToken: string, deviceName: string): Promise<string>;
  /** Asks the daemon whether the token still works. */
  probe(): Promise<"ok" | "unauthorized" | "unreachable">;
  createWorker(projectId: string, input: NewWorker): Promise<ApiWorker>;
  startWorker(id: string): Promise<void>;
  interruptWorker(id: string): Promise<void>;
  checklist(): Promise<Checklist>;
  createWorkspace(slug: string, name: string): Promise<void>;
  createProject(workspace: string, name: string, repoUrl: string): Promise<void>;
}

type Fetch = typeof fetch;

/**
 * REST calls to the daemon, same origin as the page: Vite proxies /api in development and the
 * daemon serves the page itself in production. A 401 on an authenticated call means the
 * token was revoked, which `onUnauthorized` hears about.
 */
export function createApi(getToken: () => string | null, onUnauthorized: () => void, doFetch: Fetch = (...a) => fetch(...a)): Api {
  async function request(method: string, path: string, body?: unknown, authenticated = true): Promise<unknown> {
    const headers: Record<string, string> = {};
    if (body !== undefined) headers["Content-Type"] = "application/json";
    const token = getToken();
    if (authenticated && token) headers.Authorization = `Bearer ${token}`;
    let res: Response;
    try {
      res = await doFetch(`/api/v1${path}`, { method, headers, body: body === undefined ? undefined : JSON.stringify(body) });
    } catch {
      throw new ApiError(0, "unreachable", "Couldn't reach the daemon.");
    }
    if (res.status === 204) return null;
    let payload: unknown = null;
    try {
      payload = await res.json();
    } catch {
      // An answer without a JSON body: the status says enough.
    }
    if (res.ok) return payload;
    if (res.status === 401 && authenticated) onUnauthorized();
    const err = (payload as { error?: { code?: string; message?: string } } | null)?.error;
    throw new ApiError(res.status, err?.code ?? "error", err?.message ?? `The daemon answered ${res.status}.`);
  }

  return {
    async pair(pairingToken, deviceName) {
      const out = (await request("POST", "/pairing", { token: pairingToken, name: deviceName }, false)) as { token: string };
      return out.token;
    },
    async probe() {
      try {
        await request("GET", "/workspaces");
        return "ok";
      } catch (e) {
        return e instanceof ApiError && e.status === 401 ? "unauthorized" : "unreachable";
      }
    },
    async createWorker(projectId, input) {
      return (await request("POST", `/projects/${encodeURIComponent(projectId)}/workers`, input)) as ApiWorker;
    },
    async startWorker(id) {
      await request("POST", `/workers/${encodeURIComponent(id)}/start`);
    },
    async interruptWorker(id) {
      await request("POST", `/workers/${encodeURIComponent(id)}/interrupt`);
    },
    async checklist() {
      return (await request("GET", "/checklist")) as Checklist;
    },
    async createWorkspace(slug, name) {
      await request("POST", "/workspaces", { slug, name });
    },
    async createProject(workspace, name, repoUrl) {
      await request("POST", "/projects", { workspace, name, repoUrl });
    },
  };
}
