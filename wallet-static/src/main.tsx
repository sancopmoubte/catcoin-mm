import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import App from "./App";
import { appBasePath } from "./deployment";
import "./styles.css";

if ("serviceWorker" in navigator) {
  window.addEventListener("load", () =>
    navigator.serviceWorker
      .register(`${appBasePath}sw.js`, { scope: appBasePath })
      .catch(() => undefined),
  );
}

createRoot(document.getElementById("root")!).render(<StrictMode><App /></StrictMode>);
