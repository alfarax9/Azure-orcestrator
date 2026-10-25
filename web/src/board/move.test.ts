import { expect, test } from "vitest";
import { moveTarget } from "./move";

const ids = ["a", "b", "c", "d"];

test("moves a card next to its new neighbour", () => {
  expect(moveTarget(ids, "c", "up")).toEqual({ target: "b", position: "before", index: 1 });
  expect(moveTarget(ids, "b", "down")).toEqual({ target: "c", position: "after", index: 2 });
  expect(moveTarget(ids, "c", "top")).toEqual({ target: "a", position: "before", index: 0 });
  expect(moveTarget(ids, "b", "bottom")).toEqual({ target: "d", position: "after", index: 3 });
});

test("does nothing at the edges of the column", () => {
  expect(moveTarget(ids, "a", "up")).toBeNull();
  expect(moveTarget(ids, "a", "top")).toBeNull();
  expect(moveTarget(ids, "d", "down")).toBeNull();
  expect(moveTarget(ids, "d", "bottom")).toBeNull();
  expect(moveTarget(ids, "x", "up")).toBeNull();
});
