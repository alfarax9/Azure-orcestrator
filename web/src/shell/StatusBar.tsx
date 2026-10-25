import { useStatuses } from "../workers/store";
import { SimulationControls } from "../sim/SimulationControls";

export function StatusBar({ simRate, onSimulatingChange }: { simRate: number | null; onSimulatingChange: (on: boolean) => void }) {
  const statuses = useStatuses();
  const working = Object.values(statuses).filter((s) => s === "working").length;
  return (
    <>
      <span className="status-bar-brand">Astra</span>
      <span>Sample data, not connected to a daemon</span>
      <span>{working} working</span>
      {simRate !== null && <SimulationControls rate={simRate} onRunningChange={onSimulatingChange} />}
    </>
  );
}
