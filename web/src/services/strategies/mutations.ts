import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";

export function useCreateStrategy() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ name, config }: { name: string; config: unknown }) =>
      api("strategies", {
        method: "POST",
        body: JSON.stringify({ name, config_json: config }),
      }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["strategies"] }),
  });
}

export function useUpdateStrategy() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, name, config }: { id: string; name: string; config: unknown }) =>
      api(`strategies/${id}`, {
        method: "PATCH",
        body: JSON.stringify({ name, config_json: config }),
      }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["strategies"] }),
  });
}

export function useDeleteStrategy() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api(`strategies/${id}`, { method: "DELETE" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["strategies"] }),
  });
}

export function useDeployStrategy() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      strategyId,
      symbol,
      positionSizeCents,
    }: {
      strategyId: string;
      symbol: string;
      positionSizeCents: number;
    }) =>
      api(`strategies/${strategyId}/deploy`, {
        method: "POST",
        body: JSON.stringify({
          symbol,
          position_size_cents: positionSizeCents,
        }),
      }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["deployments"] }),
  });
}

export function useStopDeployment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (deploymentId: string) =>
      api(`strategies/deployments/${deploymentId}/stop`, { method: "POST" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["deployments"] }),
  });
}

export function useDeleteDeployment() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (deploymentId: string) =>
      api(`strategies/deployments/${deploymentId}`, { method: "DELETE" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["deployments"] }),
  });
}


export type BacktestTimeframe = "DAY" | "HOUR" | "MIN";

export interface BacktestRequest {
    config: unknown;
  symbol: string;
  timeframe: BacktestTimeframe;
  positionSizeCents: number;
  start?: string;
  end?: string;
  initialCapitalCents?: number;
}

export interface BacktestTrade {
  entry_time_unix: number;
  exit_time_unix: number;
  entry_price_cents: number;
  exit_price_cents: number;
  qty: number;
  pnl_cents: number;
  return_pct: number;
  hold_bars: number;
  exit_reason: string;
}

export interface BacktestEquityPoint {
  time_unix: number;
  equity_cents: number;
}

export interface BacktestResult {
  total_return_pct: number;
  max_drawdown_pct: number;
  sharpe: number;
  win_rate: number;
  profit_factor: number;
  num_trades: number;
  avg_hold_bars: number;
  final_equity_cents: number;
  bars_used: number;
  warmup_bars: number;
  equity_curve: BacktestEquityPoint[];
  trades: BacktestTrade[];
}

export function useRunBacktest() {
  return useMutation({
    mutationFn: (req: BacktestRequest): Promise<BacktestResult> =>
      api("strategies/backtest", {
        method: "POST",
        body: JSON.stringify({
          config_json: req.config,
          symbol: req.symbol,
          timeframe: req.timeframe,
          position_size_cents: req.positionSizeCents,
          ...(req.start ? { start: req.start } : {}),
          ...(req.end ? { end: req.end } : {}),
          ...(req.initialCapitalCents
            ? { initial_capital_cents: req.initialCapitalCents }
            : {}),
        }),
      }),
  });
}
