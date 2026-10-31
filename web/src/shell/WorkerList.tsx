import { useRoute, workerHref } from "../router";
import { STATUS } from "../workers/status";
import { useOrder, useStatuses, useWorkerStore } from "../workers/store";
import "./WorkerList.css";

/** The sidebar's list of workers: title and status in words, the open worker marked (DESIGN_SYSTEM.md §9.1). */
export function WorkerList() {
  const order = useOrder();
  const statuses = useStatuses();
  const store = useWorkerStore();
  const route = useRoute();
  const openId = route.name === "worker" ? route.id : null;

  return (
    <section className="worker-list" aria-labelledby="worker-list-title">
      <h2 id="worker-list-title" className="worker-list-title">
        Workers
      </h2>
      {order.length === 0 ? (
        <p className="worker-list-empty">No workers yet. Start a task from the Board.</p>
      ) : (
        <ul className="worker-list-items">
          {order.map((id) => {
            const worker = store.getWorker(id);
            const status = statuses[id];
            if (!worker || !status) return null;
            return (
              <li key={id}>
                <a className="worker-list-item" href={workerHref(id)} aria-current={id === openId ? "page" : undefined}>
                  <span className="worker-list-name">{worker.title}</span>
                  <span className="worker-list-status">{STATUS[status].label}</span>
                </a>
              </li>
            );
          })}
        </ul>
      )}
    </section>
  );
}
