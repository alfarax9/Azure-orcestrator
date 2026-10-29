import { ArrowLeft } from "lucide-react";
import { lazy, Suspense } from "react";
import { Tab, TabList, TabPanel, Tabs } from "react-aria-components";
import { StatusMark } from "../board/StatusMark";
import { useDaemon, useFacts } from "../daemon/DaemonContext";
import { boardHref } from "../router";
import { statusDetail } from "../workers/status";
import { useAnnouncement, useStatuses, useWorker } from "../workers/store";
import type { Worker } from "../workers/types";
import { LoopStrip } from "./LoopStrip";
import "./WorkerDetail.css";

// xterm.js is most of the page's weight and only this screen needs it, so it loads on demand.
const TerminalView = lazy(() => import("../terminal/TerminalView").then((m) => ({ default: m.TerminalView })));

/** The main finding of the screen, in a sentence next to the title (DESIGN_RULES.md §4). */
function finding(w: Worker): string | null {
  if (w.agentActivity === "needs_input") return "The agent is waiting for your answer. Reply in the terminal.";
  if (w.agentActivity === "idle" && !w.previewLive && w.gate.state === "none" && w.deployment.state === "none") {
    return "The agent is waiting for your next message. Type it in the terminal.";
  }
  return null;
}

function BackToBoard() {
  return (
    <a className="back-link" href={boardHref}>
      <ArrowLeft size={16} strokeWidth={1.5} aria-hidden="true" />
      Board
    </a>
  );
}

/** Worker detail: "How far along is this task, and what is the next step?" (DESIGN_SYSTEM.md §9.3). */
export function WorkerDetail({ id }: { id: string }) {
  const worker = useWorker(id);
  const status = useStatuses()[id];
  const announcement = useAnnouncement();
  const { loaded } = useFacts();
  const { api, token } = useDaemon();

  if (!worker || !status) {
    return (
      <div className="worker">
        <BackToBoard />
        {loaded ? (
          <>
            <h1>Worker not found</h1>
            <p className="worker-finding">It may have been removed. Go back to the Board to see what is running.</p>
          </>
        ) : (
          <span className="worker-title-placeholder" aria-hidden="true" />
        )}
      </div>
    );
  }

  const detail = statusDetail(worker, status);
  const next = finding(worker);
  return (
    <div className="worker">
      <BackToBoard />
      <p className="visually-hidden" aria-live="polite">
        {announcement}
      </p>
      <header className="worker-header">
        <h1>{worker.title}</h1>
        <div className="worker-meta">
          <StatusMark status={status} />
          {detail && <span>{detail}</span>}
          <span className="worker-place">
            {worker.project} · <span className="worker-branch">{worker.branch}</span>
          </span>
          <span>
            +{worker.added} −{worker.removed}
          </span>
        </div>
        {next && <p className="worker-finding">{next}</p>}
      </header>
      <LoopStrip worker={worker} />
      <Tabs defaultSelectedKey="terminal">
        <TabList aria-label="Worker" className="worker-tabs">
          <Tab id="terminal" className="worker-tab">
            Terminal
          </Tab>
        </TabList>
        <TabPanel id="terminal" className="worker-panel">
          <Suspense fallback={<div className="terminal-loading" aria-hidden="true" />}>
            <TerminalView workerId={worker.id} title={worker.title} agent={worker.agent} token={token} api={api} />
          </Suspense>
        </TabPanel>
      </Tabs>
    </div>
  );
}
