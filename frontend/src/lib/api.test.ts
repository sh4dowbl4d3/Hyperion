import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { ApiError, apiFetch, setAccessTokenProvider } from "./api";

describe("apiFetch", () => {
  beforeEach(() => {
    setAccessTokenProvider(() => null);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  function mockFetchOnce(response: Response) {
    const fetchMock = vi.fn().mockResolvedValue(response);
    vi.stubGlobal("fetch", fetchMock);
    return fetchMock;
  }

  function jsonResponse(status: number, body: unknown, headers: Record<string, string> = {}) {
    return new Response(JSON.stringify(body), {
      status,
      headers: { "Content-Type": "application/json", ...headers },
    });
  }

  it("sends GET without auth header by default and parses JSON", async () => {
    const fetchMock = mockFetchOnce(jsonResponse(200, { ok: true }));
    const data = await apiFetch<{ ok: boolean }>("/healthz");
    expect(data).toEqual({ ok: true });
    const [url, init] = fetchMock.mock.calls[0];
    expect(url).toBe("/api/v1/healthz");
    expect(init.method).toBe("GET");
    expect((init.headers as Record<string, string>)["Authorization"]).toBeUndefined();
  });

  it("attaches bearer token when a provider is registered", async () => {
    setAccessTokenProvider(() => "token-abc");
    const fetchMock = mockFetchOnce(jsonResponse(200, {}));
    await apiFetch("/auth/me");
    const [, init] = fetchMock.mock.calls[0];
    expect((init.headers as Record<string, string>)["Authorization"]).toBe("Bearer token-abc");
  });

  it("serializes JSON bodies with content type", async () => {
    const fetchMock = mockFetchOnce(jsonResponse(201, { user: {} }));
    await apiFetch("/auth/register", { method: "POST", body: { email: "a@b.c" } });
    const [, init] = fetchMock.mock.calls[0];
    expect(init.method).toBe("POST");
    expect(init.body).toBe(JSON.stringify({ email: "a@b.c" }));
    expect((init.headers as Record<string, string>)["Content-Type"]).toBe("application/json");
  });

  it("throws ApiError from the backend error envelope", async () => {
    mockFetchOnce(
      jsonResponse(409, { error: { code: "email_taken", message: "this email is already registered" } }),
    );
    const err = (await apiFetch("/auth/register", { method: "POST", body: {} }).catch((e) => e)) as ApiError;
    expect(err).toBeInstanceOf(ApiError);
    expect(err.status).toBe(409);
    expect(err.code).toBe("email_taken");
    expect(err.message).toBe("this email is already registered");
  });

  it("maps non-envelope failures to unexpected_response", async () => {
    mockFetchOnce(new Response("gateway timeout", { status: 504 }));
    const err = (await apiFetch("/anything").catch((e) => e)) as ApiError;
    expect(err).toBeInstanceOf(ApiError);
    expect(err.code).toBe("unexpected_response");
    expect(err.status).toBe(504);
  });

  it("maps network failures to network_error", async () => {
    vi.stubGlobal("fetch", vi.fn().mockRejectedValue(new TypeError("Failed to fetch")));
    const err = (await apiFetch("/anything").catch((e) => e)) as ApiError;
    expect(err).toBeInstanceOf(ApiError);
    expect(err.code).toBe("network_error");
  });
});
