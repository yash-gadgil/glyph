import { API_BASE_URL } from "@/lib/api";
import { isMockMode } from "@/lib/mock";
import type { CustomStrategy } from "@/lib/strategies";

export type GeneratedStrategy = {
  name: string;
  rationale: string;
  template: string;
  config: CustomStrategy;
};

const MOCK_STRATEGY: GeneratedStrategy = {
  name: "RSI Dip Buyer demo",
  rationale: "Your book leans on a couple of names, so a mean reversion entry lets you add on pullbacks without chasing.",
  template: "rsi_dip_buyer",
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
};

export async function generateStrategy(): Promise<GeneratedStrategy> {
  if (isMockMode()) {
    await new Promise((r) => setTimeout(r, 500));
    return MOCK_STRATEGY;
  }

  const res = await fetch(`${API_BASE_URL}/advisor/strategy`, {
    method: "POST",
    credentials: "include",
  });
  if (!res.ok) {
    throw new Error("Unable to generate a strategy");
  }
  return res.json();
}
