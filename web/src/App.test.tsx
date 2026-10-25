import { render, screen, within } from "@testing-library/react";
import { expect, test } from "vitest";
import { App } from "./App";
import { sampleWorkers } from "./board/sampleWorkers";

test("the Board shows every sample worker in exactly one column", () => {
  render(<App />);
  expect(screen.getAllByRole("heading", { level: 1 })).toHaveLength(1);
  expect(screen.getByRole("heading", { level: 1 }).textContent).toBe("Board · 6 need you");

  const counts = ["Working", "Needs you", "Review", "Ready to ship", "Live"].map((name) => {
    const column = screen.getByRole("grid", { name });
    return within(column).getAllByRole("row").length;
  });
  expect(counts).toEqual([7, 6, 7, 5, 5]);
  expect(counts.reduce((a, b) => a + b)).toBe(sampleWorkers(0).length);
});
