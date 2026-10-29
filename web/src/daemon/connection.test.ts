import { afterEach, beforeEach, expect, test, vi } from "vitest";
import { connectEvents, protocols, type ConnectionState, type SocketLike } from "./connection";
import type { ServerMessage } from "./wire";

class FakeSocket implements SocketLike {
  onmessage: ((e: { data: unknown }) => void) | null = null;
  onclose: (() => void) | null = null;
  closed = false;
  constructor(
    readonly url: string,
    readonly protocols: string[],
  ) {}
  close() {
    this.closed = true;
  }
  send(data: unknown) {
    this.onmessage?.({ data: typeof data === "string" ? data : JSON.stringify(data) });
  }
  drop() {
    this.onclose?.();
  }
}

const snapshot = JSON.stringify({ type: "snapshot", data: { workspaces: [], projects: [], workers: [] } });

function setup(probe: () => Promise<"ok" | "unauthorized" | "unreachable"> = () => Promise.resolve("unreachable")) {
  const sockets: FakeSocket[] = [];
  const states: ConnectionState[] = [];
  const messages: ServerMessage[] = [];
  const stop = connectEvents({
    url: "ws://daemon/api/v1/events",
    token: "tok",
    probe,
    onMessage: (m) => messages.push(m),
    onState: (s) => states.push(s),
    createSocket: (url, p) => {
      const s = new FakeSocket(url, p);
      sockets.push(s);
      return s;
    },
    now: () => Date.now(),
  });
  return { sockets, states, messages, stop, last: () => states[states.length - 1] };
}

beforeEach(() => {
  vi.useFakeTimers();
  vi.setSystemTime(1_000_000);
});
afterEach(() => vi.useRealTimers());

test("the socket carries its version and token as subprotocols, and the first message makes it live", () => {
  const { sockets, messages, last } = setup();
  expect(sockets[0]?.protocols).toEqual(["astra.v1", "bearer.tok"]);
  expect(protocols("x")).toEqual(["astra.v1", "bearer.x"]);
  expect(last()?.phase).toBe("connecting");
  sockets[0]?.send(snapshot);
  expect(last()).toEqual({ phase: "live", lastUpdate: 1_000_000, retryAt: null });
  expect(messages).toHaveLength(1);
  sockets[0]?.send("not a message");
  sockets[0]?.send({ type: "worker" });
  expect(messages).toHaveLength(1);
});

test("a drop goes offline, keeps the time of the last data, and retries with a growing delay", async () => {
  const { sockets, last } = setup();
  sockets[0]?.send(snapshot);
  vi.setSystemTime(1_005_000);
  sockets[0]?.drop();
  await vi.advanceTimersByTimeAsync(0);
  expect(last()).toEqual({ phase: "offline", lastUpdate: 1_000_000, retryAt: 1_006_000 });

  await vi.advanceTimersByTimeAsync(1000);
  expect(sockets).toHaveLength(2);
  expect(last()?.phase).toBe("connecting");
  expect(last()?.lastUpdate).toBe(1_000_000);

  // Failing again without ever going live doubles the wait, up to a ceiling.
  const waits: number[] = [];
  for (let i = 0; i < 6; i++) {
    sockets[sockets.length - 1]?.drop();
    await vi.advanceTimersByTimeAsync(0);
    const s = last();
    waits.push((s?.retryAt ?? 0) - Date.now());
    await vi.advanceTimersByTimeAsync(s?.retryAt ? s.retryAt - Date.now() : 0);
  }
  expect(waits).toEqual([2000, 4000, 8000, 15000, 15000, 15000]);
});

test("a snapshot after a reconnect starts the delay over", async () => {
  const { sockets, last } = setup();
  sockets[0]?.drop();
  await vi.advanceTimersByTimeAsync(1000);
  sockets[1]?.drop();
  await vi.advanceTimersByTimeAsync(2000);
  sockets[2]?.send(snapshot);
  sockets[2]?.drop();
  await vi.advanceTimersByTimeAsync(0);
  expect((last()?.retryAt ?? 0) - Date.now()).toBe(1000);
});

test("a revoked token stops the retrying and says so", async () => {
  const { sockets, last } = setup(() => Promise.resolve("unauthorized"));
  sockets[0]?.drop();
  await vi.advanceTimersByTimeAsync(60_000);
  expect(last()?.phase).toBe("unauthorized");
  expect(sockets).toHaveLength(1);
});

test("closing it for good ends the socket, the retries and the messages", async () => {
  const { sockets, messages, stop, states } = setup();
  sockets[0]?.drop();
  await vi.advanceTimersByTimeAsync(0);
  stop();
  await vi.advanceTimersByTimeAsync(60_000);
  expect(sockets).toHaveLength(1);
  const before = states.length;
  sockets[0]?.send(snapshot);
  sockets[0]?.drop();
  expect(states).toHaveLength(before);
  expect(messages).toHaveLength(0);
  expect(sockets[0]?.closed).toBe(true);
});

test("a drop that arrives after the socket was replaced is ignored", async () => {
  const { sockets } = setup();
  sockets[0]?.drop();
  await vi.advanceTimersByTimeAsync(1000);
  sockets[0]?.drop(); // the old socket reporting late
  await vi.advanceTimersByTimeAsync(5000);
  expect(sockets).toHaveLength(2);
});
