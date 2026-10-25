import { memo, useMemo } from "react";
import { DropIndicator, GridList, GridListItem, useDragAndDrop } from "react-aria-components";
import { deriveStatus, STATUS } from "../workers/status";
import { useWorker, useWorkerStore } from "../workers/store";
import { WorkerCard } from "./WorkerCard";

const CardItem = memo(function CardItem({ id }: { id: string }) {
  const worker = useWorker(id);
  if (!worker) return null;
  const status = deriveStatus(worker);
  return (
    <GridListItem id={id} textValue={`${worker.title}, ${STATUS[status].label}`} className="card-row">
      <WorkerCard worker={worker} status={status} />
    </GridListItem>
  );
});

/**
 * One Board column. Cards change column only when their facts change (F4-01);
 * drag and drop, by mouse or keyboard, reorders cards inside a column.
 */
export function ColumnList({ label, ids }: { label: string; ids: string[] }) {
  const store = useWorkerStore();
  const items = useMemo(() => ids.map((id) => ({ id })), [ids]);
  const { dragAndDropHooks } = useDragAndDrop({
    getItems: (keys) => [...keys].map((key) => ({ "text/plain": store.getWorker(String(key))?.title ?? String(key) })),
    onReorder: (e) =>
      store.reorder(
        [...e.keys].map(String),
        String(e.target.key),
        e.target.dropPosition === "before" ? "before" : "after",
      ),
    renderDropIndicator: (target) => <DropIndicator target={target} className="drop-indicator" />,
  });

  return (
    <GridList
      aria-label={label}
      items={items}
      dragAndDropHooks={dragAndDropHooks}
      className="card-list"
      renderEmptyState={() => <p className="column-empty">No workers here</p>}
    >
      {(item) => <CardItem id={item.id} />}
    </GridList>
  );
}
