import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import queryBuilder from "@/lib/query";

export const getPortfolio = queryBuilder(["portfolio"], "portfolio");

export interface PortfolioHistoryPoint {
  time_unix: number;
  equity_cents: number;
  cash_cents: number;
  market_value_cents: number;
}

export function usePortfolioHistory(hours: number) {
  return useQuery({
    queryKey: ["portfolio", "history", hours],
    queryFn: () => api(`portfolio/history?hours=${hours}`),
    refetchInterval: 60_000,
    staleTime: 60_000,
  });
}

export const getHoldings = queryBuilder(
  ["portfolio", "holdings"],
  "portfolio/holdings",
);

export const getPositions = queryBuilder(
  ["portfolio", "positions"],
  "portfolio/positions",
);
