import { FitAddon } from "@xterm/addon-fit";
import { Terminal } from "@xterm/xterm";
import "@xterm/xterm/css/xterm.css";
import { useCallback, useEffect, useRef, useState } from "react";
import { Button } from "react-aria-components";
import type { Api } from "../daemon/api";
import { connectTerminal, type TerminalSession, type TerminalSocketFactory, type TerminalState } from "./session";
import { terminalFont, terminalTheme } from "./theme";
import "./TerminalView.css";

const encoder = new TextEncoder();

export function terminalUrl(workerId: string): string {
  const scheme = window.location.protocol === "https:" ? "wss" : "ws";
  return `${scheme}://${window.location.host}/api/v1/sessions/${encodeURIComponent(workerId)}/pty`;
}

function phaseText(s: TerminalState): string {
  switch (s.phase) {
    case "connecting":
      return "Connecting";
    case "open":
      return "Connected";
    case "reconnecting":
      return "Reconnecting";
    case "unavailable":
      return "Not running";
    case "exited":
      return `Exited with code ${s.exitCode ?? 0}`;
  }
}

/** Ctrl+J or Cmd+J: focus the terminal (DESIGN_RULES.md §7.1). */
function isFocusShortcut(e: KeyboardEvent): boolean {
  return (e.ctrlKey || e.metaKey) && !e.altKey && !e.shiftKey && e.key.toLowerCase() === "j";
}

/**
 * A worker's agent terminal (DESIGN_SYSTEM.md §8.10): the real terminal of the agent, in a
 * header with its state and Interrupt. The daemon repaints it from the host's scrollback each
 * time it connects, so leaving and coming back loses nothing.
 */
export function TerminalView({
  workerId,
  title,
  agent,
  token,
  api,
  createSocket,
}: {
  workerId: string;
  title: string;
  agent: string;
  token: string;
  api: Pick<Api, "startWorker" | "interruptWorker">;
  createSocket?: TerminalSocketFactory;
}) {
  const container = useRef<HTMLDivElement>(null);
  const header = useRef<HTMLDivElement>(null);
  const terminal = useRef<Terminal | null>(null);
  const session = useRef<TerminalSession | null>(null);
  const [state, setState] = useState<TerminalState>({ phase: "connecting" });
  const [busy, setBusy] = useState(false);
  const [problem, setProblem] = useState<string | null>(null);

  useEffect(() => {
    const el = container.current;
    if (!el) return;
    const term = new Terminal({
      fontFamily: terminalFont(),
      fontSize: 13,
      lineHeight: 18 / 13,
      cursorBlink: true,
      scrollback: 5000,
      theme: terminalTheme(),
    });
    const fit = new FitAddon();
    term.loadAddon(fit);
    term.open(el);
    terminal.current = term;

    const s = connectTerminal({
      url: terminalUrl(workerId),
      token,
      createSocket,
      onOutput: (bytes) => term.write(bytes),
      onReset: () => term.reset(),
      onState: setState,
    });
    session.current = s;

    const sendSize = () => {
      fit.fit();
      s.resize(term.rows, term.cols);
    };
    sendSize();
    const observer = new ResizeObserver(sendSize);
    observer.observe(el);

    const data = term.onData((d) => s.send(encoder.encode(d)));
    // Mouse reports and similar arrive as a string of byte values, not as text.
    const binary = term.onBinary((d) => s.send(Uint8Array.from(d, (c) => c.charCodeAt(0))));
    // Only Ctrl+K belongs to the app while the terminal has focus (DESIGN_RULES.md §7.1), and Shift+Esc is the way out.
    term.attachCustomKeyEventHandler((e) => {
      if (e.type === "keydown" && e.shiftKey && e.key === "Escape") {
        header.current?.focus(); // the next Tab reaches Interrupt
        return false;
      }
      return true;
    });
    const onKey = (e: KeyboardEvent) => {
      if (!isFocusShortcut(e)) return;
      e.preventDefault();
      term.focus();
    };
    document.addEventListener("keydown", onKey);

    return () => {
      document.removeEventListener("keydown", onKey);
      observer.disconnect();
      data.dispose();
      binary.dispose();
      s.close();
      term.dispose();
      terminal.current = null;
      session.current = null;
    };
  }, [workerId, token, createSocket]);

  const start = useCallback(async () => {
    setBusy(true);
    setProblem(null);
    try {
      await api.startWorker(workerId);
      session.current?.reconnect();
    } catch (e) {
      setProblem(e instanceof Error ? e.message : "Couldn't start the agent.");
    } finally {
      setBusy(false);
    }
  }, [api, workerId]);

  async function interrupt() {
    setProblem(null);
    await api.interruptWorker(workerId).catch((e: unknown) => setProblem(e instanceof Error ? e.message : "Couldn't interrupt the agent."));
  }

  const running = state.phase === "open" || state.phase === "reconnecting" || state.phase === "connecting";
  return (
    <section className="terminal" aria-label={`Terminal for ${title}`}>
      <div className="terminal-header" ref={header} tabIndex={-1}>
        <span className="terminal-agent">{agent}</span>
        <span className="terminal-state" role="status">
          {phaseText(state)}
        </span>
        <Button className="button-secondary terminal-action" onPress={interrupt} isDisabled={state.phase !== "open"}>
          Interrupt
        </Button>
      </div>
      <div className="terminal-body">
        <div ref={container} className="terminal-screen" hidden={!running} />
        {!running && (
          <div className="terminal-empty">
            <p>
              {state.phase === "exited"
                ? `The agent exited with code ${state.exitCode ?? 0}. Start it again to continue in the same worktree.`
                : "The agent isn't running, so there is no terminal to show."}
            </p>
            {problem && (
              <p className="field-error" role="alert">
                {problem}
              </p>
            )}
            <div className="terminal-empty-actions">
              <Button className="button-primary" onPress={start} isDisabled={busy}>
                {busy ? "Starting…" : "Start agent"}
              </Button>
              <Button className="button-secondary" onPress={() => session.current?.reconnect()}>
                Retry
              </Button>
            </div>
          </div>
        )}
      </div>
      {problem && running && (
        <p className="field-error terminal-problem" role="alert">
          {problem}
        </p>
      )}
      <p className="terminal-hint">Ctrl+J focuses the terminal · Shift+Esc leaves it</p>
    </section>
  );
}
