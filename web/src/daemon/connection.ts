import { parseMessage, type ServerMessage } from "./wire";

type Phase = "connecting" | "live" | "offline" | "unauthorized";

export interface ConnectionState {
  phase: Phase;
  /** When the last message arrived; kept while offline so the screen can say how old its data is. */
  lastUpdate: number | null;
  /** When the next attempt is made, while offline. */
  retryAt: number | null;
}

export interface SocketLike {
  onmessage: ((e: { data: unknown }) => void) | null;
  onclose: (() => void) | null;
  close(): void;
}

type SocketFactory = (url: string, protocols: string[]) => SocketLike;

export interface ConnectionOptions {
  url: string;
  token: string;
  /** Tells a revoked token from an unreachable daemon: a browser cannot read why a handshake failed. */
  probe: () => Promise<"ok" | "unauthorized" | "unreachable">;
  onMessage: (m: ServerMessage) => void;
  onState: (s: ConnectionState) => void;
  createSocket?: SocketFactory;
  now?: () => number;
}

const FIRST_RETRY_MS = 1000;
const MAX_RETRY_MS = 15_000;

/** Subprotocols a socket offers: the version, and the token a browser cannot put in a header (ADR 0006). */
export function protocols(token: string): string[] {
  return ["astra.v1", `bearer.${token}`];
}

/**
 * Keeps the events socket open: after a drop it checks the token, then retries with a growing
 * delay. Every connection starts with a snapshot, so nothing is replayed by hand. Returns a
 * function that closes it for good.
 */
export function connectEvents(opts: ConnectionOptions): () => void {
  const createSocket: SocketFactory = opts.createSocket ?? ((url, p) => new WebSocket(url, p) as unknown as SocketLike);
  const now = opts.now ?? Date.now;
  let state: ConnectionState = { phase: "connecting", lastUpdate: null, retryAt: null };
  let socket: SocketLike | null = null;
  let timer: ReturnType<typeof setTimeout> | undefined;
  let attempt = 0;
  let closed = false;

  function set(next: Partial<ConnectionState>) {
    state = { ...state, ...next };
    opts.onState(state);
  }

  function open() {
    set({ phase: "connecting", retryAt: null });
    const s = createSocket(opts.url, protocols(opts.token));
    socket = s;
    s.onmessage = (e) => {
      const message = typeof e.data === "string" ? parseMessage(e.data) : null;
      if (!message || closed) return;
      if (message.type === "snapshot") attempt = 0;
      set({ phase: "live", lastUpdate: now() });
      opts.onMessage(message);
    };
    s.onclose = () => {
      if (closed || socket !== s) return;
      void dropped();
    };
  }

  async function dropped() {
    const answer = await opts.probe();
    if (closed) return;
    if (answer === "unauthorized") {
      set({ phase: "unauthorized", retryAt: null });
      return;
    }
    const delay = Math.min(MAX_RETRY_MS, FIRST_RETRY_MS * 2 ** attempt);
    attempt += 1;
    set({ phase: "offline", retryAt: now() + delay });
    timer = setTimeout(open, delay);
  }

  open();
  return () => {
    closed = true;
    clearTimeout(timer);
    socket?.close();
  };
}
