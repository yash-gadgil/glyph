import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { api, ApiError } from "./api";

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

describe("api (live mode)", () => {
  const fetchMock = vi.fn();

  beforeEach(() => {
    vi.stubEnv("NEXT_PUBLIC_MOCK_API", "false");
    vi.stubGlobal("fetch", fetchMock);
    fetchMock.mockReset();
  });

  afterEach(() => {
    vi.unstubAllEnvs();
    vi.unstubAllGlobals();
  });

  it("returns parsed JSON on success", async () => {
    fetchMock.mockResolvedValueOnce(jsonResponse({ hello: "world" }));

    const result = await api("account/me");

    expect(result).toEqual({ hello: "world" });
    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8080/account/me",
      expect.objectContaining({ credentials: "include" })
    );
  });

  it("sends JSON content type by default", async () => {
    fetchMock.mockResolvedValueOnce(jsonResponse({}));

    await api("orders", { method: "POST", body: "{}" });

    const init = fetchMock.mock.calls[0][1];
    expect(init.headers["Content-Type"]).toBe("application/json");
    expect(init.method).toBe("POST");
  });

  it("returns null for non-JSON responses", async () => {
    fetchMock.mockResolvedValueOnce(new Response("plain text", { status: 200 }));

    expect(await api("account/funds")).toBeNull();
  });

  it("throws ApiError with the server message on HTTP errors", async () => {
    fetchMock.mockResolvedValueOnce(jsonResponse({ message: "nope" }, 400));

    const err = await api("orders").catch((e) => e);
    expect(err).toBeInstanceOf(ApiError);
    expect(err.message).toBe("nope");
    expect(err.type).toBe("http");
    expect(err.statusCode).toBe(400);
  });

  it("falls back to a generic message when the error body is not JSON", async () => {
    fetchMock.mockResolvedValueOnce(new Response("boom", { status: 500 }));

    const err = await api("orders").catch((e) => e);
    expect(err.message).toContain("500");
  });

  it("throws a network ApiError when fetch rejects", async () => {
    fetchMock.mockRejectedValueOnce(new TypeError("fetch failed"));

    const err = await api("orders").catch((e) => e);
    expect(err).toBeInstanceOf(ApiError);
    expect(err.type).toBe("network");
  });

  it("refreshes the session on 401 and retries once", async () => {
    fetchMock
      .mockResolvedValueOnce(jsonResponse({ message: "unauthorized" }, 401))
      .mockResolvedValueOnce(jsonResponse({ ok: true }))
      .mockResolvedValueOnce(jsonResponse({ data: 42 }));

    const result = await api("portfolio");

    expect(result).toEqual({ data: 42 });
    expect(fetchMock).toHaveBeenCalledTimes(3);
    expect(fetchMock.mock.calls[1][0]).toBe("http://localhost:8080/auth/refresh");
  });

  it("gives up when the refresh also fails", async () => {
    fetchMock
      .mockResolvedValueOnce(jsonResponse({}, 401))
      .mockResolvedValueOnce(new Response(null, { status: 401 }));

    const err = await api("portfolio").catch((e) => e);
    expect(err).toBeInstanceOf(ApiError);
    expect(err.statusCode).toBe(401);
    expect(err.message).toContain("Session expired");
  });

  it("does not retry twice on repeated 401s", async () => {
    fetchMock
      .mockResolvedValueOnce(jsonResponse({}, 401))
      .mockResolvedValueOnce(jsonResponse({ ok: true }))
      .mockResolvedValueOnce(jsonResponse({ message: "still unauthorized" }, 401));

    const err = await api("portfolio").catch((e) => e);
    expect(err.statusCode).toBe(401);
    expect(fetchMock).toHaveBeenCalledTimes(3);
  });
});

describe("api (mock mode)", () => {
  beforeEach(() => {
    vi.stubEnv("NEXT_PUBLIC_MOCK_API", "true");
  });

  afterEach(() => {
    vi.unstubAllEnvs();
  });

  it("routes through the in-browser mock without fetch", async () => {
    const fetchSpy = vi.fn();
    vi.stubGlobal("fetch", fetchSpy);

    const result = await api("account/me");

    expect(result).toEqual({ id: "mock_user", user_id: "mock_user" });
    expect(fetchSpy).not.toHaveBeenCalled();
    vi.unstubAllGlobals();
  });

  it("wraps mock errors in ApiError", async () => {
    const err = await api("watchlists/wl_does_not_exist/info").catch((e) => e);
    expect(err).toBeInstanceOf(ApiError);
    expect(err.statusCode).toBe(404);
  });

  it("returns null for unknown routes, matching a 204-style response", async () => {
    expect(await api("definitely/not/a/route")).toBeNull();
  });
});
