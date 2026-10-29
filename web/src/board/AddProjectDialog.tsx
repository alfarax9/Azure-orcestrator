import { useState, type FormEvent } from "react";
import { Button, Dialog, Heading, Input, Label, Modal, ModalOverlay, TextField } from "react-aria-components";
import type { Api } from "../daemon/api";
import type { WorkspaceOption } from "../daemon/facts";

const NEW = "";

/** A workspace slug from a client's name: lowercase letters, digits and hyphens, starting with a letter. */
export function slugFrom(name: string): string {
  const slug = name
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, 20)
    .replace(/-+$/, "");
  if (slug === "") return "";
  return /^[a-z]/.test(slug) ? slug : `c-${slug}`.slice(0, 20).replace(/-+$/, "");
}

/** A project name from a repository URL: its last path element without .git. */
export function projectFrom(url: string): string {
  const last = url.trim().replace(/\/+$/, "").split(/[/:]/).pop() ?? "";
  return last.replace(/\.git$/, "").toLowerCase().replace(/[^a-z0-9._-]+/g, "-").replace(/^[^a-z0-9]+/, "");
}

/**
 * Adds a project: a repository cloned on the server, or an empty one, in a client's workspace
 * (F2-01, F2-02). A client with no workspace yet gets one first.
 */
export function AddProjectDialog({
  api,
  workspaces,
  onClose,
  onAdded,
}: {
  api: Pick<Api, "createWorkspace" | "createProject">;
  workspaces: WorkspaceOption[];
  onClose: () => void;
  onAdded: () => void;
}) {
  const [workspace, setWorkspace] = useState(workspaces[0]?.slug ?? NEW);
  const [clientName, setClientName] = useState("");
  const [repo, setRepo] = useState("");
  const [name, setName] = useState("");
  const [nameTouched, setNameTouched] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const creating = workspace === NEW;
  const slug = creating ? slugFrom(clientName) : workspace;
  const projectName = nameTouched ? name : projectFrom(repo);
  const ready = slug !== "" && projectName !== "" && (!creating || clientName.trim() !== "");

  async function submit(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError(null);
    try {
      if (creating) await api.createWorkspace(slug, clientName.trim());
      await api.createProject(slug, projectName, repo.trim());
      onAdded();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Couldn't add the project.");
      setBusy(false);
    }
  }

  return (
    <ModalOverlay isOpen isDismissable onOpenChange={(open) => !open && onClose()} className="modal-scrim">
      <Modal>
        <Dialog className="dialog" aria-label="Add a project">
          <Heading slot="title">Add a project</Heading>
          <form className="dialog-form" onSubmit={submit}>
            <div className="field">
              <label htmlFor="add-project-workspace">Client</label>
              <select id="add-project-workspace" value={workspace} onChange={(e) => setWorkspace(e.target.value)}>
                {workspaces.map((w) => (
                  <option key={w.slug} value={w.slug}>
                    {w.name}
                  </option>
                ))}
                <option value={NEW}>New client…</option>
              </select>
            </div>
            {creating && (
              <TextField className="field" isRequired autoFocus value={clientName} onChange={setClientName}>
                <Label>Client name</Label>
                <Input maxLength={80} />
                <p className="field-help">{slug ? `Folder and user name: ${slug}` : "Each client gets its own folder and its own Linux user on the server."}</p>
              </TextField>
            )}
            <TextField className="field" autoFocus={!creating} value={repo} onChange={setRepo}>
              <Label>Repository URL</Label>
              <Input spellCheck={false} autoCapitalize="off" />
              <p className="field-help">https://, ssh:// or git@host:… Leave it empty to start an empty repository.</p>
            </TextField>
            <TextField
              className="field"
              isRequired
              value={projectName}
              onChange={(v) => {
                setName(v);
                setNameTouched(true);
              }}
            >
              <Label>Project name</Label>
              <Input spellCheck={false} autoCapitalize="off" maxLength={63} />
            </TextField>
            {error && (
              <p className="field-error" role="alert">
                {error}
              </p>
            )}
            <div className="dialog-actions">
              <Button className="button-secondary" onPress={onClose}>
                Cancel
              </Button>
              <Button type="submit" className="button-primary" isDisabled={busy || !ready}>
                {busy ? "Adding…" : "Add project"}
              </Button>
            </div>
          </form>
        </Dialog>
      </Modal>
    </ModalOverlay>
  );
}
