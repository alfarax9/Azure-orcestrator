import { useState } from "react";
import { Board } from "./board/Board";
import { sampleWorkers } from "./board/sampleWorkers";
import { Shell } from "./shell/Shell";
import { StatusBar } from "./shell/StatusBar";
import { createWorkerStore, WorkerStoreContext } from "./workers/store";

/** `?sim` turns on the fact simulator; `?sim=250` sets its rate. */
function simulationRate(search: string): number | null {
  const value = new URLSearchParams(search).get("sim");
  if (value === null) return null;
  const rate = Number(value);
  return Number.isFinite(rate) && rate > 0 ? rate : 100;
}

export function App() {
  const [store] = useState(() => createWorkerStore(sampleWorkers(Date.now())));
  const [simulating, setSimulating] = useState(false);
  return (
    <WorkerStoreContext value={store}>
      <Shell statusBar={<StatusBar simRate={simulationRate(location.search)} onSimulatingChange={setSimulating} />}>
        <Board quiet={simulating} />
      </Shell>
    </WorkerStoreContext>
  );
}
