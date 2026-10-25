import { useEffect, useState } from "react";
import { Button } from "react-aria-components";
import { useWorkerStore } from "../workers/store";
import { startFrameMeter } from "./frameMeter";
import { startSimulation } from "./simulator";

type Stats = { perSecond: number; frames: number; dropped: number; worstMs: number; longTasks: number };

/** Shown with ?sim or ?sim=<rate> in the URL (ADR 0001, validation 3). */
export function SimulationControls({ rate, onRunningChange }: { rate: number; onRunningChange: (on: boolean) => void }) {
  const store = useWorkerStore();
  const [running, setRunning] = useState(false);
  const [stats, setStats] = useState<Stats | null>(null);

  useEffect(() => {
    onRunningChange(running);
    if (!running) return;
    const meter = startFrameMeter();
    const sim = startSimulation(store, rate);
    let lastCount = 0;
    let lastTime = performance.now();
    const timer = setInterval(() => {
      const now = performance.now();
      const count = sim.count();
      const perSecond = Math.round(((count - lastCount) * 1000) / (now - lastTime));
      lastCount = count;
      lastTime = now;
      setStats({ perSecond, ...meter.read() });
    }, 1000);
    return () => {
      clearInterval(timer);
      sim.stop();
      meter.stop();
    };
  }, [running, rate, store, onRunningChange]);

  return (
    <>
      <Button className="status-bar-button" onPress={() => setRunning((on) => !on)}>
        {running ? "Stop simulation" : `Simulate ${rate} facts/s`}
      </Button>
      {stats && (
        <span data-testid="sim-stats">
          {stats.perSecond} facts/s · {stats.dropped} of {stats.frames} frames dropped · worst frame {stats.worstMs} ms ·{" "}
          {stats.longTasks} long tasks
        </span>
      )}
    </>
  );
}
