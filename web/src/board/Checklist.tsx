import { useEffect, useState } from "react";
import { Button } from "react-aria-components";
import type { Api, Checklist as ChecklistData } from "../daemon/api";
import "./Checklist.css";

// Sign-in to an agent happens in a terminal and no event announces it, so the list is read again now and then.
const REFRESH_MS = 15_000;

function itemState(done: boolean, soon: boolean): string {
  if (soon) return "Soon";
  return done ? "Done" : "To do";
}

/**
 * First-run checklist (F1-07): what a new user still has to do, read from facts, with a real
 * percentage. It goes away once every step that can be done now is done, and it stays quiet
 * when the daemon cannot say.
 */
export function Checklist({
  api,
  live,
  facts,
  onAddProject,
}: {
  api: Pick<Api, "checklist">;
  /** Whether the daemon is connected; the list is read again when it is, and when the facts change. */
  live: boolean;
  /** Changes whenever the projects do, so adding one is seen at once. */
  facts: number;
  onAddProject: () => void;
}) {
  const [data, setData] = useState<ChecklistData | null>(null);

  useEffect(() => {
    if (!live) return;
    let cancelled = false;
    const read = () =>
      api
        .checklist()
        .then((c) => !cancelled && setData(c))
        .catch(() => undefined); // keep what was shown; the status bar already says when the daemon is away
    void read();
    const timer = setInterval(read, REFRESH_MS);
    return () => {
      cancelled = true;
      clearInterval(timer);
    };
  }, [api, live, facts]);

  if (!data || data.complete) return null;
  return (
    <section className="checklist" aria-labelledby="checklist-title">
      <div className="checklist-head">
        <h2 id="checklist-title">Get started</h2>
        <span className="checklist-percent">{data.percent}% done</span>
      </div>
      <ol className="checklist-items">
        {data.items.map((item) => (
          <li key={item.id} className="checklist-item" data-done={item.done} data-soon={item.soon}>
            <span className="checklist-state">{itemState(item.done, item.soon)}</span>
            <div className="checklist-text">
              <span className="checklist-title">{item.title}</span>
              <span className="checklist-detail">{item.detail}</span>
            </div>
            {item.id === "project" && !item.done && (
              <Button className="button-secondary" onPress={onAddProject}>
                Add a project
              </Button>
            )}
          </li>
        ))}
      </ol>
    </section>
  );
}
