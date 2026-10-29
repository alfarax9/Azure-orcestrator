import { useCallback, useEffect, useMemo, useState } from "react";
import { Board } from "./board/Board";
import { AddProjectDialog } from "./board/AddProjectDialog";
import { Checklist } from "./board/Checklist";
import { NewTaskDialog } from "./board/NewTaskDialog";
import { sampleWorkers } from "./board/sampleWorkers";
import { createApi } from "./daemon/api";
import { connectEvents, type ConnectionState } from "./daemon/connection";
import { DaemonContext, useDaemon, useFacts, type Daemon } from "./daemon/DaemonContext";
import { createFacts } from "./daemon/facts";
import { clearToken, loadToken, saveToken } from "./daemon/token";
import type { ApiWorker } from "./daemon/wire";
import { PairScreen } from "./pair/PairScreen";
import { useRoute, workerHref } from "./router";
import { Shell } from "./shell/Shell";
import { ConnectionStatus } from "./shell/ConnectionStatus";
import { StatusBar } from "./shell/StatusBar";
import { WorkerDetail } from "./worker/WorkerDetail";
import { createWorkerStore, useWorkerStore, WorkerStoreContext, type WorkerStore } from "./workers/store";

/** `?sample` shows the Board with made-up workers and no daemon (the design inspector uses it); `?sim` adds the fact simulator. */
function sampleMode(search: string): boolean {
  const params = new URLSearchParams(search);
  return params.has("sample") || params.has("sim");
}

/** `?sim=250` sets the simulator's rate. */
function simulationRate(search: string): number | null {
  const value = new URLSearchParams(search).get("sim");
  if (value === null) return null;
  const rate = Number(value);
  return Number.isFinite(rate) && rate > 0 ? rate : 100;
}

export function App() {
  return sampleMode(location.search) ? <SampleApp /> : <LiveApp />;
}

let sampleTasks = 0;

function addSampleTask(store: WorkerStore) {
  sampleTasks += 1;
  store.add({
    id: `new-${sampleTasks}`,
    title: "Untitled task",
    project: "warung-sari / web",
    branch: `task/untitled-${sampleTasks}`,
    agent: "Claude Code",
    createdAt: Date.now(),
    clientReview: true,
    agentActivity: "working",
    previewLive: false,
    shared: false,
    newFeedback: 0,
    gate: { state: "none" },
    deployment: { state: "none" },
    added: 0,
    removed: 0,
  });
}

function SampleBoard({ quiet }: { quiet: boolean }) {
  const store = useWorkerStore();
  return <Board quiet={quiet} newTask={{ onPress: () => addSampleTask(store) }} />;
}

function SampleApp() {
  const [store] = useState(() => createWorkerStore(sampleWorkers(Date.now())));
  const [simulating, setSimulating] = useState(false);
  return (
    <WorkerStoreContext value={store}>
      <Shell statusBar={<StatusBar simRate={simulationRate(location.search)} onSimulatingChange={setSimulating} />}>
        <SampleBoard quiet={simulating} />
      </Shell>
    </WorkerStoreContext>
  );
}

function eventsUrl(): string {
  return `${location.protocol === "https:" ? "wss" : "ws"}://${location.host}/api/v1/events`;
}

function LiveApp() {
  const [token, setToken] = useState(loadToken);
  const [notice, setNotice] = useState<string | undefined>();
  const pairingApi = useMemo(() => createApi(() => null, () => undefined), []);

  const signOut = useCallback((why?: string) => {
    clearToken();
    setNotice(why);
    setToken(null);
  }, []);

  if (!token) {
    return (
      <PairScreen
        api={pairingApi}
        notice={notice}
        onPaired={(t) => {
          saveToken(t);
          setNotice(undefined);
          setToken(t);
        }}
      />
    );
  }
  return <Connected token={token} onSignOut={signOut} />;
}

/** The Board and worker screens, kept true by the daemon's event stream. */
function Connected({ token, onSignOut }: { token: string; onSignOut: (why?: string) => void }) {
  const [store] = useState(() => createWorkerStore([]));
  const [facts] = useState(() => createFacts(store));
  const [connection, setConnection] = useState<ConnectionState>({ phase: "connecting", lastUpdate: null, retryAt: null });
  const api = useMemo(() => createApi(() => token, () => onSignOut("This device was disconnected from the daemon. Pair it again to continue.")), [token, onSignOut]);

  useEffect(
    () => connectEvents({ url: eventsUrl(), token, probe: api.probe, onMessage: facts.apply, onState: setConnection }),
    [token, api, facts],
  );
  useEffect(() => {
    if (connection.phase === "unauthorized") onSignOut("This device was disconnected from the daemon. Pair it again to continue.");
  }, [connection.phase, onSignOut]);

  const daemon: Daemon = useMemo(
    () => ({ api, token, connection, facts, signOut: () => onSignOut() }),
    [api, token, connection, facts, onSignOut],
  );
  return (
    <DaemonContext value={daemon}>
      <WorkerStoreContext value={store}>
        <Shell statusBar={<ConnectionStatus state={connection} />}>
          <Screens />
        </Shell>
      </WorkerStoreContext>
    </DaemonContext>
  );
}

function Screens() {
  const route = useRoute();
  const { api, connection, facts } = useDaemon();
  const { loaded, projects, workspaces } = useFacts();
  const [dialog, setDialog] = useState<"task" | "project" | null>(null);

  const started = (w: ApiWorker) => {
    facts.apply({ type: "worker", data: w }); // already known when the screen opens, before the event arrives
    setDialog(null);
    window.location.hash = workerHref(w.id);
  };

  if (route.name === "worker") return <WorkerDetail id={route.id} />;
  return (
    <>
      <Board
        loading={!loaded}
        checklist={<Checklist api={api} live={connection.phase === "live"} facts={projects.length} onAddProject={() => setDialog("project")} />}
        newTask={{
          onPress: () => setDialog("task"),
          disabledReason: connection.phase === "live" ? undefined : "Offline: reconnect to the daemon to start a task.",
        }}
        onOpenWorker={(id) => {
          window.location.hash = workerHref(id);
        }}
      />
      {dialog === "task" && (
        <NewTaskDialog api={api} projects={projects} onClose={() => setDialog(null)} onStarted={started} onAddProject={() => setDialog("project")} />
      )}
      {dialog === "project" && <AddProjectDialog api={api} workspaces={workspaces} onClose={() => setDialog(null)} onAdded={() => setDialog(null)} />}
    </>
  );
}
