import { expect, test, vi } from "vitest";
import { ApiError, createApi } from "./api";
import { clearToken, loadToken, saveToken } from "./token";

function reply(status: number, body?: unknown): Response {
  return new Response(body === undefined ? null : JSON.stringify(body), { status });
}

function setup(answer: () => Response | Promise<Response>, token: string | null = "tok") {
  const fetch = vi.fn<typeof globalThis.fetch>(async () => answer());
  const onUnauthorized = vi.fn();
  const api = createApi(() => token, onUnauthorized, fetch);
  return { api, fetch, onUnauthorized };
}

function request(fetch: ReturnType<typeof setup>["fetch"]) {
  const [url, init] = fetch.mock.calls[0] ?? [];
  return { url: String(url), init: init as RequestInit };
}

test("pairing sends the pairing token without credentials and returns the device token", async () => {
  const { api, fetch } = setup(() => reply(200, { device: { id: "d" }, token: "device-token" }), "old");
  await expect(api.pair("pairing", "Laptop")).resolves.toBe("device-token");
  const { url, init } = request(fetch);
  expect(url).toBe("/api/v1/pairing");
  expect(init.method).toBe("POST");
  expect(init.body).toBe(JSON.stringify({ token: "pairing", name: "Laptop" }));
  expect((init.headers as Record<string, string>).Authorization).toBeUndefined();
});

test("authenticated calls carry the bearer token and quote ids in the path", async () => {
  const { api, fetch } = setup(() => reply(200, { id: "w1" }));
  await api.createWorker("p/1", { title: "Fix", agent: "codex", clientReview: false });
  const { url, init } = request(fetch);
  expect(url).toBe("/api/v1/projects/p%2F1/workers");
  expect((init.headers as Record<string, string>).Authorization).toBe("Bearer tok");
  expect(JSON.parse(String(init.body))).toEqual({ title: "Fix", agent: "codex", clientReview: false });
});

test("starting a worker needs no body and accepts the daemon's answer", async () => {
  const { api, fetch } = setup(() => reply(200, { id: "w1" }));
  await expect(api.startWorker("w1")).resolves.toBeUndefined();
  expect(request(fetch).url).toBe("/api/v1/workers/w1/start");
  expect(request(fetch).init.body).toBeUndefined();
});

test("failures carry the daemon's code and sentence", async () => {
  const { api } = setup(() => reply(409, { error: { code: "no_session", message: "the worker's agent is not running" } }));
  const err = await api.startWorker("w1").catch((e: unknown) => e);
  expect(err).toBeInstanceOf(ApiError);
  expect(err).toMatchObject({ status: 409, code: "no_session", message: "the worker's agent is not running" });
});

test("an answer that is not JSON still reports its status", async () => {
  const { api } = setup(() => new Response("<html>bad gateway</html>", { status: 502 }));
  await expect(api.startWorker("w1")).rejects.toMatchObject({ status: 502, code: "error", message: "The daemon answered 502." });
});

test("an unreachable daemon is its own failure", async () => {
  const { api } = setup(() => Promise.reject(new TypeError("network")));
  await expect(api.startWorker("w1")).rejects.toMatchObject({ status: 0, code: "unreachable" });
});

test("a 401 on an authenticated call means the token was revoked; on pairing it only means a wrong token", async () => {
  const revoked = setup(() => reply(401, { error: { code: "unauthorized", message: "pair this device first" } }));
  await expect(revoked.api.startWorker("w1")).rejects.toMatchObject({ status: 401 });
  expect(revoked.onUnauthorized).toHaveBeenCalledTimes(1);

  const wrong = setup(() => reply(401, { error: { code: "unauthorized", message: "pair this device first" } }));
  await expect(wrong.api.pair("guess", "Laptop")).rejects.toMatchObject({ status: 401 });
  expect(wrong.onUnauthorized).not.toHaveBeenCalled();
});

test("probe tells a working token from a revoked one and from no daemon", async () => {
  expect(await setup(() => reply(200, [])).api.probe()).toBe("ok");
  expect(await setup(() => reply(401, { error: { code: "unauthorized", message: "x" } })).api.probe()).toBe("unauthorized");
  expect(await setup(() => Promise.reject(new TypeError("network"))).api.probe()).toBe("unreachable");
  expect(await setup(() => reply(500, { error: { code: "internal", message: "x" } })).api.probe()).toBe("unreachable");
});

test("the checklist and the first workspace and project go through their routes", async () => {
  const { api, fetch } = setup(() => reply(200, { items: [], percent: 50, complete: false }));
  await expect(api.checklist()).resolves.toMatchObject({ percent: 50 });
  expect(request(fetch).url).toBe("/api/v1/checklist");

  const created = setup(() => reply(201, {}));
  await created.api.createWorkspace("warung-sari", "Warung Sari");
  expect(JSON.parse(String(request(created.fetch).init.body))).toEqual({ slug: "warung-sari", name: "Warung Sari" });
  const project = setup(() => reply(201, {}));
  await project.api.createProject("warung-sari", "web", "https://git.warung-sari.id/web.git");
  expect(request(project.fetch).url).toBe("/api/v1/projects");
  expect(JSON.parse(String(request(project.fetch).init.body))).toEqual({ workspace: "warung-sari", name: "web", repoUrl: "https://git.warung-sari.id/web.git" });
});

test("the device token survives in storage until it is cleared", () => {
  clearToken();
  expect(loadToken()).toBeNull();
  saveToken("secret");
  expect(loadToken()).toBe("secret");
  clearToken();
  expect(loadToken()).toBeNull();
});

test("a browser that blocks storage still keeps the token for the tab", () => {
  const blocked = vi.spyOn(Storage.prototype, "getItem").mockImplementation(() => {
    throw new Error("blocked");
  });
  const set = vi.spyOn(Storage.prototype, "setItem").mockImplementation(() => {
    throw new Error("blocked");
  });
  const remove = vi.spyOn(Storage.prototype, "removeItem").mockImplementation(() => {
    throw new Error("blocked");
  });
  saveToken("kept");
  expect(loadToken()).toBe("kept");
  clearToken();
  expect(loadToken()).toBeNull();
  blocked.mockRestore();
  set.mockRestore();
  remove.mockRestore();
});
