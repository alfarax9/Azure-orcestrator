import { memo, useCallback, useMemo } from "react";
import { DropIndicator, GridList, GridListItem, useDragAndDrop } from "react-aria-components";
import { deriveStatus, STATUS } from "../workers/status";
import { useWorker, useWorkerStore } from "../workers/store";
import { moveTarget, type MoveTarget } from "./move";
import { WorkerCard } from "./WorkerCard";

type CardItemProps = { id: string; index: number; count: number; onMove: (id: string, to: MoveTarget) => void };

const CardItem = memo(function CardItem({ id, index, count, onMove }: CardItemProps) {
  const worker = useWorker(id);
  if (!worker) return null;
  const status = deriveStatus(worker);
  return (
    <GridListItem id={id} textValue={`${worker.title}, ${STATUS[status].label}`} className="card-row">
      <WorkerCard worker={worker} status={status} index={index} count={count} onMove={(to) => onMove(id, to)} />
    </GridListItem>
  );
});

/**
 * One Board column. Cards change column only when their facts change (F4-01);
 * drag and drop (mouse or keyboard) and the card menu reorder cards inside a column.
 */
export function ColumnList({ label, ids, onOpen }: { label: string; ids: string[]; onOpen?: (id: string) => void }) {
  const store = useWorkerStore();
  const items = useMemo(() => ids.map((id, index) => ({ id, index })), [ids]);
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

  const move = useCallback(
    (id: string, to: MoveTarget) => {
      const next = moveTarget(ids, id, to);
      if (!next) return;
      store.reorder([id], next.target, next.position);
      const title = store.getWorker(id)?.title ?? id;
      store.announce(`${title} moved to position ${next.index + 1} of ${ids.length} in ${label}`);
    },
    [ids, label, store],
  );

  return (
    <GridList
      aria-label={label}
      items={items}
      dragAndDropHooks={dragAndDropHooks}
      className="card-list"
      onAction={onOpen ? (key) => onOpen(String(key)) : undefined}
      renderEmptyState={() => <p className="column-empty">No workers here</p>}
    >
      {(item) => <CardItem id={item.id} index={item.index} count={ids.length} onMove={move} />}
    </GridList>
  );
}
