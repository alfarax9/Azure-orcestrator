import { STATUS, type WorkerStatus } from "../workers/status";
import "./WorkerCard.css";

/** The status as words with a mark; colour never carries the meaning alone (DESIGN_SYSTEM.md §8.5). */
export function StatusMark({ status }: { status: WorkerStatus }) {
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
