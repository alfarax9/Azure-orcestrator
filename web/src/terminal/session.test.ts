import { afterEach, beforeEach, expect, test, vi } from "vitest";
import { connectTerminal, type TerminalSession, type TerminalState } from "./session";

class FakeSocket {
  binaryType = "blob";
  readyState = 0;
  onopen: (() => void) | null = null;
  onmessage: ((e: { data: unknown }) => void) | null = null;
  onclose: (() => void) | null = null;
  sent: (string | Uint8Array)[] = [];
  closed = false;
  constructor(
    readonly url: string,
    readonly protocols: string[],
  ) {}
  send(data: string | Uint8Array) {
    this.sent.push(data);
  }
  close() {
    this.closed = true;
  }
  open() {
    this.readyState = 1;
    this.onopen?.();
  }
  drop() {
    this.readyState = 3;
    this.onclose?.();
  }
}

function setup() {
  const sockets: FakeSocket[] = [];
  const states: TerminalState[] = [];
  const output: Uint8Array[] = [];
  let resets = 0;
  const session: TerminalSession = connectTerminal({
    url: "ws://daemon/api/v1/sessions/w1/pty",
    token: "tok",
    onOutput: (b) => output.push(b),
    onReset: () => resets++,
    onState: (s) => states.push(s),
    createSocket: (url, protocols) => {
      const s = new FakeSocket(url, protocols);
      sockets.push(s);
      return s;
    },
  });
  return { session, sockets, states, output, resets: () => resets, phase: () => states[states.length - 1]?.phase };
}

beforeEach(() => vi.useFakeTimers());
afterEach(() => vi.useRealTimers());

test("it connects with the token in the subprotocol, resets the screen on open, and shows output", () => {
  const { sockets, phase, resets, output } = setup();
  expect(sockets[0]?.protocols).toEqual(["astra.v1", "bearer.tok"]);
  expect(sockets[0]?.binaryType).toBe("arraybuffer");
  expect(phase()).toBe("connecting");
  sockets[0]?.open();
  expect(phase()).toBe("open");
  expect(resets()).toBe(1);
  sockets[0]?.onmessage?.({ data: new Uint8Array([104, 105]).buffer });
  expect(Array.from(output[0] ?? [])).toEqual([104, 105]);
});

test("typing and resizing reach the agent only while connected, and the size is repeated on every connection", () => {
  const { session, sockets } = setup();
  session.send(new Uint8Array([1]));
  session.resize(30, 100);
  expect(sockets[0]?.sent).toEqual([]);
  sockets[0]?.open();
  expect(sockets[0]?.sent).toEqual([JSON.stringify({ type: "resize", rows: 30, cols: 100 })]);
  session.send(new Uint8Array([108, 115, 13]));
  session.resize(40, 120);
  expect(sockets[0]?.sent.slice(1)).toEqual([new Uint8Array([108, 115, 13]), JSON.stringify({ type: "resize", rows: 40, cols: 120 })]);
});

test("a connection that never opens means there is no session: it stops and waits for the person", async () => {
  const { sockets, phase, session } = setup();
  sockets[0]?.drop();
  expect(phase()).toBe("unavailable");
  await vi.advanceTimersByTimeAsync(60_000);
  expect(sockets).toHaveLength(1);

  session.reconnect();
  expect(sockets).toHaveLength(2);
  expect(phase()).toBe("connecting");
  sockets[1]?.open();
  expect(phase()).toBe("open");
});

test("a connection that drops after it was open is retried with a growing delay, and starts clean each time", async () => {
  const { sockets, phase, resets } = setup();
  sockets[0]?.open();
  sockets[0]?.drop();
  expect(phase()).toBe("reconnecting");
  await vi.advanceTimersByTimeAsync(999);
  expect(sockets).toHaveLength(1);
  await vi.advanceTimersByTimeAsync(1);
  expect(sockets).toHaveLength(2);

  sockets[1]?.drop(); // fails before opening again: still retried, since it was open before
  await vi.advanceTimersByTimeAsync(2000);
  expect(sockets).toHaveLength(3);
  sockets[2]?.open();
  expect(resets()).toBe(2);
  expect(phase()).toBe("open");
});

test("the agent's exit is final: no retry, and the code is kept", async () => {
  const { sockets, states, phase } = setup();
  sockets[0]?.open();
  sockets[0]?.onmessage?.({ data: JSON.stringify({ type: "exit", code: 3 }) });
  sockets[0]?.drop();
  await vi.advanceTimersByTimeAsync(60_000);
  expect(sockets).toHaveLength(1);
  expect(phase()).toBe("exited");
  expect(states[states.length - 1]).toEqual({ phase: "exited", exitCode: 3 });
  sockets[0]?.onmessage?.({ data: "not json" });
  sockets[0]?.onmessage?.({ data: JSON.stringify({ type: "other" }) });
  expect(phase()).toBe("exited");
});

test("closing ends the socket and every retry", async () => {
  const { session, sockets } = setup();
  sockets[0]?.open();
  sockets[0]?.drop();
  session.close();
  await vi.advanceTimersByTimeAsync(60_000);
  expect(sockets).toHaveLength(1);
  session.reconnect();
  expect(sockets).toHaveLength(1);
});
