import { expect, test } from "vitest";
import cases from "../../../contract/status-cases.json";
import { deriveStages } from "./stages";
import { deriveStatus } from "./status";
import { testWorker } from "./testWorker";
import type { Worker } from "./types";

// The daemon (internal/domain) derives status and stages from the same cases; neither side
// may drift from the other, so both read contract/status-cases.json.
test.each(cases)("$name", (c) => {
  const worker = testWorker(c.facts as unknown as Partial<Worker>);
  expect(deriveStatus(worker)).toBe(c.status);
  expect(deriveStages(worker).map((s) => s.state)).toEqual(c.stages);
});
