import { afterEach, describe, expect, it, vi } from "vitest";

afterEach(() => {
  vi.unstubAllEnvs();
  vi.unstubAllGlobals();
  vi.resetModules();
});

describe("socketForWatchlist", () => {
  it("returns a MockSocket in mock mode", async () => {
    vi.stubEnv("NEXT_PUBLIC_MOCK_API", "true");
    vi.resetModules();
    const { socketForWatchlist } = await import("./socket");
    const { MockSocket } = await import("./mock");

    const socket = socketForWatchlist("wl_default");
    expect(socket).toBeInstanceOf(MockSocket);
    (socket as unknown as { close: () => void }).close();
  });

  it("opens a real WebSocket against the gateway otherwise", async () => {
    vi.stubEnv("NEXT_PUBLIC_MOCK_API", "false");
    const wsSpy = vi.fn();
    vi.stubGlobal(
      "WebSocket",
      class {
        url: string;
        constructor(url: string) {
          this.url = url;
          wsSpy(url);
        }
      }
    );
    vi.resetModules();
    const { socketForWatchlist } = await import("./socket");

    socketForWatchlist("wl_42");
    expect(wsSpy).toHaveBeenCalledWith("ws://localhost:8080/watchlists/wl_42");
  });
});
