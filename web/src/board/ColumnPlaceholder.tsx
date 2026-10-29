/** A column before its facts arrive: flat bars the size of two cards, so nothing shifts when they do (DESIGN_SYSTEM.md §8.15). */
export function ColumnPlaceholder() {
  return (
    <div className="column-placeholder" aria-hidden="true">
      <span className="column-placeholder-bar" />
      <span className="column-placeholder-bar" />
    </div>
  );
}
