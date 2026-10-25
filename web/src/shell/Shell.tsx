import { Columns3, FolderGit2, Server, Settings, type LucideIcon } from "lucide-react";
import type { ReactNode } from "react";
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

function RailItem({ label, icon: Icon, current = false }: { label: string; icon: LucideIcon; current?: boolean }) {
  const hint = current ? label : `${label} (not in this prototype)`;
  return (
    <TooltipTrigger delay={300}>
      <Focusable>
        <a className="rail-item" href={current ? "#main" : undefined} aria-current={current ? "page" : undefined} aria-disabled={current ? undefined : true} tabIndex={0}>
          <Icon size={18} strokeWidth={1.5} aria-hidden="true" />
          <span className="rail-label">{label}</span>
        </a>
      </Focusable>
      <Tooltip className="tooltip" placement="end" offset={8}>
        {hint}
      </Tooltip>
    </TooltipTrigger>
  );
}

export function Shell({ children, statusBar }: { children: ReactNode; statusBar: ReactNode }) {
  return (
    <div className="shell">
      <a className="skip-link" href="#main">
        Skip to content
      </a>
      <div className="phone-brand">
        <ProductMark />
        <span>Astra</span>
      </div>
      <nav className="rail" aria-label="Main">
        <span className="rail-brand" role="img" aria-label="Astra">
          <ProductMark />
        </span>
        <ul className="rail-items">
          <li>
            <RailItem label="Board" icon={Columns3} current />
          </li>
          <li>
            <RailItem label="Projects" icon={FolderGit2} />
          </li>
          <li>
            <RailItem label="Servers" icon={Server} />
          </li>
          <li className="rail-settings">
            <RailItem label="Settings" icon={Settings} />
          </li>
        </ul>
      </nav>
      <main id="main" className="shell-main" tabIndex={-1}>
        {children}
      </main>
      <footer className="status-bar">{statusBar}</footer>
    </div>
  );
}
