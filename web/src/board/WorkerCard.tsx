import { EllipsisVertical, GripVertical } from "lucide-react";
import { Button, Menu, MenuItem, MenuTrigger, Popover, Tooltip, TooltipTrigger } from "react-aria-components";
import { currentStageLabel } from "../workers/stages";
import { statusDetail, type WorkerStatus } from "../workers/status";
import type { Worker } from "../workers/types";
import { LoopStripMini } from "./LoopStripMini";
import { StatusMark } from "./StatusMark";
import type { MoveTarget } from "./move";
import { useNow } from "./useNow";
import "./WorkerCard.css";

function age(createdAt: number, now: number): string {
  const minutes = Math.max(0, Math.round((now - createdAt) / 60_000));
  if (minutes < 60) return `${minutes}m`;
  if (minutes < 24 * 60) return `${Math.floor(minutes / 60)}h`;
  return `${Math.floor(minutes / (24 * 60))}d`;
}

/** Single-click and keyboard alternative to dragging (DESIGN_RULES.md §6). */
function MoveMenu({ title, index, count, onMove }: { title: string; index: number; count: number; onMove: (to: MoveTarget) => void }) {
  const disabledKeys = [...(index === 0 ? ["top", "up"] : []), ...(index === count - 1 ? ["down", "bottom"] : [])];
  return (
    <MenuTrigger>
      <TooltipTrigger delay={300}>
        <Button className="card-icon-button" aria-label={`Move ${title}`}>
          <EllipsisVertical size={16} strokeWidth={1.5} aria-hidden="true" />
        </Button>
        <Tooltip className="tooltip" offset={6}>
          Move
        </Tooltip>
      </TooltipTrigger>
      <Popover className="menu-popover" placement="bottom end">
        <Menu className="menu" disabledKeys={disabledKeys} onAction={(key) => onMove(key as MoveTarget)}>
          <MenuItem id="top" className="menu-item">
            Move to top
          </MenuItem>
          <MenuItem id="up" className="menu-item">
            Move up
          </MenuItem>
          <MenuItem id="down" className="menu-item">
            Move down
          </MenuItem>
          <MenuItem id="bottom" className="menu-item">
            Move to bottom
          </MenuItem>
        </Menu>
      </Popover>
    </MenuTrigger>
  );
}

export function WorkerCard({
  worker: w,
  status,
  index,
  count,
  onMove,
}: {
  worker: Worker;
  status: WorkerStatus;
  index: number;
  count: number;
  onMove: (to: MoveTarget) => void;
}) {
  const now = useNow();
  const detail = statusDetail(w, status);
  const place = `${w.project} · ${w.branch}`;
  return (
    <div className="card" data-card-id={w.id}>
      <div className="card-head">
        <span className="card-title">{w.title}</span>
        <TooltipTrigger delay={300}>
          <Button slot="drag" className="card-icon-button card-drag" aria-label={`Reorder ${w.title}`}>
            <GripVertical size={16} strokeWidth={1.5} aria-hidden="true" />
          </Button>
          <Tooltip className="tooltip" offset={6}>
            Drag to reorder
          </Tooltip>
        </TooltipTrigger>
        <MoveMenu title={w.title} index={index} count={count} onMove={onMove} />
      </div>
      <div className="card-status">
        <StatusMark status={status} />
        {detail && <span className="card-detail">{detail}</span>}
      </div>
      <div className="card-line" title={place}>
        {w.project} · <span className="card-branch">{w.branch}</span>
      </div>
      <LoopStripMini worker={w} />
      <span className="visually-hidden">{currentStageLabel(w)}</span>
      <div className="card-line">
        {w.agent} · {age(w.createdAt, now)} · +{w.added} −{w.removed}
      </div>
    </div>
  );
}
