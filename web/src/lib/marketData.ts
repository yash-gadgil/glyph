
import { useQuery, keepPreviousData } from "@tanstack/react-query";
import { api } from "./api";

export const TIMEFRAMES = ["1D", "1W", "1M", "3M", "1Y", "5Y"] as const;
export type Timeframe = (typeof TIMEFRAMES)[number];

export interface RealBar {
  time: string;
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
}

type ApiBar = {
  time: string;
  open?: number;
  high?: number;
  low?: number;
  close?: number;
  volume?: number;
};

type HistoryResponse = {
  symbol_bars?: { symbol: string; bars?: ApiBar[] }[];
};

const TIMEFRAME_REQUEST: Record<
  Timeframe,
  { timeframe: "DAY" | "HOUR" | "MIN"; days: number; intraday: boolean }
> = {
  "1D": { timeframe: "MIN", days: 5, intraday: true },
  "1W": { timeframe: "HOUR", days: 8, intraday: true },
  "1M": { timeframe: "HOUR", days: 31, intraday: true },
  "3M": { timeframe: "DAY", days: 93, intraday: false },
  "1Y": { timeframe: "DAY", days: 366, intraday: false },
  "5Y": { timeframe: "DAY", days: 366 * 5, intraday: false },
};

function toRealBars(data: HistoryResponse | null, symbol: string): RealBar[] {
  const entry = data?.symbol_bars?.find((sb) => sb.symbol === symbol);
  return (entry?.bars ?? []).map((b) => ({
    time: b.time,
    open: b.open ?? 0,
    high: b.high ?? 0,
    low: b.low ?? 0,
    close: b.close ?? 0,
    volume: b.volume ?? 0,
  }));
}

function sliceLastSession(bars: RealBar[]): RealBar[] {
  if (bars.length === 0) return bars;
  const lastDay = bars[bars.length - 1].time.slice(0, 10);
  return bars.filter((b) => b.time.slice(0, 10) === lastDay);
}

export function useSymbolBars(symbol: string, tf: Timeframe) {
  const req = TIMEFRAME_REQUEST[tf];
  return useQuery<RealBar[]>({
    queryKey: ["bars", symbol, tf],
    enabled: !!symbol,
    staleTime: 60 * 1000,
    placeholderData: keepPreviousData,
    queryFn: async () => {
      const data: HistoryResponse | null = await api("watchlists/history", {
        method: "POST",
        body: JSON.stringify({
          symbols: [symbol],
          timeframe: req.timeframe,
          days: req.days,
        }),
      });
      const bars = toRealBars(data, symbol);
      return tf === "1D" ? sliceLastSession(bars) : bars;
    },
  });
}

export function toChartBars(bars: RealBar[], tf: Timeframe) {
  const intraday = TIMEFRAME_REQUEST[tf].intraday;
  return bars.map((b) => ({
    time: intraday ? b.time : b.time.slice(0, 10),
    open: b.open,
    high: b.high,
    low: b.low,
    close: b.close,
  }));
}


export interface SymbolAnalytics {
  last: number;
  change: number;
  changePct: number;
  high: number;
  low: number;
  open: number;
  prevClose: number;
  volume: number;
  avgVolume: number;
  fiftyTwoHigh: number;
  fiftyTwoLow: number;
  range52Position: number;
  sma20: number;
  sma50: number;
  rsi: number;
  volatility: number;
}

function rsi(closes: number[], period = 14): number {
  if (closes.length < period + 1) return 50;
  let gains = 0;
  let losses = 0;
  for (let i = closes.length - period; i < closes.length; i++) {
    const diff = closes[i] - closes[i - 1];
    if (diff >= 0) gains += diff;
    else losses -= diff;
  }
  if (losses === 0) return 100;
  const rs = gains / losses;
  return +(100 - 100 / (1 + rs)).toFixed(2);
}

function sma(closes: number[], period: number): number {
  const slice = closes.slice(-period);
  if (slice.length === 0) return 0;
  return +(slice.reduce((a, b) => a + b, 0) / slice.length).toFixed(2);
}

function stdev(values: number[]): number {
  if (values.length < 2) return 0;
  const mean = values.reduce((a, b) => a + b, 0) / values.length;
  const variance =
    values.reduce((a, b) => a + (b - mean) ** 2, 0) / (values.length - 1);
  return Math.sqrt(variance);
}

export function computeAnalytics(
  dailyBars: RealBar[],
  livePrice?: number
): SymbolAnalytics {
  const closes = dailyBars.map((b) => b.close);
  const lastBar = dailyBars[dailyBars.length - 1];
  const last = livePrice ?? lastBar?.close ?? 0;
  const prevClose = closes[closes.length - 2] ?? last;
  const change = last - prevClose;
  const changePct = prevClose ? (change / prevClose) * 100 : 0;

  const volumes = dailyBars.map((b) => b.volume);
  const fiftyTwoHigh = Math.max(...dailyBars.map((b) => b.high), last);
  const fiftyTwoLow = Math.min(...dailyBars.map((b) => b.low), last);

  const dailyReturns: number[] = [];
  for (let i = 1; i < closes.length; i++) {
    if (closes[i - 1] > 0) dailyReturns.push((closes[i] - closes[i - 1]) / closes[i - 1]);
  }

  return {
    last,
    change,
    changePct,
    high: lastBar?.high ?? last,
    low: lastBar?.low ?? last,
    open: lastBar?.open ?? last,
    prevClose,
    volume: lastBar?.volume ?? 0,
    avgVolume: volumes.length
      ? Math.round(volumes.reduce((a, b) => a + b, 0) / volumes.length)
      : 0,
    fiftyTwoHigh,
    fiftyTwoLow,
    range52Position:
      fiftyTwoHigh === fiftyTwoLow
        ? 0.5
        : (last - fiftyTwoLow) / (fiftyTwoHigh - fiftyTwoLow),
    sma20: sma(closes, 20),
    sma50: sma(closes, 50),
    rsi: rsi(closes),
    volatility: +(stdev(dailyReturns) * Math.sqrt(252) * 100).toFixed(2),
  };
}

export function windowReturnPct(dailyBars: RealBar[], days: number): number {

  if (dailyBars.length < 2) return 0;

  const closes = dailyBars.map((b) => b.close);
  const last = closes[closes.length - 1];
  const startIdx = Math.max(0, closes.length - 1 - days);
  const first = closes[startIdx];
  
  return first ? ((last - first) / first) * 100 : 0;
}

export function formatVolume(value: number): string {

  if (value >= 1e9) return `${(value / 1e9).toFixed(2)}B`;

  if (value >= 1e6) return `${(value / 1e6).toFixed(2)}M`;

  if (value >= 1e3) return `${(value / 1e3).toFixed(1)}K`;

  return `${value}`;
}
