import { describe, expect, it, vi } from "vitest";
import React from "react";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

vi.mock("./api", () => ({
  api: vi.fn(async (route: string) => ({ route })),
}));

import queryBuilder from "./query";
import { api } from "./api";

function wrapper({ children }: { children: React.ReactNode }) {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
}

describe("queryBuilder", () => {
  it("builds a hook that fetches the given route", async () => {
    const useOrders = queryBuilder(["orders"], "orders?status=all");

    const { result } = renderHook(() => useOrders(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual({ route: "orders?status=all" });
    expect(api).toHaveBeenCalledWith("orders?status=all");
  });

  it("surfaces api errors through react-query", async () => {
    vi.mocked(api).mockRejectedValueOnce(new Error("boom"));
    const useBroken = queryBuilder(["broken"], "broken");

    const { result } = renderHook(() => useBroken(), { wrapper });

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect((result.current.error as Error).message).toBe("boom");
  });
});
