// Global styles load first so component styles can override them.
import "./styles/tokens.css";
import "./styles/fonts.css";
import "./styles/base.css";
import "./styles/controls.css";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { App } from "./App";

const root = document.getElementById("root");
if (!root) throw new Error("index.html is missing #root");

createRoot(root).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
