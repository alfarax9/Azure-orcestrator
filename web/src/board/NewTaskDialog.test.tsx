import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { expect, test, vi } from "vitest";
import type { ApiWorker } from "../daemon/wire";
import { NewTaskDialog } from "./NewTaskDialog";

const created = { id: "w9", title: "Fix header" } as ApiWorker;
const projects = [
  { id: "p1", label: "acme / site" },
  { id: "p2", label: "acme / shop" },
];

function setup(api = { createWorker: vi.fn(() => Promise.resolve(created)), startWorker: vi.fn(() => Promise.resolve()) }, list = projects) {
  const onClose = vi.fn();
  const onStarted = vi.fn();
  const onAddProject = vi.fn();
  render(<NewTaskDialog api={api} projects={list} onClose={onClose} onStarted={onStarted} onAddProject={onAddProject} />);
  return { api, onClose, onStarted, onAddProject };
}

test("a task becomes a worker with its agent started, and the person lands on it", async () => {
  const { api, onStarted } = setup();
  fireEvent.change(screen.getByLabelText("Project"), { target: { value: "p2" } });
  fireEvent.change(screen.getByLabelText("What should the agent do?"), { target: { value: "  Fix header  " } });
  fireEvent.click(screen.getByRole("radio", { name: "Codex" }));
  fireEvent.click(screen.getByRole("checkbox", { name: /Share the preview/ }));
  fireEvent.click(screen.getByRole("button", { name: "Start task" }));
  await waitFor(() => expect(onStarted).toHaveBeenCalledWith(created));
  expect(api.createWorker).toHaveBeenCalledWith("p2", { title: "Fix header", agent: "codex", clientReview: false });
  expect(api.startWorker).toHaveBeenCalledWith("w9");
});

test("the defaults are the first project, Claude Code and a client review", async () => {
  const { api } = setup();
  fireEvent.change(screen.getByLabelText("What should the agent do?"), { target: { value: "Task" } });
  fireEvent.click(screen.getByRole("button", { name: "Start task" }));
  await waitFor(() => expect(api.createWorker).toHaveBeenCalledWith("p1", { title: "Task", agent: "claude", clientReview: true }));
});

test("a task needs a title", () => {
  setup();
  expect(screen.getByRole("button", { name: "Start task" }).hasAttribute("disabled")).toBe(true);
});

test("a failed start still opens the worker, where the agent can be started again", async () => {
  const { onStarted } = setup({ createWorker: vi.fn(() => Promise.resolve(created)), startWorker: vi.fn(() => Promise.reject(new Error("unit exists"))) });
  fireEvent.change(screen.getByLabelText("What should the agent do?"), { target: { value: "Task" } });
  fireEvent.click(screen.getByRole("button", { name: "Start task" }));
  await waitFor(() => expect(onStarted).toHaveBeenCalledWith(created));
});

test("a task that cannot be created says why and stays open", async () => {
  const { onStarted } = setup({ createWorker: vi.fn(() => Promise.reject(new Error("a task needs a title"))), startWorker: vi.fn() });
  fireEvent.change(screen.getByLabelText("What should the agent do?"), { target: { value: "Task" } });
  fireEvent.click(screen.getByRole("button", { name: "Start task" }));
  expect((await screen.findByRole("alert")).textContent).toBe("a task needs a title");
  expect(onStarted).not.toHaveBeenCalled();
  expect(screen.getByRole("button", { name: "Start task" }).hasAttribute("disabled")).toBe(false);
});

test("without projects the dialog offers to add one", () => {
  const { onClose, onAddProject } = setup(undefined, []);
  expect(screen.getByText(/No projects yet/)).toBeTruthy();
  expect(screen.queryByRole("button", { name: "Start task" })).toBeNull();
  fireEvent.click(screen.getByRole("button", { name: "Add a project" }));
  expect(onAddProject).toHaveBeenCalled();
  fireEvent.click(screen.getByRole("button", { name: "Close" }));
  expect(onClose).toHaveBeenCalled();
});

test("Cancel closes without creating anything", () => {
  const { api, onClose } = setup();
  fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
  expect(onClose).toHaveBeenCalled();
  expect(api.createWorker).not.toHaveBeenCalled();
});
