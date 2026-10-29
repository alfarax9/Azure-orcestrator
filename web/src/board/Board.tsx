import { Plus } from "lucide-react";
import { useMemo, useRef, type ReactNode } from "react";
import { Button, Tab, TabList, TabPanel, Tabs } from "react-aria-components";
import { COLUMNS, STATUS, type BoardColumn } from "../workers/status";
import { useAnnouncement, useOrder, useStatuses } from "../workers/store";
import { ColumnList } from "./ColumnList";
import { ColumnPlaceholder } from "./ColumnPlaceholder";
import { useFlip } from "./useFlip";
import { useMediaQuery } from "./useMediaQuery";
import "./Board.css";

function useColumns(): Record<BoardColumn, string[]> {
  const statuses = useStatuses();
  const order = useOrder();
  return useMemo(() => {
    const columns: Record<BoardColumn, string[]> = { working: [], needs_you: [], review: [], ready_to_ship: [], live: [] };
    for (const id of order) {
      const status = statuses[id];
      if (status) columns[STATUS[status].column].push(id);
    }
    return columns;
  }, [statuses, order]);
}

export interface NewTaskAction {
  onPress: () => void;
  /** Why the action is unavailable right now, shown next to the button; absent when it works. */
  disabledReason?: string;
}

/** Board: "What needs me now?" (DESIGN_SYSTEM.md §9.2). `quiet` mutes announcements while simulating. */
export function Board({
  quiet = false,
  loading = false,
  newTask,
  onOpenWorker,
  checklist,
}: {
  quiet?: boolean;
  /** Before the first facts arrive the columns keep their place and show flat bars. */
  loading?: boolean;
  newTask: NewTaskAction;
  onOpenWorker?: (id: string) => void;
  /** The first-run checklist, shown above the columns while there is something left to do. */
  checklist?: ReactNode;
}) {
  const columns = useColumns();
  const announcement = useAnnouncement();
  const phone = useMediaQuery("(max-width: 640px)");
  const ref = useRef<HTMLDivElement>(null);
  useFlip(ref);

  const needYou = columns.needs_you.length;
  const empty = !loading && COLUMNS.every((c) => columns[c.id].length === 0);
  const count = (id: BoardColumn) => (loading ? "—" : columns[id].length);
  const list = (id: BoardColumn, label: string) =>
    loading ? <ColumnPlaceholder /> : <ColumnList label={label} ids={columns[id]} onOpen={onOpenWorker} />;

  return (
    <div className="board">
      <header className="board-header">
        <h1>
          Board{" "}
          {!loading && (
            <span className="board-title-count">
              · {needYou} {needYou === 1 ? "needs" : "need"} you
            </span>
          )}
        </h1>
        <div className="board-actions">
          {newTask.disabledReason && <span className="board-hint">{newTask.disabledReason}</span>}
          <Button className="button-primary" onPress={newTask.onPress} isDisabled={newTask.disabledReason !== undefined}>
            <Plus size={16} strokeWidth={1.5} aria-hidden="true" />
            New task
          </Button>
        </div>
      </header>

      <p className="visually-hidden" aria-live="polite">
        {quiet ? "" : announcement}
      </p>

      {checklist}

      {empty && <p className="board-empty">No workers yet. Start a task to put an agent to work.</p>}

      <div ref={ref} className="board-body">
        {phone ? (
          <Tabs defaultSelectedKey="needs_you">
            <TabList aria-label="Board columns" className="board-tabs">
              {COLUMNS.map((c) => (
                <Tab key={c.id} id={c.id} className="board-tab">
                  {c.label} <span className="column-count">{count(c.id)}</span>
                </Tab>
              ))}
            </TabList>
            {COLUMNS.map((c) => (
              <TabPanel key={c.id} id={c.id} className="board-panel">
                {list(c.id, c.label)}
              </TabPanel>
            ))}
          </Tabs>
        ) : (
          <div className="board-columns">
            {COLUMNS.map((c) => (
              <section key={c.id} className="board-column" aria-labelledby={`column-${c.id}`}>
                <h2 id={`column-${c.id}`} className="column-title">
                  {c.label} <span className="column-count">{count(c.id)}</span>
                </h2>
                {list(c.id, c.label)}
              </section>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
