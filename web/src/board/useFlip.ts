import { useLayoutEffect, useRef, type RefObject } from "react";

type Point = { x: number; y: number };

/**
 * Cards glide to their new place when they change column or order (DESIGN_RULES.md §8),
 * instead of vanishing and reappearing. Runs after every render of the container.
 */
export function useFlip(root: RefObject<HTMLElement | null>) {
  const last = useRef(new Map<string, Point>());

  useLayoutEffect(() => {
    const el = root.current;
    if (!el) return;
    const style = getComputedStyle(el);
    const duration = parseFloat(style.getPropertyValue("--dur-base")) || 0;
    const easing = style.getPropertyValue("--ease-out").trim() || "ease-out";
    const base = el.getBoundingClientRect();
    const next = new Map<string, Point>();

    for (const card of el.querySelectorAll<HTMLElement>("[data-card-id]")) {
      const id = card.dataset.cardId;
      if (!id || typeof card.animate !== "function") continue;
      // A card still gliding starts from where it is on screen now, not where it was headed.
      const running = card.getAnimations();
      const seen = card.getBoundingClientRect();
      running.forEach((a) => a.cancel());
      const box = running.length ? card.getBoundingClientRect() : seen;
      const at = { x: box.left - base.left, y: box.top - base.top };
      const from = running.length ? { x: seen.left - base.left, y: seen.top - base.top } : last.current.get(id);
      next.set(id, at);

      if (!from || duration === 0) continue;
      const dx = from.x - at.x;
      const dy = from.y - at.y;
      if (Math.abs(dx) < 1 && Math.abs(dy) < 1) continue;
      card.animate([{ transform: `translate(${dx}px, ${dy}px)` }, { transform: "none" }], { duration, easing });
    }
    last.current = next;
  });
}
