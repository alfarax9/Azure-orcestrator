import { deriveStages, type StageState } from "../workers/stages";
import type { Worker } from "../workers/types";
import "./LoopStrip.css";

const STATE_WORDS: Record<StageState, string> = {
  done: "done",
  running: "in progress",
  pending: "not started",
  failed: "failed",
  skipped: "skipped",
};

/** The reason shown on a failed stage's label (DESIGN_SYSTEM.md §8.9). */
function failure(w: Worker): string | null {
  if (w.deployment.state === "rolled_back") return w.deployment.reason;
  if (w.gate.state === "failed") return w.gate.cause;
  return null;
}

/** Full-size loop strip: seven stages, each with its state in words as well as in line style (F4-02). */
export function LoopStrip({ worker }: { worker: Worker }) {
  const reason = failure(worker);
  return (
    <ol className="loop" aria-label="Loop">
      {deriveStages(worker).map((s) => (
        <li key={s.stage} className="loop-stage" data-state={s.state} aria-current={s.state === "running" ? "step" : undefined}>
          <span className="loop-line" aria-hidden="true" />
          <span className="loop-label">
            {s.stage}
            {s.state === "failed" && reason ? `: ${reason}` : ""}
            <span className="visually-hidden"> ({STATE_WORDS[s.state]})</span>
          </span>
        </li>
      ))}
    </ol>
  );
}
