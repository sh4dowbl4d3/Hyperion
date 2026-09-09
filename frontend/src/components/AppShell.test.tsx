import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";
import { AppShell } from "./AppShell";
import { AuthProvider } from "../hooks/useAuth";

describe("AppShell", () => {
  it("does not render a footer on the dashboard", () => {
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

  it("does not render a footer on the labs page", () => {
    const { container } = render(
      <AuthProvider>
        <MemoryRouter initialEntries={["/labs"]}>
          <AppShell />
        </MemoryRouter>
      </AuthProvider>,
    );

    expect(container.querySelector("footer")).toBeNull();
  });

  it("does not render a footer on root path (/)", () => {
    const { container } = render(
      <AuthProvider>
        <MemoryRouter initialEntries={["/"]}>
          <AppShell />
        </MemoryRouter>
      </AuthProvider>,
    );

    expect(container.querySelector("footer")).toBeNull();
  });
});
