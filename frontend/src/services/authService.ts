import { apiFetch } from "../lib/api";

export type AuthUser = {
  id: string;
  email: string;
  role: string;
  created_at: string;
};

type UserResponse = { user: AuthUser };

export type LoginResult = {
  token: string;
  expires_at: string;
  user: AuthUser;
};

export function register(email: string, password: string): Promise<UserResponse> {
  return apiFetch<UserResponse>("/auth/register", {
    method: "POST",
    body: { email, password },
  });
}

export function login(email: string, password: string): Promise<LoginResult> {
  return apiFetch<LoginResult>("/auth/login", {
    method: "POST",
    body: { email, password },
  });
}

export function me(): Promise<UserResponse> {
  return apiFetch<UserResponse>("/auth/me");
}
