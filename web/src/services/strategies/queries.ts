import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import queryBuilder from "@/lib/query";

export type ApiStrategy = {
  id: string;
  name: string;
  config_json: string;
  created_at: string;
  updated_at: string;
};

export const useStrategies = queryBuilder<ApiStrategy[]>(["strategies"], "strategies", {
  select: (data) => data?.strategies ?? [],
});

export type Deployment = {
  id: string;
  strategy_id: string;
  symbol: string;
  position_size_cents: number;
  status: "running" | "stopped";
  in_position: boolean;
  entry_price_cents: number;
  qty: number;
  strategy_name: string;
  created_at: string;
  updated_at: string;
};

export type StrategyFill = {
  trade_id: string;
  order_id: string;
  symbol: string;
  side: "buy" | "sell";
  qty: number;
  price_cents: number;
  executed_at: string;
};

export const useDeployments = queryBuilder<Deployment[]>(["deployments"], "strategies/deployments", {
  refetchInterval: 30_000,
  select: (data) => data?.deployments ?? [],
});

export function useStrategyTrades(strategyId: string | null) {
  return useQuery({
    queryKey: ["strategy-trades", strategyId],
    enabled: !!strategyId,
    refetchInterval: 30_000,
    queryFn: async (): Promise<StrategyFill[]> => {
      const data = await api(`strategies/${strategyId}/trades`);
      return data?.fills ?? [];
    },
  });
}
