import { API_BASE_URL } from "@/lib/api";
import { isMockMode } from "@/lib/mock";
import type { CustomStrategy, RuleGroup } from "@/lib/strategies";

export type BacktestSummary = {
  total_return_pct: number;
  max_drawdown_pct: number;
  sharpe: number;
  win_rate: number;
  profit_factor: number;
  num_trades: number;
};

export type StrategyJobState = "" | "running" | "succeeded" | "failed";

export type StrategyJob = {
  state: StrategyJobState;
  name: string;
  rationale: string;
  error: string;
  config?: CustomStrategy;
  backtest?: BacktestSummary;
};

export type GeneratedStrategy = {
  name: string;
  rationale: string;
  config: CustomStrategy;
  backtest?: BacktestSummary;
};

const POLL_INTERVAL_MS = 1500;
const POLL_TIMEOUT_MS = 120000;

const MOCK_STRATEGY: GeneratedStrategy = {
  name: "RSI Dip Buyer demo",
  rationale: "Your book leans on a couple of names, so a mean reversion entry lets you add on pullbacks without chasing.",
  config: {
    id: "",
    name: "RSI Dip Buyer demo",
    description: "",
    risk: "medium",
    tags: ["Mean Reversion", "RSI"],
    entry: { combinator: "AND", rules: [{ id: "e1", lhs: { kind: "rsi", params: { period: 14 } }, op: "<", rhs: { kind: "value", value: 30 } }] },
    exit: { combinator: "OR", rules: [{ id: "x1", lhs: { kind: "rsi", params: { period: 14 } }, op: ">", rhs: { kind: "value", value: 60 } }] },
    stopLossPct: 2,
    takeProfitPct: 3,
    createdAt: new Date().toISOString(),
  },
  backtest: {
    total_return_pct: 7.4,
    max_drawdown_pct: 3.1,
    sharpe: 1.2,
    win_rate: 0.58,
    profit_factor: 1.6,
    num_trades: 12,
  },
};

function ensureRuleIds(group: RuleGroup): RuleGroup {
  return {
    combinator: group.combinator,
    rules: (group.rules ?? []).map((r) => ({ ...r, id: r.id || Math.random().toString(36).slice(2, 10) })),
  };
}

function toGenerated(job: StrategyJob): GeneratedStrategy {
  if (!job.config) {
    throw new Error("Strategy job completed without a config");
  }
  return {
    name: job.name,
    rationale: job.rationale,
    backtest: job.backtest,
    config: {
      ...job.config,
      entry: ensureRuleIds(job.config.entry),
      exit: ensureRuleIds(job.config.exit),
    },
  };
}

async function startJob(symbol: string): Promise<StrategyJob> {
  const res = await fetch(`${API_BASE_URL}/advisor/strategy`, {
    method: "POST",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ symbol }),
  });
  if (!res.ok) {
    throw new Error("Unable to start strategy generation");
  }
  return res.json();
}

export async function getStrategyJob(): Promise<StrategyJob> {
  const res = await fetch(`${API_BASE_URL}/advisor/strategy/status`, {
    credentials: "include",
  });
  if (!res.ok) {
    throw new Error("Unable to read strategy job");
  }
  return res.json();
}

function terminal(job: StrategyJob): GeneratedStrategy | null {
  if (job.state === "succeeded") {
    return toGenerated(job);
  }
  if (job.state === "failed") {
    throw new Error(job.error || "Strategy generation failed");
  }
  return null;
}

async function pollUntilDone(): Promise<GeneratedStrategy> {
  const deadline = Date.now() + POLL_TIMEOUT_MS;
  while (Date.now() < deadline) {
    await new Promise((r) => setTimeout(r, POLL_INTERVAL_MS));
    const done = terminal(await getStrategyJob());
    if (done) {
      return done;
    }
  }
  throw new Error("Strategy generation timed out");
}

export async function generateStrategy(symbol: string): Promise<GeneratedStrategy> {
  if (isMockMode()) {
    await new Promise((r) => setTimeout(r, 600));
    return { ...MOCK_STRATEGY, name: `${symbol} ${MOCK_STRATEGY.name}` };
  }

  const done = terminal(await startJob(symbol));
  if (done) {
    return done;
  }
  return pollUntilDone();
}

export async function isStrategyGenerationRunning(): Promise<boolean> {
  if (isMockMode()) {
    return false;
  }
  try {
    const job = await getStrategyJob();
    return job.state === "running";
  } catch {
    return false;
  }
}

export async function resumeStrategyGeneration(): Promise<GeneratedStrategy> {
  if (isMockMode()) {
    return MOCK_STRATEGY;
  }
  return pollUntilDone();
}
