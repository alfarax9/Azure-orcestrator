import { expect, test } from "vitest";
import { parseRoute, workerHref } from "./router";

test("a worker hash opens that worker, and everything else the Board", () => {
  expect(parseRoute("#/workers/abc-123")).toEqual({ name: "worker", id: "abc-123" });
  expect(parseRoute("#/workers/abc-123/")).toEqual({ name: "worker", id: "abc-123" });
  for (const hash of ["", "#", "#/", "#/workers", "#/workers/", "#/workers/a/b", "#/other", "#/workers/%E0%A4%A"]) {
    expect(parseRoute(hash), hash).toEqual({ name: "board" });
  }
});

test("ids with odd characters survive the round trip", () => {
  const id = "a b/c?d";
  expect(parseRoute(workerHref(id))).toEqual({ name: "worker", id });
});
