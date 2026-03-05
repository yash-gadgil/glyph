import { beforeEach, describe, expect, it, vi } from "vitest";

async function freshMock() {
  vi.resetModules();
  return await import("./mock");
}

beforeEach(() => {
  window.localStorage.clear();
});

describe("isMockMode", () => {
  it("reflects the env flag", async () => {
    vi.stubEnv("NEXT_PUBLIC_MOCK_API", "true");
    let { isMockMode } = await freshMock();
    expect(isMockMode()).toBe(true);

    vi.stubEnv("NEXT_PUBLIC_MOCK_API", "false");
    ({ isMockMode } = await freshMock());
    expect(isMockMode()).toBe(false);
    vi.unstubAllEnvs();
  });
});

describe("mockApi auth + account", () => {
  it("signs in and returns the mock user", async () => {
    const { mockApi } = await freshMock();
    const res = await mockApi("auth/signin", {
      method: "POST",
      body: JSON.stringify({ email: "demo@glyph.dev" }),
    });
    expect(res.success).toBe(true);
    expect(res.email).toBe("demo@glyph.dev");
  });

  it("serves account endpoints in cents", async () => {
    const { mockApi } = await freshMock();
    expect((await mockApi("account/me", {})).id).toBe("mock_user");

    const account = await mockApi("account", {});
    expect(account.currency).toBe("USD");
    expect(account.cash_balance_cents).toBeGreaterThan(0);
    expect(account.equity_cents).toBeGreaterThanOrEqual(account.cash_balance_cents);
  });

  it("resets the account to the fixed starting balance", async () => {
    const { mockApi } = await freshMock();

    expect(
      await mockApi("account/funds", {
        method: "POST",
        body: JSON.stringify({ amount_cents: 50_000 }),
      })
    ).toBeNull();

    await mockApi("account/reset", { method: "POST" });
    const reset = await mockApi("account", {});
    expect(reset.cash_balance_cents).toBe(10_000_000);
    const positions = await mockApi("portfolio/positions", {});
    expect(positions.positions).toHaveLength(0);
  });

  it("returns null for unknown routes", async () => {
    const { mockApi } = await freshMock();
    expect(await mockApi("nope/nothing", {})).toBeNull();
  });

  it("404s a missing watchlist", async () => {
    const { mockApi } = await freshMock();
    const err = await mockApi("watchlists/wl_missing/info", {}).catch((e) => e);
    expect(err.statusCode).toBe(404);
  });
});

describe("mockApi watchlists", () => {
  it("lists seeded watchlists", async () => {
    const { mockApi } = await freshMock();
    const res = await mockApi("watchlists", {});
    expect(res.w_metadata.length).toBeGreaterThanOrEqual(2);
    expect(res.w_metadata[0]).toHaveProperty("id");
    expect(res.w_metadata[0]).toHaveProperty("name");
  });

  it("creates, modifies, and deletes a watchlist", async () => {
    const { mockApi } = await freshMock();

    const created = await mockApi("watchlists", {
      method: "POST",
      body: JSON.stringify({ name: "Earnings Week" }),
    });
    expect(created.name).toBe("Earnings Week");

    const subscribed = await mockApi(`watchlists/${created.id}?action=subscribe`, {
      method: "PATCH",
      body: JSON.stringify({ symbols: ["AAPL", "TSLA"] }),
    });
    expect(subscribed.symbols).toEqual(["AAPL", "TSLA"]);

    const again = await mockApi(`watchlists/${created.id}?action=subscribe`, {
      method: "PATCH",
      body: JSON.stringify({ symbols: ["AAPL"] }),
    });
    expect(again.symbols).toEqual(["AAPL", "TSLA"]);

    const unsubscribed = await mockApi(`watchlists/${created.id}?action=unsubscribe`, {
      method: "PATCH",
      body: JSON.stringify({ symbols: ["AAPL"] }),
    });
    expect(unsubscribed.symbols).toEqual(["TSLA"]);

    const info = await mockApi(`watchlists/${created.id}/info`, {});
    expect(info.symbols).toEqual(["TSLA"]);

    const deleted = await mockApi(`watchlists/${created.id}`, { method: "DELETE" });
    expect(deleted.success).toBe(true);

    const err = await mockApi(`watchlists/${created.id}/info`, {}).catch((e) => e);
    expect(err.statusCode).toBe(404);
  });

  it("persists watchlists to localStorage", async () => {
    const { mockApi } = await freshMock();
    await mockApi("watchlists", {
      method: "POST",
      body: JSON.stringify({ name: "Persisted" }),
    });

    const raw = window.localStorage.getItem("glyph-mock:watchlists");
    expect(raw).toBeTruthy();
    expect(JSON.parse(raw!).some((w: { name: string }) => w.name === "Persisted")).toBe(true);
  });

  it("searches symbols by ticker and company name", async () => {
    const { mockApi } = await freshMock();

    const byTicker = await mockApi("watchlists/symbols?q=AAPL", {});
    expect(byTicker.symbols.some((s: { name: string }) => s.name === "AAPL")).toBe(true);

    const byName = await mockApi("watchlists/symbols?q=tesla", {});
    expect(byName.symbols.some((s: { name: string }) => s.name === "TSLA")).toBe(true);

    const limited = await mockApi("watchlists/symbols?q=&limit=5", {});
    expect(limited.symbols).toHaveLength(5);
  });

  it("returns OHLC history bars for requested symbols", async () => {
    const { mockApi } = await freshMock();
    const res = await mockApi("watchlists/history", {
      method: "POST",
      body: JSON.stringify({ symbols: ["AAPL", "TSLA"] }),
    });

    expect(res.symbol_bars).toHaveLength(2);
    const bars = res.symbol_bars[0].bars;
    expect(bars.length).toBeGreaterThan(30);
    for (const bar of bars.slice(0, 5)) {
      expect(bar.high).toBeGreaterThanOrEqual(bar.low);
      expect(bar.high).toBeGreaterThanOrEqual(Math.min(bar.open, bar.close));
    }
  });
});

describe("mockApi orders", () => {
  it("places an order and lists it", async () => {
    const { mockApi } = await freshMock();

    const placed = await mockApi("orders", {
      method: "POST",
      body: JSON.stringify({
        symbol: "AAPL",
        side: "buy",
        orderType: "limit",
        timeInForce: "gtc",
        quantity: 5,
        price: 180,
      }),
    });

    expect(placed.order.symbol).toBe("AAPL");
    expect(placed.order.qty).toBe(5);
    expect(placed.order.price).toBe(180);
    expect(placed.order.status).toBe("pending");

    const list = await mockApi("orders", {});
    expect(list.orders.some((o: { id: string }) => o.id === placed.order.id)).toBe(true);

    const single = await mockApi(`orders/${placed.order.id}`, {});
    expect(single.id).toBe(placed.order.id);
  });

  it("defaults market order price to the symbol base price", async () => {
    const { mockApi } = await freshMock();
    const placed = await mockApi("orders", {
      method: "POST",
      body: JSON.stringify({ symbol: "AAPL", orderType: "market", quantity: 1 }),
    });
    expect(placed.order.price).toBe(18_950);
  });

  it("cancels an order", async () => {
    const { mockApi } = await freshMock();
    const placed = await mockApi("orders", {
      method: "POST",
      body: JSON.stringify({ symbol: "TSLA", quantity: 2 }),
    });

    await mockApi(`orders/${placed.order.id}`, { method: "DELETE" });

    const after = await mockApi(`orders/${placed.order.id}`, {});
    expect(after.status).toBe("cancelled");
  });

  it("filters orders by status", async () => {
    const { mockApi } = await freshMock();
    const placed = await mockApi("orders", {
      method: "POST",
      body: JSON.stringify({ symbol: "AAPL", quantity: 1 }),
    });
    await mockApi(`orders/${placed.order.id}`, { method: "DELETE" });

    const cancelled = await mockApi("orders?status=cancelled", {});
    expect(cancelled.orders.every((o: { status: string }) => o.status === "cancelled")).toBe(true);
    expect(cancelled.orders.length).toBeGreaterThanOrEqual(1);
  });

  it("404s unknown order ids", async () => {
    const { mockApi } = await freshMock();
    const err = await mockApi("orders/ord_missing", {}).catch((e) => e);
    expect(err.statusCode).toBe(404);
  });
});

describe("mockApi portfolio", () => {
  it("returns portfolio, holdings and positions", async () => {
    const { mockApi } = await freshMock();

    const portfolio = await mockApi("portfolio", {});
    expect(portfolio.currency).toBe("USD");
    expect(portfolio.cash_balance_cents).toBeGreaterThan(0);
    expect(portfolio.buying_power_cents).toBe(
      portfolio.cash_balance_cents - portfolio.reserved_cash_cents
    );

    const holdings = await mockApi("portfolio/holdings", {});
    expect(holdings.holdings.length).toBeGreaterThan(0);
    expect(holdings.holdings[0]).toHaveProperty("last_price_cents");
    expect(holdings.holdings[0]).toHaveProperty("unrealized_pnl_cents");
    expect(holdings.total_market_value_cents).toBeGreaterThan(0);

    const positions = await mockApi("portfolio/positions", {});
    expect(positions.positions.length).toBeGreaterThan(0);
    expect(positions.positions[0]).toHaveProperty("cost_basis_cents");
  });

  it("a filled order settles cash and positions", async () => {
    vi.useFakeTimers();
    try {
      const { mockApi } = await freshMock();

      const call = async (path: string, options: RequestInit = {}) => {
        const pending = mockApi(path, options);
        await vi.advanceTimersByTimeAsync(200);
        return pending;
      };

      const before = await call("account");
      const placed = await call("orders", {
        method: "POST",
        body: JSON.stringify({
          symbol: "GOOGL",
          side: "buy",
          orderType: "limit",
          quantity: 2,
          price: 17_000,
        }),
      });

      await vi.advanceTimersByTimeAsync(2_000);

      const order = await call(`orders/${placed.order.id}`);
      expect(order.status).toBe("filled");

      const after = await call("account");
      expect(after.cash_balance_cents).toBe(before.cash_balance_cents - 2 * 17_000);

      const positions = await call("portfolio/positions");
      const googl = positions.positions.find((p: { symbol: string }) => p.symbol === "GOOGL");
      expect(googl.qty).toBe(2);
      expect(googl.cost_basis_cents).toBe(2 * 17_000);

      const trades = await call("account/trades");
      expect(trades.fills.some((f: { order_id: string }) => f.order_id === placed.order.id)).toBe(true);
    } finally {
      vi.useRealTimers();
    }
  });
});

describe("mockApi explore + strategies", () => {
  it("serves news and movers", async () => {
    const { mockApi } = await freshMock();

    const news = await mockApi("explore/news", {});
    expect(news.articles.length).toBeGreaterThan(0);
    expect(news.articles[0]).toHaveProperty("headline");
    expect(news.articles[0]).toHaveProperty("created_at");

    const movers = await mockApi("explore/movers", {});
    expect(movers.gainers.length).toBeGreaterThan(0);
    expect(movers.losers.length).toBeGreaterThan(0);
    expect(movers.gainers[0].price_cents).toBeGreaterThan(0);
  });

  it("persists strategies through CRUD", async () => {
    const { mockApi } = await freshMock();

    const created = await mockApi("strategies", {
      method: "POST",
      body: JSON.stringify({ name: "Momentum", config_json: { risk: "high" } }),
    });
    expect(created.id).toBeTruthy();
    expect(created.name).toBe("Momentum");

    const dup = await mockApi("strategies", {
      method: "POST",
      body: JSON.stringify({ name: "Momentum", config_json: {} }),
    }).catch((e) => e);
    expect(dup.statusCode).toBe(409);

    const list = await mockApi("strategies", {});
    expect(list.strategies).toHaveLength(1);

    const updated = await mockApi(`strategies/${created.id}`, {
      method: "PATCH",
      body: JSON.stringify({ name: "Momentum v2", config_json: { risk: "low" } }),
    });
    expect(updated.name).toBe("Momentum v2");

    await mockApi(`strategies/${created.id}`, { method: "DELETE" });
    const after = await mockApi("strategies", {});
    expect(after.strategies).toHaveLength(0);
  });
});
