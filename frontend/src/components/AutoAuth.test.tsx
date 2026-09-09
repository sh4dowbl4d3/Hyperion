import { cleanup, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { AutoAuth } from "./AutoAuth";
import { AuthProvider } from "../hooks/useAuth";
import * as authService from "../services/authService";

const mockUser = {
  id: "u-demo",
  email: "demo@hyperion.test",
  role: "user",
  created_at: "2026-01-01T00:00:00Z",
};

describe("AutoAuth", () => {
  beforeEach(() => {
    window.sessionStorage.clear();
    vi.restoreAllMocks();
  });

  afterEach(() => {
    cleanup();
  });

  it("renders children when already authenticated", async () => {
    window.sessionStorage.setItem("hyperion.session", "demo-token");
    vi.spyOn(authService, "me").mockResolvedValue({ user: mockUser });

    render(
      <AuthProvider>
        <AutoAuth>
          <div data-testid="auth-content">Content Visible</div>
        </AutoAuth>
      </AuthProvider>,
    );

    await waitFor(() => {
      expect(screen.queryByTestId("auth-content")).not.toBeNull();
    });
  });

  it("automatically logs in when unauthenticated", async () => {
    const loginSpy = vi.spyOn(authService, "login").mockResolvedValue({
      token: "auto-demo-token",
      expires_at: "2026-01-01T02:00:00Z",
      user: mockUser,
    });

    render(
      <AuthProvider>
        <AutoAuth>
          <div data-testid="guest-content">Auto Auth Success</div>
        </AutoAuth>
      </AuthProvider>,
    );

    await waitFor(() => {
      expect(loginSpy).toHaveBeenCalledWith("demo@hyperion.test", "Password123!");
    });

    await waitFor(() => {
      expect(screen.queryByTestId("guest-content")).not.toBeNull();
    });
  });
});
