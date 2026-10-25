export type MoveTarget = "top" | "up" | "down" | "bottom";

/**
 * Where a card lands when moved from its menu, expressed as the same reorder target
 * drag and drop uses. Null when the card is already there.
 */
export function moveTarget(
  ids: string[],
  id: string,
  to: MoveTarget,
): { target: string; position: "before" | "after"; index: number } | null {
  const from = ids.indexOf(id);
  if (from === -1) return null;
  const last = ids.length - 1;
  const index = to === "top" ? 0 : to === "up" ? from - 1 : to === "down" ? from + 1 : last;
  const target = ids[index];
  if (target === undefined || index === from) return null;
  return { target, position: index < from ? "before" : "after", index };
}
