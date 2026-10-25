import { Plus } from "lucide-react";
import { useMemo, useRef } from "react";
import { Button, Tab, TabList, TabPanel, Tabs } from "react-aria-components";
import { COLUMNS, STATUS, type BoardColumn } from "../workers/status";
import { useAnnouncement, useOrder, useStatuses, useWorkerStore } from "../workers/store";
import { ColumnList } from "./ColumnList";
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

let newTasks = 0;

/** Board: "What needs me now?" (DESIGN_SYSTEM.md §9.2). `quiet` mutes announcements while simulating. */
export function Board({ quiet = false }: { quiet?: boolean }) {
  const store = useWorkerStore();
  const columns = useColumns();
  const announcement = useAnnouncement();
  const phone = useMediaQuery("(max-width: 640px)");
  const ref = useRef<HTMLDivElement>(null);
  useFlip(ref);

  const needYou = columns.needs_you.length;

  function newTask() {
    newTasks += 1;
    store.add({
      id: `new-${newTasks}`,
      title: "Untitled task",
      project: "warung-sari / web",
      branch: `task/untitled-${newTasks}`,
      agent: "Claude Code",
      createdAt: Date.now(),
      clientReview: true,
      agentActivity: "working",
      previewLive: false,
      shared: false,
      newFeedback: 0,
      gate: { state: "none" },
      deployment: { state: "none" },
      added: 0,
      removed: 0,
    });
  }

  return (
    <div className="board">
      <header className="board-header">
        <h1>
          Board <span className="board-title-count">· {needYou} {needYou === 1 ? "needs" : "need"} you</span>
        </h1>
        <Button className="button-primary" onPress={newTask}>
          <Plus size={16} strokeWidth={1.5} aria-hidden="true" />
          New task
        </Button>
      </header>

      <p className="visually-hidden" aria-live="polite">
        {quiet ? "" : announcement}
      </p>

      <div ref={ref} className="board-body">
        {phone ? (
          <Tabs defaultSelectedKey="needs_you">
            <TabList aria-label="Board columns" className="board-tabs">
              {COLUMNS.map((c) => (
                <Tab key={c.id} id={c.id} className="board-tab">
                  {c.label} <span className="column-count">{columns[c.id].length}</span>
                </Tab>
              ))}
            </TabList>
            {COLUMNS.map((c) => (
              <TabPanel key={c.id} id={c.id} className="board-panel">
                <ColumnList label={c.label} ids={columns[c.id]} />
              </TabPanel>
            ))}
          </Tabs>
        ) : (
          <div className="board-columns">
            {COLUMNS.map((c) => (
              <section key={c.id} className="board-column" aria-labelledby={`column-${c.id}`}>
                <h2 id={`column-${c.id}`} className="column-title">
                  {c.label} <span className="column-count">{columns[c.id].length}</span>
                </h2>
                <ColumnList label={c.label} ids={columns[c.id]} />
              </section>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
