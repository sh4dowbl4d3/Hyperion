export type ApiErrorBody = {
  code: string;
  message: string;
};

export class ApiError extends Error {
  readonly status: number;
  readonly code: string;

  constructor(status: number, body: ApiErrorBody) {
    super(body.message);
    this.name = "ApiError";
    this.status = status;
    this.code = body.code;
  }
}

const API_BASE = "/api/v1";

let accessTokenProvider: () => string | null = () => null;

export function setAccessTokenProvider(provider: () => string | null): void {
  accessTokenProvider = provider;
}

type RequestOptions = {
  method?: "GET" | "POST" | "PUT" | "PATCH" | "DELETE";
  body?: unknown;
  signal?: AbortSignal;
  headers?: Record<string, string>;
};

export async function apiFetch<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const { method = "GET", body, signal, headers: extraHeaders } = options;

  const headers: Record<string, string> = { Accept: "application/json", ...extraHeaders };
  if (body !== undefined) headers["Content-Type"] = "application/json";

  const token = accessTokenProvider();
  if (token) headers["Authorization"] = `Bearer ${token}`;

  let response: Response;
  try {
    response = await fetch(`${API_BASE}${path}`, {
      method,
      headers,
      body: body !== undefined ? JSON.stringify(body) : undefined,
      signal,
    });
  } catch (cause) {
    if (cause instanceof DOMException && cause.name === "AbortError") throw cause;
    throw new ApiError(0, { code: "network_error", message: "Unable to reach the server." });
  }

  if (response.status === 204) return undefined as T;

  const contentType = response.headers.get("Content-Type") ?? "";
  const payload: unknown = contentType.includes("application/json")
    ? await response.json().catch(() => null)
    : await response.text();

  if (!response.ok) {
    const envelope = payload as { error?: ApiErrorBody } | null;
    if (envelope?.error?.code) {
      throw new ApiError(response.status, envelope.error);
    }
    throw new ApiError(response.status, {
      code: "unexpected_response",
      message: `Request failed with status ${response.status}.`,
    });
  }

  return payload as T;
}

export function isApiError(value: unknown): value is ApiError {
  return value instanceof ApiError;
}
