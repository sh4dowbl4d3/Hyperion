import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter } from "react-router-dom";
import App from "./App";
import "./styles/global.css";

try {
  localStorage.removeItem("hyperion_theme");
} catch {
  // ignore
}
document.documentElement.setAttribute("data-theme", "dark");
document.documentElement.classList.add("dark");
document.documentElement.classList.remove("light");

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <BrowserRouter>
      <App />
    </BrowserRouter>
  </StrictMode>,
);
