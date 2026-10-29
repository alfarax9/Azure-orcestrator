import { act, fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, expect, test, vi } from "vitest";
import { TerminalView, terminalUrl } from "./TerminalView";

// xterm.js needs a real layout engine; the view's own behaviour is what is tested here.
const { instances, FakeTerminal } = vi.hoisted(() => {
  const instances: InstanceType<typeof FakeTerminal>[] = [];
  class FakeTerminal {
    rows = 24;
    cols = 80;
    written: Uint8Array[] = [];
    resets = 0;
    focused = false;
    disposed = false;
    dataHandlers: ((d: string) => void)[] = [];
    keyHandler: ((e: KeyboardEvent) => boolean) | null = null;
    constructor(readonly options: Record<string, unknown>) {
      instances.push(this);
    }
    loadAddon() {}
    open() {}
    write(b: Uint8Array) {
      this.written.push(b);
    }
    reset() {
      this.resets++;
    }
    focus() {
      this.focused = true;
    }
    dispose() {
      this.disposed = true;
    }
    onData(h: (d: string) => void) {
      this.dataHandlers.push(h);
      return { dispose: () => undefined };
    }
    onBinary() {
      return { dispose: () => undefined };
    }
    attachCustomKeyEventHandler(h: (e: KeyboardEvent) => boolean) {
      this.keyHandler = h;
    }
  }
  return { instances, FakeTerminal };
});

vi.mock("@xterm/xterm", () => ({ Terminal: FakeTerminal }));
vi.mock("@xterm/xterm/css/xterm.css", () => ({}));
vi.mock("@xterm/addon-fit", () => ({ FitAddon: class { fit() {} } }));

class FakeSocket {
  static all: FakeSocket[] = [];
  binaryType = "";
  readyState = 0;
  onopen: (() => void) | null = null;
  onmessage: ((e: { data: unknown }) => void) | null = null;
  onclose: (() => void) | null = null;
  sent: unknown[] = [];
  constructor(
    readonly url: string,
    readonly protocols: string[],
  ) {
    FakeSocket.all.push(this);
  }
  send(d: unknown) {
    this.sent.push(d);
  }
  close() {}
  open() {
    this.readyState = 1;
    this.onopen?.();
  }
  drop() {
    this.onclose?.();
  }
}

beforeEach(() => {
  instances.length = 0;
  FakeSocket.all = [];
  vi.stubGlobal("ResizeObserver", class { observe() {} disconnect() {} });
});

function setup(api = { startWorker: vi.fn(() => Promise.resolve()), interruptWorker: vi.fn(() => Promise.resolve()) }) {
  const view = render(<TerminalView workerId="w1" title="Fix header" agent="Claude Code" token="tok" api={api} createSocket={(u, p) => new FakeSocket(u, p)} />);
  return { api, view, socket: () => FakeSocket.all[FakeSocket.all.length - 1] as FakeSocket, term: () => instances[0] as InstanceType<typeof FakeTerminal> };
}

test("the terminal is the agent's: output goes in, typing goes out, and the header says it is connected", () => {
  const { socket, term } = setup();
  expect(socket().url).toBe(terminalUrl("w1"));
  expect(screen.getByText("Connecting")).toBeTruthy();
  act(() => socket().open());
  expect(screen.getByText("Connected")).toBeTruthy();
  expect(term().resets).toBe(1);
  act(() => socket().onmessage?.({ data: new Uint8Array([65]).buffer }));
  expect(Array.from(term().written[0] ?? [])).toEqual([65]);
  term().dataHandlers[0]?.("ls\r");
  expect(socket().sent.at(-1)).toEqual(new TextEncoder().encode("ls\r"));
  expect(screen.getByRole("region", { name: "Terminal for Fix header" })).toBeTruthy();
});

test("Ctrl+J focuses the terminal, and Shift+Esc hands focus to the header, where Tab reaches Interrupt", () => {
  const { term } = setup();
  fireEvent.keyDown(document, { key: "j", ctrlKey: true });
  expect(term().focused).toBe(true);
  term().focused = false;
  fireEvent.keyDown(document, { key: "j" });
  fireEvent.keyDown(document, { key: "j", ctrlKey: true, shiftKey: true });
  expect(term().focused).toBe(false);

  const leave = new KeyboardEvent("keydown", { key: "Escape", shiftKey: true });
  expect(term().keyHandler?.(leave)).toBe(false);
  expect(document.activeElement?.classList.contains("terminal-header")).toBe(true);
  expect(term().keyHandler?.(new KeyboardEvent("keydown", { key: "a" }))).toBe(true);
});

test("Interrupt is available while connected and reports a failure", async () => {
  const interruptWorker = vi.fn(() => Promise.reject(new Error("the worker's agent is not running")));
  const { socket } = setup({ startWorker: vi.fn(() => Promise.resolve()), interruptWorker });
  expect(screen.getByRole("button", { name: "Interrupt" }).hasAttribute("disabled")).toBe(true);
  act(() => socket().open());
  await act(async () => fireEvent.click(screen.getByRole("button", { name: "Interrupt" })));
  expect(interruptWorker).toHaveBeenCalledWith("w1");
  expect((await screen.findByRole("alert")).textContent).toBe("the worker's agent is not running");
});

test("when there is no session the person can start the agent, and the terminal connects again", async () => {
  const { api, socket } = setup();
  act(() => socket().drop());
  expect(screen.getByText("Not running")).toBeTruthy();
  expect(screen.getByText("The agent isn't running, so there is no terminal to show.")).toBeTruthy();
  const before = FakeSocket.all.length;
  await act(async () => fireEvent.click(screen.getByRole("button", { name: "Start agent" })));
  expect(api.startWorker).toHaveBeenCalledWith("w1");
  expect(FakeSocket.all.length).toBe(before + 1);
  act(() => socket().open());
  expect(screen.getByText("Connected")).toBeTruthy();
});

test("a failed start says why, and Retry connects without starting", async () => {
  const startWorker = vi.fn(() => Promise.reject(new Error("unit exists")));
  const { socket } = setup({ startWorker, interruptWorker: vi.fn(() => Promise.resolve()) });
  act(() => socket().drop());
  await act(async () => fireEvent.click(screen.getByRole("button", { name: "Start agent" })));
  expect((await screen.findByRole("alert")).textContent).toBe("unit exists");
  const before = FakeSocket.all.length;
  fireEvent.click(screen.getByRole("button", { name: "Retry" }));
  expect(FakeSocket.all.length).toBe(before + 1);
});

test("an agent that exited says so with its code", () => {
  const { socket } = setup();
  act(() => socket().open());
  act(() => socket().onmessage?.({ data: JSON.stringify({ type: "exit", code: 2 }) }));
  expect(screen.getByText("Exited with code 2")).toBeTruthy();
  expect(screen.getByText(/The agent exited with code 2/)).toBeTruthy();
});

test("leaving the screen closes the terminal", () => {
  const { term, view } = setup();
  expect(term().disposed).toBe(false);
  view.unmount();
  expect(term().disposed).toBe(true);
});
