import { protocols } from "../daemon/connection";

type TerminalPhase = "connecting" | "open" | "reconnecting" | "unavailable" | "exited";

export interface TerminalState {
  phase: TerminalPhase;
  exitCode?: number;
}

interface TerminalSocket {
  binaryType: string;
  readyState: number;
  onopen: (() => void) | null;
  onmessage: ((e: { data: unknown }) => void) | null;
  onclose: (() => void) | null;
  send(data: string | Uint8Array): void;
  close(): void;
}

export type TerminalSocketFactory = (url: string, protocols: string[]) => TerminalSocket;

export interface TerminalOptions {
  url: string;
  token: string;
  /** Bytes the agent wrote. */
  onOutput: (bytes: Uint8Array) => void;
  /** A connection opened: the host repaints the screen from its scrollback, so the terminal starts clean. */
  onReset: () => void;
  onState: (s: TerminalState) => void;
  createSocket?: TerminalSocketFactory;
}

export interface TerminalSession {
  /** Types bytes into the agent; dropped while disconnected, never queued for later. */
  send(bytes: Uint8Array): void;
  resize(rows: number, cols: number): void;
  /** Connects again now, for the Retry button and after the agent was started. */
  reconnect(): void;
  close(): void;
}

const OPEN = 1;
const FIRST_RETRY_MS = 1000;
const MAX_RETRY_MS = 15_000;

/**
 * One worker's terminal over the daemon's PTY socket (ADR 0006). A connection that fails before
 * it ever opens means there is nothing to attach to, so it stops and says so; one that drops
 * after being open is retried with a growing delay, since the agent keeps running without us.
 */
export function connectTerminal(opts: TerminalOptions): TerminalSession {
  const createSocket: TerminalSocketFactory = opts.createSocket ?? ((url, p) => new WebSocket(url, p) as unknown as TerminalSocket);
  let socket: TerminalSocket | null = null;
  let timer: ReturnType<typeof setTimeout> | undefined;
  let size: { rows: number; cols: number } | null = null;
  let attempt = 0;
  let closed = false;
  let finished = false;

  function sendSize() {
    if (size && socket?.readyState === OPEN) socket.send(JSON.stringify({ type: "resize", ...size }));
  }

  function open(first: boolean) {
    clearTimeout(timer);
    socket?.close();
    opts.onState({ phase: first ? "connecting" : "reconnecting" });
    const s = createSocket(opts.url, protocols(opts.token));
    s.binaryType = "arraybuffer";
    socket = s;
    let opened = false;
    s.onopen = () => {
      opened = true;
      attempt = 0;
      opts.onReset();
      opts.onState({ phase: "open" });
      sendSize();
    };
    s.onmessage = (e) => {
      if (e.data instanceof ArrayBuffer) {
        opts.onOutput(new Uint8Array(e.data));
      } else if (typeof e.data === "string") {
        const exit = exitCode(e.data);
        if (exit !== null) {
          finished = true;
          opts.onState({ phase: "exited", exitCode: exit });
        }
      }
    };
    s.onclose = () => {
      if (closed || finished || socket !== s) return;
      if (!opened && first) {
        opts.onState({ phase: "unavailable" });
        return;
      }
      const delay = Math.min(MAX_RETRY_MS, FIRST_RETRY_MS * 2 ** attempt);
      attempt += 1;
      opts.onState({ phase: "reconnecting" });
      timer = setTimeout(() => open(false), delay);
    };
  }

  open(true);
  return {
    send(bytes) {
      if (socket?.readyState === OPEN) socket.send(bytes);
    },
    resize(rows, cols) {
      size = { rows, cols };
      sendSize();
    },
    reconnect() {
      if (closed) return;
      finished = false;
      attempt = 0;
      open(true);
    },
    close() {
      closed = true;
      clearTimeout(timer);
      socket?.close();
    },
  };
}

function exitCode(text: string): number | null {
  try {
    const m = JSON.parse(text) as { type?: unknown; code?: unknown };
    return m.type === "exit" && typeof m.code === "number" ? m.code : null;
  } catch {
    return null;
  }
}
