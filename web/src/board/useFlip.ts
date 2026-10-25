import { useLayoutEffect, useRef, type RefObject } from "react";

type Point = { x: number; y: number };

/**
 * Cards glide to their new place when they change column or order (DESIGN_RULES.md §8),
 * instead of vanishing and reappearing. Runs after every render of the container.
 *
 * Reads and writes are batched (read all, cancel all, read all, animate all) so the
 * browser lays the board out twice per render, not twice per card.
 */
export function useFlip(root: RefObject<HTMLElement | null>) {
  const last = useRef(new Map<string, Point>());

  useLayoutEffect(() => {
    const el = root.current;
    if (!el) return;
    const style = getComputedStyle(el);
    const duration = parseFloat(style.getPropertyValue("--dur-base")) || 0;
    const easing = style.getPropertyValue("--ease-out").trim() || "ease-out";
    const cards = [...el.querySelectorAll<HTMLElement>("[data-card-id]")].filter((c) => typeof c.animate === "function");

    // 1. Read: where each card is on screen now, mid-glide included.
    const base = el.getBoundingClientRect();
    const seen = cards.map((card) => {
      const r = card.getBoundingClientRect();
      return { x: r.left - base.left, y: r.top - base.top, moving: card.getAnimations().length > 0 };
    });

    // 2. Write: stop every glide so the next read gives the resting place.
    cards.forEach((card) => card.getAnimations().forEach((a) => a.cancel()));

    // 3. Read: resting place in the new layout.
    const rest = cards.map((card) => {
      const r = card.getBoundingClientRect();
      return { x: r.left - base.left, y: r.top - base.top };
    });

    // 4. Write: glide from where the card was to where it rests.
    const next = new Map<string, Point>();
    cards.forEach((card, i) => {
      const id = card.dataset.cardId;
      const at = rest[i];
      const now = seen[i];
      if (!id || !at || !now) return;
      next.set(id, at);
      const from = now.moving ? now : last.current.get(id);
      if (!from || duration === 0) return;
      const dx = from.x - at.x;
      const dy = from.y - at.y;
      if (Math.abs(dx) < 1 && Math.abs(dy) < 1) return;
      card.animate([{ transform: `translate(${dx}px, ${dy}px)` }, { transform: "none" }], { duration, easing });
    });
    last.current = next;
  });
}
