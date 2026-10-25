import { GripVertical } from "lucide-react";
import { Button } from "react-aria-components";
import { currentStageLabel } from "../workers/stages";
import { STATUS, statusDetail, type WorkerStatus } from "../workers/status";
import type { Worker } from "../workers/types";
import { LoopStripMini } from "./LoopStripMini";
import { useNow } from "./useNow";
import "./WorkerCard.css";

function age(createdAt: number, now: number): string {
  const minutes = Math.max(0, Math.round((now - createdAt) / 60_000));
  if (minutes < 60) return `${minutes}m`;
  if (minutes < 24 * 60) return `${Math.floor(minutes / 60)}h`;
  return `${Math.floor(minutes / (24 * 60))}d`;
}

function StatusMark({ status }: { status: WorkerStatus }) {
  const { label, mark } = STATUS[status];
  if (mark === "pulse-accent" || mark === "dot-positive") {
    return (
      <span className="status-dot">
        <span className={`dot ${mark === "pulse-accent" ? "dot-accent dot-pulse" : "dot-positive"}`} aria-hidden="true" />
        {label}
      </span>
    );
  }
  return <span className={`pill ${mark}`}>{label}</span>;
}

export function WorkerCard({ worker: w, status }: { worker: Worker; status: WorkerStatus }) {
  const now = useNow();
  const detail = statusDetail(w, status);
  const place = `${w.project} · ${w.branch}`;
  return (
    <div className="card" data-card-id={w.id}>
      <div className="card-head">
        <span className="card-title">{w.title}</span>
        <Button slot="drag" className="card-drag" aria-label={`Reorder ${w.title}`}>
          <GripVertical size={16} strokeWidth={1.5} aria-hidden="true" />
        </Button>
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
