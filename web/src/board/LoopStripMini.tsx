import { deriveStages } from "../workers/stages";
import type { Worker } from "../workers/types";

/** 4px strip without labels for cards (DESIGN_SYSTEM.md §8.9). The card carries the stage as text. */
export function LoopStripMini({ worker }: { worker: Worker }) {
  return (
    <ol className="loop-mini" aria-hidden="true">
      {deriveStages(worker).map((s) => (
        <li key={s.stage} className="loop-mini-seg" data-state={s.state} />
      ))}
    </ol>
  );
}
