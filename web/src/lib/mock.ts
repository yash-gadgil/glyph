
export function isMockMode(): boolean {
  return process.env.NEXT_PUBLIC_MOCK_API === "true";
}


const LS_PREFIX = "glyph-mock:";

function canUseLS(): boolean {
  return typeof window !== "undefined" && !!window.localStorage;
}

function load<T>(key: string, fallback: T): T {
  if (!canUseLS()) return fallback;
  try {
    const raw = window.localStorage.getItem(LS_PREFIX + key);
    return raw ? (JSON.parse(raw) as T) : fallback;
  } catch {
    return fallback;
  }
}

function save<T>(key: string, value: T) {
  if (!canUseLS()) return;
  try {
    window.localStorage.setItem(LS_PREFIX + key, JSON.stringify(value));
  } catch {}
}


type SymbolEntry = { name: string; company_name: string; basePrice: number };

const SYMBOL_POOL: SymbolEntry[] = [
  { name: "AAPL", company_name: "Apple Inc.", basePrice: 189.5 },
  { name: "MSFT", company_name: "Microsoft Corporation", basePrice: 418.2 },
  { name: "NVDA", company_name: "NVIDIA Corporation", basePrice: 875.3 },
  { name: "GOOGL", company_name: "Alphabet Inc.", basePrice: 172.1 },
  { name: "GOOG", company_name: "Alphabet Inc. Class C", basePrice: 173.8 },
  { name: "AMZN", company_name: "Amazon.com, Inc.", basePrice: 184.4 },
  { name: "META", company_name: "Meta Platforms, Inc.", basePrice: 502.1 },
  { name: "TSLA", company_name: "Tesla, Inc.", basePrice: 248.9 },
  { name: "AVGO", company_name: "Broadcom Inc.", basePrice: 1420.5 },
  { name: "BRK.B", company_name: "Berkshire Hathaway Class B", basePrice: 405.2 },
  { name: "JPM", company_name: "JPMorgan Chase & Co.", basePrice: 198.3 },
  { name: "V", company_name: "Visa Inc.", basePrice: 278.4 },
  { name: "MA", company_name: "Mastercard Incorporated", basePrice: 457.8 },
  { name: "UNH", company_name: "UnitedHealth Group", basePrice: 518.2 },
  { name: "JNJ", company_name: "Johnson & Johnson", basePrice: 148.9 },
  { name: "XOM", company_name: "Exxon Mobil Corporation", basePrice: 116.7 },
  { name: "PG", company_name: "Procter & Gamble Co.", basePrice: 160.4 },
  { name: "HD", company_name: "The Home Depot, Inc.", basePrice: 342.1 },
  { name: "CVX", company_name: "Chevron Corporation", basePrice: 158.9 },
  { name: "KO", company_name: "The Coca-Cola Company", basePrice: 62.3 },
  { name: "PEP", company_name: "PepsiCo, Inc.", basePrice: 171.5 },
  { name: "LLY", company_name: "Eli Lilly and Company", basePrice: 780.2 },
  { name: "COST", company_name: "Costco Wholesale Corporation", basePrice: 735.9 },
  { name: "WMT", company_name: "Walmart Inc.", basePrice: 60.1 },
  { name: "MRK", company_name: "Merck & Co., Inc.", basePrice: 128.4 },
  { name: "ABBV", company_name: "AbbVie Inc.", basePrice: 167.2 },
  { name: "BAC", company_name: "Bank of America Corporation", basePrice: 39.8 },
  { name: "ADBE", company_name: "Adobe Inc.", basePrice: 482.6 },
  { name: "CRM", company_name: "Salesforce, Inc.", basePrice: 274.3 },
  { name: "AMD", company_name: "Advanced Micro Devices", basePrice: 158.7 },
  { name: "NFLX", company_name: "Netflix, Inc.", basePrice: 626.4 },
  { name: "DIS", company_name: "The Walt Disney Company", basePrice: 108.9 },
  { name: "INTC", company_name: "Intel Corporation", basePrice: 31.2 },
  { name: "ORCL", company_name: "Oracle Corporation", basePrice: 118.5 },
  { name: "CSCO", company_name: "Cisco Systems, Inc.", basePrice: 48.1 },
  { name: "QCOM", company_name: "QUALCOMM Incorporated", basePrice: 172.6 },
  { name: "TXN", company_name: "Texas Instruments", basePrice: 194.8 },
  { name: "IBM", company_name: "IBM Corporation", basePrice: 172.9 },
  { name: "UBER", company_name: "Uber Technologies, Inc.", basePrice: 68.5 },
  { name: "SHOP", company_name: "Shopify Inc.", basePrice: 72.1 },
  { name: "SQ", company_name: "Block, Inc.", basePrice: 71.3 },
  { name: "PYPL", company_name: "PayPal Holdings, Inc.", basePrice: 64.8 },
  { name: "COIN", company_name: "Coinbase Global, Inc.", basePrice: 218.9 },
  { name: "SPOT", company_name: "Spotify Technology S.A.", basePrice: 301.4 },
  { name: "PLTR", company_name: "Palantir Technologies Inc.", basePrice: 22.8 },
  { name: "SNOW", company_name: "Snowflake Inc.", basePrice: 156.2 },
  { name: "RBLX", company_name: "Roblox Corporation", basePrice: 38.4 },
  { name: "RIVN", company_name: "Rivian Automotive, Inc.", basePrice: 11.2 },
  { name: "LCID", company_name: "Lucid Group, Inc.", basePrice: 2.8 },
  { name: "F", company_name: "Ford Motor Company", basePrice: 12.4 },
];

const SYMBOL_MAP: Record<string, SymbolEntry> = Object.fromEntries(
  SYMBOL_POOL.map((s) => [s.name, s])
);

function basePrice(symbol: string): number {
  return SYMBOL_MAP[symbol]?.basePrice ?? 100;
}


type Watchlist = { id: string; name: string; symbols: string[] };

type OrderResponse = {
  id: string;
  userId: string;
  symbol: string;
  side: string;
  orderType: string;
  timeInForce: string;
  qty: number;
  filledQty: number;
  price: number;
  stopPrice: number;
  status: string;
  createdAt: string;
  updatedAt: string;
};

type Position = {
  symbol: string;
  qty: number;
  reserved_qty: number;
  cost_basis_cents: number;
  realized_pnl_cents: number;
};

type Fill = {
  trade_id: string;
  order_id: string;
  symbol: string;
  side: string;
  qty: number;
  price_cents: number;
  liquidity: string;
  executed_at: string;
};

type MockAccount = {
  cash_balance_cents: number;
  reserved_cash_cents: number;
};

type StoredStrategy = {
  id: string;
  name: string;
  config_json: string;
  created_at: string;
  updated_at: string;
};


const DEFAULT_WATCHLISTS: Watchlist[] = [
  { id: "wl_default", name: "My Watchlist", symbols: ["AAPL", "MSFT", "NVDA", "TSLA", "GOOGL"] },
  { id: "wl_tech", name: "Tech Giants", symbols: ["META", "AMZN", "GOOG", "NFLX", "ADBE"] },
];

const DEFAULT_POSITIONS: Position[] = [
  { symbol: "AAPL", qty: 25, reserved_qty: 0, cost_basis_cents: 25 * 17_230, realized_pnl_cents: 0 },
  { symbol: "MSFT", qty: 10, reserved_qty: 0, cost_basis_cents: 10 * 39_050, realized_pnl_cents: 0 },
  { symbol: "TSLA", qty: 15, reserved_qty: 0, cost_basis_cents: 15 * 21_010, realized_pnl_cents: 124_055 },
  { symbol: "NVDA", qty: 5, reserved_qty: 0, cost_basis_cents: 5 * 71_280, realized_pnl_cents: 0 },
];

const DEFAULT_ACCOUNT: MockAccount = {
  cash_balance_cents: 10_000_000,
  reserved_cash_cents: 0,
};

let watchlists: Watchlist[] = load("watchlists", DEFAULT_WATCHLISTS);
let orders: OrderResponse[] = load("orders", []);
let positions: Position[] = load("positions", DEFAULT_POSITIONS);
let account: MockAccount = load("account", DEFAULT_ACCOUNT);
let fills: Fill[] = load("fills", []);
let strategies: StoredStrategy[] = load("strategies", []);

function persistWatchlists() { save("watchlists", watchlists); }
function persistOrders() { save("orders", orders); }
function persistPositions() { save("positions", positions); }
function persistAccount() { save("account", account); }
function persistFills() { save("fills", fills); }
function persistStrategies() { save("strategies", strategies); }

function basePriceCents(symbol: string): number {
  return Math.round(basePrice(symbol) * 100);
}

function settleMockFill(order: OrderResponse) {
  const qty = order.qty;
  const priceCents = order.price;
  let pos = positions.find((p) => p.symbol === order.symbol);
  if (!pos) {
    pos = { symbol: order.symbol, qty: 0, reserved_qty: 0, cost_basis_cents: 0, realized_pnl_cents: 0 };
    positions.push(pos);
  }

  if (order.side === "sell") {
    const sellQty = Math.min(qty, pos.qty);
    const costRemoved = pos.qty > 0 ? Math.round((pos.cost_basis_cents * sellQty) / pos.qty) : 0;
    const proceeds = sellQty * priceCents;
    pos.qty -= sellQty;
    pos.cost_basis_cents -= costRemoved;
    pos.realized_pnl_cents += proceeds - costRemoved;
    account.cash_balance_cents += proceeds;
  } else {
    pos.qty += qty;
    pos.cost_basis_cents += qty * priceCents;
    account.cash_balance_cents -= qty * priceCents;
  }

  fills = [{
    trade_id: uid("trd"),
    order_id: order.id,
    symbol: order.symbol,
    side: order.side,
    qty,
    price_cents: priceCents,
    liquidity: "taker",
    executed_at: nowIso(),
  }, ...fills];

  persistPositions();
  persistAccount();
  persistFills();
}

function holdingsPayload() {
  const holdings = positions
    .filter((p) => p.qty !== 0 || p.realized_pnl_cents !== 0)
    .map((p) => {
      const last = basePriceCents(p.symbol);
      const marketValue = p.qty !== 0 ? last * p.qty : p.cost_basis_cents;
      return {
        symbol: p.symbol,
        qty: p.qty,
        avg_price_cents: p.qty !== 0 ? Math.round(p.cost_basis_cents / p.qty) : 0,
        cost_basis_cents: p.cost_basis_cents,
        last_price_cents: last,
        market_value_cents: marketValue,
        unrealized_pnl_cents: marketValue - p.cost_basis_cents,
        realized_pnl_cents: p.realized_pnl_cents,
      };
    });

  return {
    holdings,
    total_market_value_cents: holdings.reduce((a, h) => a + (h.qty !== 0 ? h.market_value_cents : 0), 0),
    total_cost_basis_cents: holdings.reduce((a, h) => a + h.cost_basis_cents, 0),
    total_unrealized_pnl_cents: holdings.reduce((a, h) => a + (h.qty !== 0 ? h.unrealized_pnl_cents : 0), 0),
    total_realized_pnl_cents: holdings.reduce((a, h) => a + h.realized_pnl_cents, 0),
  };
}


function delay(ms = 120): Promise<void> {
  return new Promise((r) => setTimeout(r, ms));
}

function uid(prefix: string): string {
  return `${prefix}_${Math.random().toString(36).slice(2, 10)}`;
}

function nowIso(): string {
  return new Date().toISOString();
}

class MockError extends Error {
  constructor(message: string, public statusCode: number) {
    super(message);
    this.name = "MockError";
  }
}


function generateBars(symbol: string, count = 90): {
  open: number; high: number; low: number; close: number; time: string; symbol: string;
}[] {
  const bars: any[] = [];
  let price = basePrice(symbol) * 0.9;
  const start = Date.now() - count * 24 * 60 * 60 * 1000;
  for (let i = 0; i < count; i++) {
    const drift = (Math.random() - 0.48) * price * 0.02;
    const open = price;
    const close = Math.max(0.5, price + drift);
    const high = Math.max(open, close) + Math.random() * price * 0.01;
    const low = Math.min(open, close) - Math.random() * price * 0.01;
    bars.push({
      symbol,
      open: +open.toFixed(2),
      high: +high.toFixed(2),
      low: +Math.max(0.5, low).toFixed(2),
      close: +close.toFixed(2),
      time: new Date(start + i * 24 * 60 * 60 * 1000).toISOString().slice(0, 10),
    });
    price = close;
  }
  return bars;
}

function seededRand(seedStr: string): () => number {
  let h = 1779033703 ^ seedStr.length;
  for (let i = 0; i < seedStr.length; i++) {
    h = Math.imul(h ^ seedStr.charCodeAt(i), 3432918353);
    h = (h << 13) | (h >>> 19);
  }
  let a = h >>> 0;
  return function () {
    a = (a + 0x6d2b79f5) | 0;
    let t = Math.imul(a ^ (a >>> 15), 1 | a);
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

function backtestMock(symbol: string, timeframe: string, initialCents: number, posSizeCents: number, start?: string, end?: string) {
  const rand = seededRand(`${symbol}|${timeframe}|${start ?? ""}|${end ?? ""}`);
  const stepSec = timeframe === "MIN" ? 60 : timeframe === "HOUR" ? 3600 : 86400;
  const parseDateSec = (s?: string): number | null => {
    if (!s) return null;
    const ms = Date.parse(s.length === 10 ? `${s}T00:00:00Z` : s);
    return Number.isNaN(ms) ? null : Math.floor(ms / 1000);
  };
  const endSec = parseDateSec(end) ?? Math.floor(Date.now() / 1000);
  const startSec = parseDateSec(start);
  const defaultBars = 126;
  const N = startSec !== null ? Math.max(8, Math.min(2000, Math.round((endSec - startSec) / stepSec))) : defaultBars;
  const warmup = N >= 90 ? 70 : Math.max(2, Math.floor(N * 0.3));

  const equity_curve: { time_unix: number; equity_cents: number }[] = [];
  for (let i = 0; i < N; i++) {
    const x = i / (N - 1);
    const trend = 0.08 * x - 0.07 * Math.exp(-Math.pow((x - 0.45) / 0.1, 2));
    const noise = (rand() - 0.5) * 0.008;
    equity_curve.push({
      time_unix: endSec - (N - 1 - i) * stepSec,
      equity_cents: Math.round(initialCents * (1 + trend + noise)),
    });
  }
  equity_curve[N - 1].equity_cents = Math.round(initialCents * 1.08);

  let peak = initialCents;
  let maxDD = 0;
  for (const p of equity_curve) {
    if (p.equity_cents > peak) peak = p.equity_cents;
    if (peak > 0) maxDD = Math.max(maxDD, ((peak - p.equity_cents) / peak) * 100);
  }

  const ppy = timeframe === "MIN" ? 252 * 390 : timeframe === "HOUR" ? 252 * 6.5 : 252;
  const rets: number[] = [];
  for (let i = 1; i < equity_curve.length; i++) {
    const prev = equity_curve[i - 1].equity_cents;
    if (prev > 0) rets.push((equity_curve[i].equity_cents - prev) / prev);
  }
  const mean = rets.reduce((a, b) => a + b, 0) / rets.length;
  const variance = rets.reduce((a, b) => a + (b - mean) ** 2, 0) / (rets.length - 1);
  const std = Math.sqrt(variance);
  const sharpe = std > 0 ? (mean / std) * Math.sqrt(ppy) : 0;

  const base = basePriceCents(symbol);
  const retPattern = [3.2, 1.8, -2.1, 2.5, 4.0, 1.1, -1.5, 0.8];
  const reasons = ["take_profit", "signal", "stop_loss", "signal", "take_profit", "signal", "stop_loss", "end_of_data"];
  const trades = retPattern.map((rp, i) => {
    const entry = Math.round(base * (0.9 + rand() * 0.2));
    const qty = Math.max(1, Math.floor(posSizeCents / entry));
    const exit = Math.round(entry * (1 + rp / 100));
    const holdBars = 4 + Math.floor(rand() * 10);
    const entryIdx = Math.min(N - 1, warmup + Math.round(((N - 1 - warmup) * i) / retPattern.length));
    const exitIdx = Math.min(entryIdx + holdBars, N - 1);
    return {
      entry_time_unix: equity_curve[entryIdx].time_unix,
      exit_time_unix: equity_curve[exitIdx].time_unix,
      entry_price_cents: entry,
      exit_price_cents: exit,
      qty,
      pnl_cents: qty * (exit - entry),
      return_pct: rp,
      hold_bars: holdBars,
      exit_reason: reasons[i],
    };
  });

  const wins = trades.filter((t) => t.pnl_cents > 0).length;
  const grossProfit = trades.filter((t) => t.pnl_cents > 0).reduce((a, t) => a + t.pnl_cents, 0);
  const grossLoss = trades.filter((t) => t.pnl_cents < 0).reduce((a, t) => a - t.pnl_cents, 0);
  const avgHold = trades.reduce((a, t) => a + t.hold_bars, 0) / trades.length;
  const finalCents = equity_curve[N - 1].equity_cents;

  return {
    total_return_pct: ((finalCents - initialCents) / initialCents) * 100,
    max_drawdown_pct: maxDD,
    sharpe,
    win_rate: (wins / trades.length) * 100,
    profit_factor: grossLoss > 0 ? grossProfit / grossLoss : 0,
    num_trades: trades.length,
    avg_hold_bars: avgHold,
    final_equity_cents: finalCents,
    bars_used: N,
    warmup_bars: warmup,
    equity_curve,
    trades,
  };
}


export async function mockApi(
  path: string,
  options: RequestInit = {}
): Promise<any> {
  await delay();

  const method = (options.method || "GET").toUpperCase();
  const [pathname, queryString] = path.split("?");
  const query = new URLSearchParams(queryString || "");
  const body =
    options.body && typeof options.body === "string"
      ? safeJsonParse(options.body)
      : undefined;

  if (pathname === "auth/signin" && method === "POST") {
    return { success: true, user_id: "mock_user", email: body?.email ?? "mock@example.com" };
  }
  if (pathname === "auth/signup" && method === "POST") {
    return { success: true };
  }
  if (pathname === "auth/signout" && method === "POST") {
    return { success: true };
  }
  if (pathname === "auth/refresh" && method === "GET") {
    return { success: true };
  }
  if (pathname === "auth/forgot-password" && method === "POST") {
    return { success: true };
  }
  if (pathname === "auth/reset-password" && method === "POST") {
    return { success: true };
  }

  if (pathname === "account/me" && method === "GET") {
    return { id: "mock_user", user_id: "mock_user" };
  }
  if (pathname === "account" && method === "GET") {
    const equity = account.cash_balance_cents + holdingsPayload().total_market_value_cents;
    return {
      user_id: "mock_user",
      email: "mock@example.com",
      user_name: "mock_user",
      cash_balance_cents: account.cash_balance_cents,
      reserved_cash_cents: account.reserved_cash_cents,
      buying_power_cents: account.cash_balance_cents - account.reserved_cash_cents,
      equity_cents: equity,
      currency: "USD",
      multiplier: 1,
    };
  }
  if (pathname === "account/reset" && method === "POST") {
    account = { cash_balance_cents: 10_000_000, reserved_cash_cents: 0 };
    positions = [];
    orders = [];
    fills = [];
    persistAccount();
    persistPositions();
    persistOrders();
    persistFills();
    return { success: true };
  }
  if (pathname === "account" && method === "DELETE") {
    account = { ...DEFAULT_ACCOUNT };
    positions = [];
    orders = [];
    fills = [];
    strategies = [];
    watchlists = DEFAULT_WATCHLISTS.map((w) => ({ ...w }));
    persistAccount();
    persistPositions();
    persistOrders();
    persistFills();
    persistStrategies();
    persistWatchlists();
    return { success: true };
  }
  if (pathname === "account/profile" && method === "GET") {
    return { user_id: "mock_user", user_name: "mock_user", email: "mock@example.com" };
  }
  if (pathname === "account/trades" && method === "GET") {
    return { fills };
  }

  if (pathname === "portfolio" && method === "GET") {
    return {
      cash_balance_cents: account.cash_balance_cents,
      reserved_cash_cents: account.reserved_cash_cents,
      buying_power_cents: account.cash_balance_cents - account.reserved_cash_cents,
      margin_used_cents: 0,
      multiplier: 1,
      currency: "USD",
    };
  }
  if (pathname === "portfolio/holdings" && method === "GET") {
    return holdingsPayload();
  }
  if (pathname === "portfolio/positions" && method === "GET") {
    return {
      positions: positions.map((p) => ({
        symbol: p.symbol,
        qty: p.qty,
        reserved_qty: p.reserved_qty,
        avg_price_cents: p.qty !== 0 ? Math.round(p.cost_basis_cents / p.qty) : 0,
        cost_basis_cents: p.cost_basis_cents,
        realized_pnl_cents: p.realized_pnl_cents,
      })),
    };
  }

  if (pathname === "watchlists" && method === "GET") {
    return { w_metadata: watchlists.map((w) => ({ id: w.id, name: w.name })) };
  }
  if (pathname === "watchlists" && method === "POST") {
    const name = body?.name ?? "New Watchlist";
    const wl: Watchlist = { id: uid("wl"), name, symbols: [] };
    watchlists = [...watchlists, wl];
    persistWatchlists();
    return { id: wl.id, name: wl.name };
  }

  if (pathname === "watchlists/symbols" && method === "GET") {
    const q = (query.get("q") ?? "").toUpperCase();
    const limit = parseInt(query.get("limit") ?? "20", 10);
    const matches = SYMBOL_POOL.filter(
      (s) =>
        s.name.includes(q) ||
        s.company_name.toUpperCase().includes(q)
    ).slice(0, limit);
    return {
      symbols: matches.map((s) => ({ name: s.name, company_name: s.company_name })),
    };
  }

  if (pathname === "watchlists/history" && method === "POST") {
    const symbols: string[] = body?.symbols ?? [];
    return {
      symbol_bars: symbols.map((symbol) => ({
        symbol,
        bars: generateBars(symbol),
      })),
    };
  }

  const wlMatch = pathname.match(/^watchlists\/([^/]+)(?:\/(.+))?$/);
  if (wlMatch) {
    const id = wlMatch[1];
    const sub = wlMatch[2];
    const wl = watchlists.find((w) => w.id === id);

    if (sub === "info" && method === "GET") {
      if (!wl) throw new MockError("Watchlist not found", 404);
      return { id: wl.id, name: wl.name, symbols: wl.symbols };
    }

    if (!sub && method === "GET") {
      if (!wl) throw new MockError("Watchlist not found", 404);
      return { id: wl.id, name: wl.name, symbols: wl.symbols };
    }

    if (!sub && method === "PATCH") {
      if (!wl) throw new MockError("Watchlist not found", 404);
      const action = query.get("action");
      const syms: string[] = body?.symbols ?? [];
      if (action === "subscribe") {
        wl.symbols = Array.from(new Set([...wl.symbols, ...syms]));
      } else if (action === "unsubscribe") {
        wl.symbols = wl.symbols.filter((s) => !syms.includes(s));
      }
      persistWatchlists();
      return { id: wl.id, symbols: wl.symbols };
    }

    if (!sub && method === "DELETE") {
      const sym = query.get("symbol");
      if (sym) {
        if (!wl) throw new MockError("Watchlist not found", 404);
        wl.symbols = wl.symbols.filter((s) => s !== sym);
        persistWatchlists();
        return { id: wl.id, symbols: wl.symbols };
      }
      watchlists = watchlists.filter((w) => w.id !== id);
      persistWatchlists();
      return { success: true };
    }
  }

  if (pathname === "orders" && method === "GET") {
    const status = query.get("status") ?? "all";
    const filtered =
      status === "all" ? orders : orders.filter((o) => o.status === status);
    return { orders: filtered };
  }
  if (pathname === "orders" && method === "POST") {
    const price =
      typeof body?.price === "number" && body.price > 0
        ? body.price
        : basePriceCents(body?.symbol ?? "");
    const qty = Number(body?.qty ?? body?.quantity ?? 0);
    const order: OrderResponse = {
      id: uid("ord"),
      userId: "mock_user",
      symbol: body?.symbol ?? "",
      side: body?.side ?? "buy",
      orderType: body?.orderType ?? "market",
      timeInForce: body?.timeInForce ?? "day",
      qty,
      filledQty: 0,
      price,
      stopPrice: Number(body?.stopPrice ?? 0),
      status: "pending",
      createdAt: nowIso(),
      updatedAt: nowIso(),
    };
    orders = [order, ...orders];
    persistOrders();
    setTimeout(() => {
      const idx = orders.findIndex((o) => o.id === order.id);
      if (idx >= 0 && orders[idx].status === "pending") {
        orders[idx] = {
          ...orders[idx],
          filledQty: orders[idx].qty,
          status: "filled",
          updatedAt: nowIso(),
        };
        persistOrders();
        settleMockFill(orders[idx]);
      }
    }, 1500);
    return { order };
  }

  const orderMatch = pathname.match(/^orders\/([^/]+)$/);
  if (orderMatch) {
    const id = orderMatch[1];
    if (method === "GET") {
      const o = orders.find((x) => x.id === id);
      if (!o) throw new MockError("Order not found", 404);
      return o;
    }
    if (method === "DELETE") {
      const idx = orders.findIndex((x) => x.id === id);
      if (idx < 0) throw new MockError("Order not found", 404);
      orders[idx] = { ...orders[idx], status: "cancelled", updatedAt: nowIso() };
      persistOrders();
      return { success: true };
    }
  }

  if (pathname === "explore/news" && method === "GET") {
    const hoursAgo = (h: number) => new Date(Date.now() - h * 3600_000).toISOString();
    return {
      articles: [
        {
          id: "news-1",
          headline: "Tech Giants Rally on New AI Breakthroughs",
          summary: "Major technology stocks surged today following a string of announcements pointing towards highly optimized unified inference chips.",
          source: "MarketWire",
          url: "#",
          symbols: ["MSFT", "NVDA", "AAPL"],
          image_url: "https://images.unsplash.com/photo-1518770660439-4636190af475?q=80&w=2000&auto=format&fit=crop",
          created_at: hoursAgo(2),
        },
        {
          id: "news-2",
          headline: "Fed Holds Rates Steady, Signals Patience",
          summary: "The Federal Reserve left its benchmark rate unchanged, citing balanced risks between inflation progress and labor-market cooling.",
          source: "Capital Daily",
          url: "#",
          symbols: ["JPM", "BAC"],
          image_url: "https://images.unsplash.com/photo-1611974789855-9c2a0a7236a3?q=80&w=2000&auto=format&fit=crop",
          created_at: hoursAgo(5),
        },
        {
          id: "news-3",
          headline: "EV Makers Slide as Price War Intensifies",
          summary: "Electric-vehicle manufacturers fell after another round of aggressive price cuts squeezed margins across the sector.",
          source: "Street Signal",
          url: "#",
          symbols: ["TSLA", "RIVN", "LCID"],
          image_url: "",
          created_at: hoursAgo(8),
        },
        {
          id: "news-4",
          headline: "Cloud Spending Reaccelerates in Enterprise Survey",
          summary: "A quarterly CIO survey shows cloud budgets growing again, with security and data tooling leading the increases.",
          source: "MarketWire",
          url: "#",
          symbols: ["AMZN", "GOOGL", "CRM"],
          image_url: "https://images.unsplash.com/photo-1451187580459-43490279c0fa?q=80&w=2000&auto=format&fit=crop",
          created_at: hoursAgo(11),
        },
      ],
    };
  }
  if (pathname === "explore/movers" && method === "GET") {
    const mover = (symbol: string, name: string, cents: number, pct: number, vol: number) => ({
      symbol,
      company_name: name,
      price_cents: cents,
      change_percent: pct,
      volume: vol,
    });
    return {
      gainers: [
        mover("NVDA", "NVIDIA Corporation", 92_453, 5.4, 45_020_100),
        mover("PLTR", "Palantir Technologies Inc.", 2_345, 4.8, 32_014_023),
        mover("AMD", "Advanced Micro Devices", 16_320, 3.1, 18_220_400),
      ],
      losers: [
        mover("RIVN", "Rivian Automotive, Inc.", 1_020, -8.5, 11_040_300),
        mover("INTC", "Intel Corporation", 3_015, -4.2, 22_003_000),
        mover("F", "Ford Motor Company", 1_210, -2.4, 9_310_220),
      ],
    };
  }

  if (pathname === "strategies" && method === "GET") {
    return { strategies };
  }
  if (pathname === "strategies" && method === "POST") {
    const name = String(body?.name ?? "").trim();
    if (!name) throw new MockError("name is required", 400);
    if (strategies.some((st) => st.name === name)) {
      throw new MockError(`a strategy named "${name}" already exists`, 409);
    }
    const row: StoredStrategy = {
      id: uid("st"),
      name,
      config_json: typeof body?.config_json === "string"
        ? body.config_json
        : JSON.stringify(body?.config_json ?? {}),
      created_at: nowIso(),
      updated_at: nowIso(),
    };
    strategies = [row, ...strategies];
    persistStrategies();
    return row;
  }
  if (pathname === "strategies/backtest" && method === "POST") {
    return backtestMock(
      String(body?.symbol ?? "AAPL"),
      String(body?.timeframe ?? "DAY"),
      Number(body?.initial_capital_cents ?? 10_000_000),
      Number(body?.position_size_cents ?? 1_000_000),
      body?.start ? String(body.start) : undefined,
      body?.end ? String(body.end) : undefined
    );
  }
  const strategyMatch = pathname.match(/^strategies\/([^/]+)$/);
  if (strategyMatch) {
    const id = strategyMatch[1];
    if (method === "PATCH") {
      const idx = strategies.findIndex((st) => st.id === id);
      if (idx < 0) throw new MockError("Strategy not found", 404);
      strategies[idx] = {
        ...strategies[idx],
        name: String(body?.name ?? strategies[idx].name),
        config_json: typeof body?.config_json === "string"
          ? body.config_json
          : JSON.stringify(body?.config_json ?? {}),
        updated_at: nowIso(),
      };
      persistStrategies();
      return strategies[idx];
    }
    if (method === "DELETE") {
      strategies = strategies.filter((st) => st.id !== id);
      persistStrategies();
      return { success: true };
    }
  }

  if (typeof console !== "undefined") {
    console.warn(`[mock] unhandled ${method} ${path}`);
  }
  return null;
}

function safeJsonParse(input: string): any {
  try {
    return JSON.parse(input);
  } catch {
    return undefined;
  }
}


type Listener = (e: any) => void;

export class MockSocket {
  readyState = 1;
  onopen: Listener | null = null;
  onmessage: Listener | null = null;
  onclose: Listener | null = null;
  onerror: Listener | null = null;

  private listeners: Record<string, Listener[]> = {};
  private interval: ReturnType<typeof setInterval> | null = null;
  private prices: Record<string, number> = {};
  private watchlistId: string;

  constructor(watchlistId: string) {
    this.watchlistId = watchlistId;

    const wl = watchlists.find((w) => w.id === watchlistId);
    const syms = wl?.symbols ?? [];
    for (const s of syms) this.prices[s] = basePrice(s);

    setTimeout(() => {
      this.dispatch("open", {});
      this.tick();
      this.interval = setInterval(() => this.tick(), 1500);
    }, 0);
  }

  private tick() {
    const wl = watchlists.find((w) => w.id === this.watchlistId);
    const syms = wl?.symbols ?? Object.keys(this.prices);
    for (const s of syms) {
      if (this.prices[s] === undefined) this.prices[s] = basePrice(s);
      const p = this.prices[s];
      const drift = (Math.random() - 0.5) * p * 0.004;
      this.prices[s] = Math.max(0.01, +(p + drift).toFixed(2));
    }
    const payload = {
      symbol_bar: syms.map((symbol) => ({ symbol, close: this.prices[symbol] })),
    };
    this.dispatch("message", { data: JSON.stringify(payload) });
  }

  addEventListener(type: string, fn: Listener) {
    (this.listeners[type] ||= []).push(fn);
  }
  removeEventListener(type: string, fn: Listener) {
    this.listeners[type] = (this.listeners[type] || []).filter((l) => l !== fn);
  }

  private dispatch(type: string, event: any) {
    const handler = (this as any)["on" + type] as Listener | null;
    if (handler) handler(event);
    for (const l of this.listeners[type] || []) l(event);
  }

  send(_data: unknown) {
  }

  close() {
    if (this.interval) clearInterval(this.interval);
    this.interval = null;
    this.readyState = 3;
    this.dispatch("close", { code: 1000, reason: "mock close" });
  }
}
