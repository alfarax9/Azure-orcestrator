import { useState, type FormEvent } from "react";
import { Button, Checkbox, Dialog, Heading, Input, Label, Modal, ModalOverlay, Radio, RadioGroup, TextField } from "react-aria-components";
import type { Api } from "../daemon/api";
import type { ProjectOption } from "../daemon/facts";
import type { ApiWorker } from "../daemon/wire";

type Agent = "claude" | "codex";

/**
 * Starting a task makes a worker: its own worktree and branch, and an agent in it (F3-01).
 * The agent then asks to trust the folder, so the worker opens straight on its terminal.
 */
export function NewTaskDialog({
  api,
  projects,
  onClose,
  onStarted,
  onAddProject,
}: {
  api: Pick<Api, "createWorker" | "startWorker">;
  projects: ProjectOption[];
  onClose: () => void;
  onStarted: (worker: ApiWorker) => void;
  onAddProject: () => void;
}) {
  const [projectId, setProjectId] = useState(projects[0]?.id ?? "");
  const [title, setTitle] = useState("");
  const [agent, setAgent] = useState<Agent>("claude");
  const [clientReview, setClientReview] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function submit(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError(null);
    let created: ApiWorker;
    try {
      created = await api.createWorker(projectId, { title: title.trim(), agent, clientReview });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Couldn't create the task.");
      setBusy(false);
      return;
    }
    // The worker exists now. If its agent does not start, the worker's screen offers to start it.
    await api.startWorker(created.id).catch(() => undefined);
    onStarted(created);
  }

  return (
    <ModalOverlay isOpen isDismissable onOpenChange={(open) => !open && onClose()} className="modal-scrim">
      <Modal>
        <Dialog className="dialog" aria-label="Start a task">
          <Heading slot="title">Start a task</Heading>
          {projects.length === 0 ? (
            <div className="dialog-form">
              <p className="field-help">No projects yet. Add your first project, then start a task in it.</p>
              <div className="dialog-actions">
                <Button className="button-secondary" onPress={onClose}>
                  Close
                </Button>
                <Button className="button-primary" onPress={onAddProject}>
                  Add a project
                </Button>
              </div>
            </div>
          ) : (
            <form className="dialog-form" onSubmit={submit}>
              <div className="field">
                <label htmlFor="new-task-project">Project</label>
                <select id="new-task-project" value={projectId} onChange={(e) => setProjectId(e.target.value)}>
                  {projects.map((p) => (
                    <option key={p.id} value={p.id}>
                      {p.label}
                    </option>
                  ))}
                </select>
              </div>
              <TextField className="field" isRequired autoFocus value={title} onChange={setTitle}>
                <Label>What should the agent do?</Label>
                <Input maxLength={200} />
              </TextField>
              <RadioGroup className="field radio-group" value={agent} onChange={(v) => setAgent(v as Agent)} orientation="horizontal">
                <Label>Agent</Label>
                <div className="radio-row">
                  <Radio className="radio" value="claude">
                    Claude Code
                  </Radio>
                  <Radio className="radio" value="codex">
                    Codex
                  </Radio>
                </div>
              </RadioGroup>
              <Checkbox className="checkbox" isSelected={clientReview} onChange={setClientReview}>
                <span className="checkbox-box" aria-hidden="true" />
                Share the preview with the client for review
              </Checkbox>
              {error && (
                <p className="field-error" role="alert">
                  {error}
                </p>
              )}
              <div className="dialog-actions">
                <Button className="button-secondary" onPress={onClose}>
                  Cancel
                </Button>
                <Button type="submit" className="button-primary" isDisabled={busy || title.trim() === ""}>
                  {busy ? "Starting…" : "Start task"}
                </Button>
              </div>
            </form>
          )}
        </Dialog>
      </Modal>
    </ModalOverlay>
  );
}
