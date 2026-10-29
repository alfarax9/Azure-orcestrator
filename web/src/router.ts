import { useSyncExternalStore } from "react";

export type Route = { name: "board" } | { name: "worker"; id: string };

/** Reads a location hash: `#/workers/<id>` opens a worker, anything else the Board. */
export function parseRoute(hash: string): Route {
  const match = /^#\/workers\/([^/?#]+)\/?$/.exec(hash);
  if (match?.[1]) {
    try {
      return { name: "worker", id: decodeURIComponent(match[1]) };
    } catch {
      return { name: "board" };
    }
  }
  return { name: "board" };
}

export function workerHref(id: string): string {
  return `#/workers/${encodeURIComponent(id)}`;
}

export const boardHref = "#/";

function subscribe(listener: () => void): () => void {
  window.addEventListener("hashchange", listener);
  return () => window.removeEventListener("hashchange", listener);
}

/** The hash is the only router state, so Wails, the PWA and a plain browser all behave alike. */
export function useRoute(): Route {
  const hash = useSyncExternalStore(subscribe, () => window.location.hash);
  return parseRoute(hash);
}
