const TOKEN_KEY = "hyperion.session";

export function loadSessionToken(): string | null {
  return window.sessionStorage.getItem(TOKEN_KEY);
}

export function saveSessionToken(token: string): void {
  window.sessionStorage.setItem(TOKEN_KEY, token);
}

export function clearSessionToken(): void {
  window.sessionStorage.removeItem(TOKEN_KEY);
}
