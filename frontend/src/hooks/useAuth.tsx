import { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState } from "react";
import type { ReactNode } from "react";
import { setAccessTokenProvider } from "../lib/api";
import { clearSessionToken, loadSessionToken, saveSessionToken } from "../lib/session";
import * as authService from "../services/authService";
import type { AuthUser } from "../services/authService";

export type AuthStatus = "loading" | "authenticated" | "unauthenticated";

type AuthContextValue = {
  user: AuthUser | null;
  status: AuthStatus;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string) => Promise<void>;
  logout: () => void;
};

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [status, setStatus] = useState<AuthStatus>("loading");
  const tokenRef = useRef<string | null>(loadSessionToken());

  useEffect(() => {
    setAccessTokenProvider(() => tokenRef.current);
    let cancelled = false;

    if (!tokenRef.current) {
      setStatus("unauthenticated");
      return;
    }
    authService
      .me()
      .then((res) => {
        if (cancelled) return;
        setUser(res.user);
        setStatus("authenticated");
      })
      .catch(() => {
        if (cancelled) return;
        clearSessionToken();
        tokenRef.current = null;
        setStatus("unauthenticated");
      });

    return () => {
      cancelled = true;
    };
  }, []);

  const applySession = useCallback((token: string, nextUser: AuthUser) => {
    saveSessionToken(token);
    tokenRef.current = token;
    setUser(nextUser);
    setStatus("authenticated");
  }, []);

  const login = useCallback(
    async (email: string, password: string) => {
      const result = await authService.login(email.trim(), password);
      applySession(result.token, result.user);
    },
    [applySession],
  );

  const register = useCallback(
    async (email: string, password: string) => {
      await authService.register(email.trim(), password);
      const result = await authService.login(email.trim(), password);
      applySession(result.token, result.user);
    },
    [applySession],
  );

  const logout = useCallback(() => {
    clearSessionToken();
    tokenRef.current = null;
    setUser(null);
    setStatus("unauthenticated");
  }, []);

  const value = useMemo(
    () => ({ user, status, login, register, logout }),
    [user, status, login, register, logout],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth must be used inside an AuthProvider");
  }
  return ctx;
}
