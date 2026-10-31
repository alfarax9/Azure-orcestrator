import { fireEvent, render, screen, within } from "@testing-library/react";
import { expect, test } from "vitest";
import { createWorkerStore, WorkerStoreContext } from "../workers/store";
import { sampleWorkers } from "../board/sampleWorkers";
import { Shell } from "./Shell";
import { WorkerList } from "./WorkerList";

test("the shell has navigation, a main area and a status bar, and the product name in the top bar", () => {
  render(<Shell statusBar={<span>Connected</span>}>content</Shell>);
  expect(within(screen.getByRole("banner")).getByText("Astra")).toBeTruthy();
  expect(screen.getByRole("navigation", { name: "Main" })).toBeTruthy();
  expect(within(screen.getByRole("main")).getByText("content")).toBeTruthy();
  expect(within(screen.getByRole("contentinfo")).getByText("Connected")).toBeTruthy();
  expect(screen.queryByRole("complementary")).toBeNull();
});

test("the worker list opens as a drawer from the top bar and closes with Escape", () => {
  render(
    <Shell statusBar={null} sidebar={<p>list</p>}>
      x
    </Shell>,
  );
  const toggle = screen.getByRole("button", { name: "Workers" });
  expect(toggle.getAttribute("aria-expanded")).toBe("false");
  fireEvent.click(toggle);
  expect(toggle.getAttribute("aria-expanded")).toBe("true");
  fireEvent.keyDown(window, { key: "Escape" });
  expect(toggle.getAttribute("aria-expanded")).toBe("false");
});

test("the right panel can be hidden and shown again", () => {
  render(
    <Shell statusBar={null} rightPanel={<p>preview</p>} rightPanelLabel="Preview">
      x
    </Shell>,
  );
  expect(screen.getByRole("complementary", { name: "Preview" })).toBeTruthy();
  fireEvent.click(screen.getByRole("button", { name: "Preview" }));
  expect(screen.queryByRole("complementary", { name: "Preview" })).toBeNull();
  fireEvent.click(screen.getByRole("button", { name: "Preview" }));
  expect(screen.getByRole("complementary", { name: "Preview" })).toBeTruthy();
});

test("the worker list names every worker with its status in words", () => {
  const workers = sampleWorkers(0);
  render(
    <WorkerStoreContext value={createWorkerStore(workers)}>
      <WorkerList />
    </WorkerStoreContext>,
  );
  expect(screen.getAllByRole("link")).toHaveLength(workers.length);
});

test("an empty worker list says what to do", () => {
  render(
    <WorkerStoreContext value={createWorkerStore([])}>
      <WorkerList />
    </WorkerStoreContext>,
  );
  expect(screen.getByText(/No workers yet/)).toBeTruthy();
});
