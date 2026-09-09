import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";
import { AppShell } from "./AppShell";
import { AuthProvider } from "../hooks/useAuth";

describe("AppShell", () => {
  it("does not render the footer on the homepage (/dashboard)", () => {
    const { container } = render(
      <AuthProvider>
        <MemoryRouter initialEntries={["/dashboard"]}>
          <AppShell />
        </MemoryRouter>
      </AuthProvider>,
    );

    expect(container.querySelector("footer")).toBeNull();
    expect(
      screen.queryByText(/LOCAL ENVIRONMENT · SYNTHETIC DATA · NO EXTERNAL TARGETS/i),
    ).toBeNull();
  });

  it("does not render the footer on root path (/)", () => {
    const { container } = render(
      <AuthProvider>
        <MemoryRouter initialEntries={["/"]}>
          <AppShell />
        </MemoryRouter>
      </AuthProvider>,
    );

    expect(container.querySelector("footer")).toBeNull();
  });

  it("renders the new footer with navigation and branding on non-homepage routes like /labs", () => {
    const { container } = render(
      <AuthProvider>
        <MemoryRouter initialEntries={["/labs"]}>
          <AppShell />
        </MemoryRouter>
      </AuthProvider>,
    );

    // Old marquee yellow banner should NOT exist
    expect(
      screen.queryByText(/LOCAL ENVIRONMENT · SYNTHETIC DATA · NO EXTERNAL TARGETS/i),
    ).toBeNull();

    // New footer should contain Hyperion branding, version, and copyright
    const footer = container.querySelector("footer");
    expect(footer).not.toBeNull();
    expect(footer?.className).not.toContain("bg-canary-banner");
    expect(footer?.textContent).toContain("Hyperion");
    expect(footer?.textContent).toContain("v0.1.0");
    expect(footer?.textContent).toContain("© 2026 Hyperion Security");
    expect(footer?.textContent).toContain("LOCAL SANDBOX · SYNTHETIC TARGETS");
  });
});
