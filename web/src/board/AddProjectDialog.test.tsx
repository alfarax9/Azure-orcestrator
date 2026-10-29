import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { expect, test, vi } from "vitest";
import { AddProjectDialog, projectFrom, slugFrom } from "./AddProjectDialog";

function setup(workspaces = [{ slug: "warung-sari", name: "Warung Sari" }], api = { createWorkspace: vi.fn(() => Promise.resolve()), createProject: vi.fn(() => Promise.resolve()) }) {
  const onClose = vi.fn();
  const onAdded = vi.fn();
  render(<AddProjectDialog api={api} workspaces={workspaces} onClose={onClose} onAdded={onAdded} />);
  return { api, onClose, onAdded };
}

test("slugs and project names are made from what people type", () => {
  expect(slugFrom("Warung Sari")).toBe("warung-sari");
  expect(slugFrom("  Toko  Maju Jaya & Co.  ")).toBe("toko-maju-jaya-co");
  expect(slugFrom("2026 Client")).toBe("c-2026-client");
  expect(slugFrom("A very long client name that goes on")).toBe("a-very-long-client-n");
  expect(slugFrom("!!!")).toBe("");
  expect(slugFrom("")).toBe("");
  expect(projectFrom("https://github.com/acme/Shop-Front.git")).toBe("shop-front");
  expect(projectFrom("git@github.com:acme/api.git")).toBe("api");
  expect(projectFrom("https://git.warung-sari.id/repo/")).toBe("repo");
  expect(projectFrom("")).toBe("");
});

test("a project is added to an existing client, named from its repository until the person names it", async () => {
  const { api, onAdded } = setup();
  fireEvent.change(screen.getByLabelText("Repository URL"), { target: { value: "https://github.com/warung/web-sari.git" } });
  expect((screen.getByLabelText("Project name") as HTMLInputElement).value).toBe("web-sari");
  fireEvent.click(screen.getByRole("button", { name: "Add project" }));
  await waitFor(() => expect(onAdded).toHaveBeenCalled());
  expect(api.createWorkspace).not.toHaveBeenCalled();
  expect(api.createProject).toHaveBeenCalledWith("warung-sari", "web-sari", "https://github.com/warung/web-sari.git");
});

test("a project name typed by hand is kept when the repository changes", async () => {
  const { api } = setup();
  fireEvent.change(screen.getByLabelText("Project name"), { target: { value: "storefront" } });
  fireEvent.change(screen.getByLabelText("Repository URL"), { target: { value: "https://git.warung-sari.id/other.git" } });
  fireEvent.click(screen.getByRole("button", { name: "Add project" }));
  await waitFor(() => expect(api.createProject).toHaveBeenCalledWith("warung-sari", "storefront", "https://git.warung-sari.id/other.git"));
});

test("a first client is created with its project, and an empty repository needs no URL", async () => {
  const { api, onAdded } = setup([]);
  expect((screen.getByLabelText("Client") as HTMLSelectElement).value).toBe("");
  fireEvent.change(screen.getByLabelText("Client name"), { target: { value: "Toko Maju" } });
  expect(screen.getByText("Folder and user name: toko-maju")).toBeTruthy();
  fireEvent.change(screen.getByLabelText("Project name"), { target: { value: "web" } });
  fireEvent.click(screen.getByRole("button", { name: "Add project" }));
  await waitFor(() => expect(onAdded).toHaveBeenCalled());
  expect(api.createWorkspace).toHaveBeenCalledWith("toko-maju", "Toko Maju");
  expect(api.createProject).toHaveBeenCalledWith("toko-maju", "web", "");
});

test("choosing a new client among existing ones asks for its name", () => {
  setup();
  expect(screen.queryByLabelText("Client name")).toBeNull();
  fireEvent.change(screen.getByLabelText("Client"), { target: { value: "" } });
  expect(screen.getByLabelText("Client name")).toBeTruthy();
});

test("nothing can be added without the facts it needs", () => {
  setup([]);
  const add = screen.getByRole("button", { name: "Add project" });
  expect(add.hasAttribute("disabled")).toBe(true);
  fireEvent.change(screen.getByLabelText("Client name"), { target: { value: "Toko" } });
  expect(add.hasAttribute("disabled")).toBe(true);
  fireEvent.change(screen.getByLabelText("Project name"), { target: { value: "web" } });
  expect(add.hasAttribute("disabled")).toBe(false);
});

test("a clone that fails says what git said, and the dialog stays open", async () => {
  const api = { createWorkspace: vi.fn(), createProject: vi.fn(() => Promise.reject(new Error("git: repository 'https://x/y.git/' not found"))) };
  const { onAdded } = setup(undefined, api);
  fireEvent.change(screen.getByLabelText("Project name"), { target: { value: "web" } });
  fireEvent.click(screen.getByRole("button", { name: "Add project" }));
  expect((await screen.findByRole("alert")).textContent).toBe("git: repository 'https://x/y.git/' not found");
  expect(onAdded).not.toHaveBeenCalled();
  expect(screen.getByRole("button", { name: "Add project" }).hasAttribute("disabled")).toBe(false);
});

test("Cancel closes without adding anything", () => {
  const { api, onClose } = setup();
  fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
  expect(onClose).toHaveBeenCalled();
  expect(api.createProject).not.toHaveBeenCalled();
});
