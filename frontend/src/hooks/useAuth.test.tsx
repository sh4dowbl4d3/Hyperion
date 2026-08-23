import { act, renderHook, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { ReactNode } from "react";
import { AuthProvider, useAuth } from "./useAuth";
import * as authService from "../services/authService";

const testUser = {
  id: "u-1",
  email: "student@example.com",
  role: "user",
  created_at: "2026-01-01T00:00:00Z",
};

function wrapper({ children }: { children: ReactNode }) {
  return <AuthProvider>{children}</AuthProvider>;
}

beforeEach(() => {
  window.sessionStorage.clear();
  vi.restoreAllMocks();
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("AuthProvider", () => {
  it("reports unauthenticated when no stored session exists", async () => {
    const { result } = renderHook(() => useAuth(), { wrapper });
    await waitFor(() => expect(result.current.status).toBe("unauthenticated"));
    expect(result.current.user).toBeNull();
  });

  it("hydrates the user from a stored token and confirms authentication", async () => {
    window.sessionStorage.setItem("hyperion.session", "stored-token");
    const meMock = vi.spyOn(authService, "me").mockResolvedValue({ user: testUser });

    const { result } = renderHook(() => useAuth(), { wrapper });

    expect(result.current.status).toBe("loading");
    await waitFor(() => expect(result.current.status).toBe("authenticated"));
    expect(result.current.user?.email).toBe(testUser.email);
    expect(meMock).toHaveBeenCalledTimes(1);
  });

  it("clears an invalid stored session", async () => {
    window.sessionStorage.setItem("hyperion.session", "expired-token");
    vi.spyOn(authService, "me").mockRejectedValue(
      Object.assign(new Error("expired"), { code: "token_expired" }),
    );

    const { result } = renderHook(() => useAuth(), { wrapper });
    await waitFor(() => expect(result.current.status).toBe("unauthenticated"));
    expect(window.sessionStorage.getItem("hyperion.session")).toBeNull();
  });

  it("login stores the token and sets the user", async () => {
    vi.spyOn(authService, "login").mockResolvedValue({
      token: "fresh-token",
      expires_at: "2026-01-01T02:00:00Z",
      user: testUser,
    });

    const { result } = renderHook(() => useAuth(), { wrapper });
    await waitFor(() => expect(result.current.status).toBe("unauthenticated"));

    await act(async () => {
      await result.current.login("student@example.com", "long-enough-password");
    });

    expect(result.current.status).toBe("authenticated");
    expect(window.sessionStorage.getItem("hyperion.session")).toBe("fresh-token");
  });

  it("register signs in after account creation", async () => {
    const registerMock = vi
      .spyOn(authService, "register")
      .mockResolvedValue({ user: testUser });
    const loginMock = vi.spyOn(authService, "login").mockResolvedValue({
      token: "auto-token",
      expires_at: "2026-01-01T02:00:00Z",
      user: testUser,
    });

    const { result } = renderHook(() => useAuth(), { wrapper });
    await waitFor(() => expect(result.current.status).toBe("unauthenticated"));

    await act(async () => {
      await result.current.register("student@example.com", "long-enough-password");
    });

    expect(registerMock).toHaveBeenCalledTimes(1);
    expect(loginMock).toHaveBeenCalledTimes(1);
    expect(result.current.user).toEqual(testUser);
  });

  it("logout clears token and user state", async () => {
    window.sessionStorage.setItem("hyperion.session", "stored-token");
    vi.spyOn(authService, "me").mockResolvedValue({ user: testUser });

    const { result } = renderHook(() => useAuth(), { wrapper });
    await waitFor(() => expect(result.current.status).toBe("authenticated"));

    act(() => {
      result.current.logout();
    });

    expect(result.current.status).toBe("unauthenticated");
    expect(result.current.user).toBeNull();
    expect(window.sessionStorage.getItem("hyperion.session")).toBeNull();
  });
});
