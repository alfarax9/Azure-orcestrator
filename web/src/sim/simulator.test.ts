import { expect, test } from "vitest";
import { sampleWorkers } from "../board/sampleWorkers";
import { deriveStatus, type WorkerStatus } from "../workers/status";
import { advance } from "./simulator";

test("advancing any sample worker cycles through the loop and back to work", () => {
  const random = () => 0.5;
  for (const start of sampleWorkers(0)) {
    const seen = new Set<WorkerStatus>();
    let w = start;
    for (let i = 0; i < 20; i++) {
      w = advance(w, random);
      seen.add(deriveStatus(w));
    }
    expect(seen.has("live")).toBe(true);
    expect(seen.has("working")).toBe(true);
  }
});
