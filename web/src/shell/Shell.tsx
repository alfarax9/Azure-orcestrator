import { Columns3, FolderGit2, PanelLeft, PanelRight, Server, Settings, type LucideIcon } from "lucide-react";
import { useEffect, useId, useState, type ReactNode } from "react";
import { Focusable, Tooltip, TooltipTrigger } from "react-aria-components";
import "./Shell.css";

function ProductMark() {
  // The mark is a short loop strip: done, done, running.
  return (
    <svg className="product-mark-svg" width="24" height="24" viewBox="0 0 24 24" aria-hidden="true">
      <rect x="1" y="11" width="6" height="2" rx="1" fill="currentColor" opacity="0.6" />
      <rect x="9" y="11" width="6" height="2" rx="1" fill="currentColor" opacity="0.6" />
      <rect x="17" y="11" width="6" height="2" rx="1" fill="var(--accent)" />
    </svg>
  );
}

function NavItem({ label, icon: Icon, current = false }: { label: string; icon: LucideIcon; current?: boolean }) {
  const hint = current ? label : `${label} (not in this prototype)`;
  return (
    <TooltipTrigger delay={300}>
      <Focusable>
        <a className="nav-item" href={current ? "#main" : undefined} aria-current={current ? "page" : undefined} aria-disabled={current ? undefined : true} tabIndex={0}>
          <Icon size={18} strokeWidth={1.5} aria-hidden="true" />
          <span className="nav-label">{label}</span>
        </a>
      </Focusable>
      <Tooltip className="tooltip" placement="top" offset={8}>
        {hint}
      </Tooltip>
    </TooltipTrigger>
  );
}

/**
 * The shell (DESIGN_SYSTEM.md §9.1, ADR 0009): a top bar and a status bar on the window, and three
 * floating panels between them: the sidebar (navigation and `sidebar`), the main area, and an optional
 * right panel. Below 1100px the `sidebar` becomes a drawer, and navigation a bottom bar.
 */
export function Shell({
  children,
  statusBar,
  sidebar,
  rightPanel,
  rightPanelLabel = "Details",
}: {
  children: ReactNode;
  statusBar: ReactNode;
  /** Shown under the navigation in the sidebar: the worker list. */
  sidebar?: ReactNode;
  /** The panel on the right (preview, inspector); the top bar gets a button for it. */
  rightPanel?: ReactNode;
  rightPanelLabel?: string;
}) {
  const drawerId = useId();
  const rightId = useId();
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [rightOpen, setRightOpen] = useState(true);

  useEffect(() => {
    if (!drawerOpen) return;
    const close = (e: KeyboardEvent) => {
      if (e.key === "Escape") setDrawerOpen(false);
    };
    window.addEventListener("keydown", close);
    return () => window.removeEventListener("keydown", close);
  }, [drawerOpen]);

  return (
    <div className="shell" data-drawer={drawerOpen ? "open" : "closed"} data-right={rightPanel && rightOpen ? "open" : "closed"}>
      <a className="skip-link" href="#main">
        Skip to content
      </a>
      <header className="topbar">
        {sidebar && (
          <button type="button" className="topbar-button topbar-drawer" aria-label="Workers" aria-expanded={drawerOpen} aria-controls={drawerId} onClick={() => setDrawerOpen((open) => !open)}>
            <PanelLeft size={16} strokeWidth={1.5} aria-hidden="true" />
          </button>
        )}
        <span className="topbar-brand">
          <ProductMark />
          <span>Astra</span>
        </span>
        {rightPanel && (
          <button type="button" className="topbar-button topbar-right" aria-label={rightPanelLabel} aria-expanded={rightOpen} aria-controls={rightId} onClick={() => setRightOpen((open) => !open)}>
            <PanelRight size={16} strokeWidth={1.5} aria-hidden="true" />
          </button>
        )}
      </header>
      <nav className="nav" aria-label="Main">
        <ul className="nav-items">
          <li>
            <NavItem label="Board" icon={Columns3} current />
          </li>
          <li>
            <NavItem label="Projects" icon={FolderGit2} />
          </li>
          <li>
            <NavItem label="Servers" icon={Server} />
          </li>
          <li className="nav-settings">
            <NavItem label="Settings" icon={Settings} />
          </li>
        </ul>
      </nav>
      {sidebar && (
        <>
          <button type="button" className="drawer-scrim" tabIndex={-1} aria-hidden="true" onClick={() => setDrawerOpen(false)} />
          <aside id={drawerId} className="panel sidebar" aria-label="Workers" onClick={(e) => e.target instanceof Element && e.target.closest("a") && setDrawerOpen(false)}>
            {sidebar}
          </aside>
        </>
      )}
      <main id="main" className="panel shell-main" tabIndex={-1}>
        {children}
      </main>
      {rightPanel && (
        <aside id={rightId} className="panel right-panel" aria-label={rightPanelLabel} hidden={!rightOpen}>
          {rightPanel}
        </aside>
      )}
      <footer className="status-bar">{statusBar}</footer>
    </div>
  );
}
