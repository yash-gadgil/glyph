'use client';

import { useEffect, useRef, useState } from "react";
import { AlertTriangle, Loader2 } from "lucide-react";
import { AreaSeries, ColorType, createChart, type Time } from "lightweight-charts";
import PixelHover from "@/components/ui/PixelHover";
import ModalPortal from "@/components/ui/ModalPortal";
import SymbolCombobox from "@/components/ui/SymbolCombobox";
import { formatCents } from "@/lib/utils";
import {
  useRunBacktest,
  type BacktestEquityPoint,
  type BacktestResult,
  type BacktestTimeframe,
} from "@/services/strategies/mutations";

const TIMEFRAMES: { value: BacktestTimeframe; label: string; window: string }[] = [
  { value: "DAY", label: "Daily", window: "6M" },
  { value: "HOUR", label: "Hourly", window: "30D" },
  { value: "MIN", label: "Minute", window: "7D" },
];

export default function BacktestModal({
  name,
  tagline,
  config,
  onClose,
}: {
  name: string;
  tagline: string;
  config: unknown;
  onClose: () => void;
}) {
  const [symbol, setSymbol] = useState("AAPL");
  const [timeframe, setTimeframe] = useState<BacktestTimeframe>("DAY");
  const [posSize, setPosSize] = useState("10000");
  const [startDate, setStartDate] = useState("");
  const [endDate, setEndDate] = useState("");
  const [error, setError] = useState("");

  const backtest = useRunBacktest();
  const result = backtest.data;
  const apiError = backtest.error instanceof Error ? backtest.error.message : "";
  const shownError = error || apiError;
  const windowLabel = TIMEFRAMES.find((t) => t.value === timeframe)?.window ?? "";

  const today = new Date().toISOString().slice(0, 10);
  const hasCustomRange = Boolean(startDate || endDate);
  const rangeDescription = hasCustomRange
    ? `from ${startDate || "the default window start"} to ${endDate || "today"}`
    : `over the last ${windowLabel}`;

  function run() {
    if (!symbol.trim()) {
      setError("Symbol is required");
      return;
    }
    const size = parseFloat(posSize);
    if (!size || size <= 0) {
      setError("Position size must be a positive number");
      return;
    }
    if (startDate && endDate && startDate >= endDate) {
      setError("Start date must be before end date");
      return;
    }
    setError("");
    backtest.mutate({
      config,
      symbol: symbol.toUpperCase().trim(),
      timeframe,
      positionSizeCents: Math.round(size * 100),
      start: startDate || undefined,
      end: endDate || undefined,
    });
  }

  return (
    <ModalPortal>
      <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
        <div className="w-full max-w-3xl rounded-2xl border border-white/10 bg-neutral-950/95 shadow-2xl overflow-hidden">
          <div className="px-6 py-5 border-b border-white/10 flex items-start justify-between gap-4">
            <div>
              <p className="text-xs text-white/40 uppercase tracking-widest mb-1 font-semibold">Backtest Strategy</p>
              <h2 className="text-xl font-bold text-white">{name}</h2>
              <p className="text-xs text-white/50 mt-1">{tagline}</p>
            </div>
            <button
              onClick={onClose}
              className="p-1.5 rounded-lg text-white/40 hover:text-white hover:bg-white/10 transition-colors mt-0.5 shrink-0"
            >
              ✕
            </button>
          </div>

          <div className="px-6 py-5 space-y-5 max-h-[75vh] overflow-y-auto">
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
              <div className="space-y-1.5">
                <label className="text-xs font-semibold text-white/50 uppercase tracking-wider">Symbol</label>
                <SymbolCombobox
                  value={symbol}
                  onChange={(v) => setSymbol(v.toUpperCase())}
                  placeholder="AAPL"
                  inputClassName="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-sm text-white placeholder:text-white/30 focus:outline-none focus:border-emerald-500/50 focus:ring-1 focus:ring-emerald-500/30 transition-all"
                />
              </div>
              <div className="space-y-1.5">
                <label className="text-xs font-semibold text-white/50 uppercase tracking-wider">Timeframe</label>
                <select
                  value={timeframe}
                  onChange={(e) => setTimeframe(e.target.value as BacktestTimeframe)}
                  className="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-sm text-white focus:outline-none focus:border-emerald-500/50 focus:ring-1 focus:ring-emerald-500/30 transition-all"
                >
                  {TIMEFRAMES.map((t) => (
                    <option key={t.value} value={t.value} className="bg-neutral-900">
                      {t.label} · {t.window}
                    </option>
                  ))}
                </select>
              </div>
              <div className="space-y-1.5">
                <label className="text-xs font-semibold text-white/50 uppercase tracking-wider">Position Size ($)</label>
                <input
                  type="number"
                  value={posSize}
                  onChange={(e) => setPosSize(e.target.value)}
                  placeholder="10000"
                  className="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-sm text-white placeholder:text-white/30 focus:outline-none focus:border-emerald-500/50 focus:ring-1 focus:ring-emerald-500/30 transition-all"
                />
              </div>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div className="space-y-1.5">
                <label className="text-xs font-semibold text-white/50 uppercase tracking-wider">
                  Start Date{" "}
                  <span className="normal-case font-normal text-white/30">(optional)</span>
                </label>
                <input
                  type="date"
                  value={startDate}
                  max={endDate || today}
                  onChange={(e) => setStartDate(e.target.value)}
                  className="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-sm text-white scheme:dark focus:outline-none focus:border-emerald-500/50 focus:ring-1 focus:ring-emerald-500/30 transition-all"
                />
              </div>
              <div className="space-y-1.5">
                <label className="text-xs font-semibold text-white/50 uppercase tracking-wider">
                  End Date{" "}
                  <span className="normal-case font-normal text-white/30">(optional)</span>
                </label>
                <input
                  type="date"
                  value={endDate}
                  min={startDate || undefined}
                  max={today}
                  onChange={(e) => setEndDate(e.target.value)}
                  className="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-sm text-white scheme:dark focus:outline-none focus:border-emerald-500/50 focus:ring-1 focus:ring-emerald-500/30 transition-all"
                />
              </div>
            </div>

            <p className="text-[10px] text-white/30 leading-relaxed">
              Replays the strategy {rangeDescription} of {timeframe.toLowerCase()} bars against historical
              prices. Lookahead-free, signals fill at the next bar&apos;s open. $100,000 starting capital, no
              commission. Sharpe is annualized for the timeframe.
            </p>

            {shownError && (
              <div className="p-3 bg-red-500/10 border border-red-500/20 rounded-xl text-red-400 text-xs font-medium flex items-center gap-2">
                <AlertTriangle size={13} /> {shownError}
              </div>
            )}

            {backtest.isPending && (
              <div className="flex items-center justify-center gap-2 py-10 text-white/40 text-sm">
                <Loader2 size={16} className="animate-spin" /> Replaying bars…
              </div>
            )}

            {result && !backtest.isPending && <BacktestResults result={result} />}
          </div>

          <div className="px-6 py-4 border-t border-white/10 flex items-center gap-3">
            <button
              onClick={onClose}
              className="flex-1 py-3 rounded-xl text-sm font-semibold uppercase tracking-wider text-white/50 border border-white/10 hover:bg-white/5 transition-all"
            >
              Close
            </button>
            <PixelHover
              variant="emerald"
              className="group flex-1 rounded-xl border border-white/10 bg-white/5 hover:bg-white/10 hover:border-emerald-500/40 transition-all active:scale-[0.98]"
            >
              <button
                onClick={run}
                disabled={backtest.isPending}
                className="w-full py-3 rounded-xl text-sm font-bold uppercase tracking-wider text-white bg-transparent group-hover:text-emerald-300 transition-colors disabled:opacity-50 flex items-center justify-center gap-2"
              >
                {backtest.isPending ? (
                  <>
                    <Loader2 size={14} className="animate-spin" /> Running…
                  </>
                ) : result ? (
                  "Run Again"
                ) : (
                  "Run Backtest"
                )}
              </button>
            </PixelHover>
          </div>
        </div>
      </div>
    </ModalPortal>
  );
}

function StatCard({ label, value, tone }: { label: string; value: string; tone?: string }) {
  return (
    <div className="rounded-xl border border-white/10 bg-white/5 p-3">
      <p className="text-[9px] uppercase tracking-widest text-white/40 mb-1">{label}</p>
      <p className={`text-lg font-bold tabular-nums ${tone ?? "text-white"}`}>{value}</p>
    </div>
  );
}

function reasonBadge(reason: string): string {
  switch (reason) {
    case "take_profit":
      return "text-emerald-400 bg-emerald-500/10 border-emerald-500/20";
    case "stop_loss":
      return "text-red-400 bg-red-500/10 border-red-500/20";
    case "end_of_data":
      return "text-amber-300 bg-amber-500/10 border-amber-500/20";
    default:
      return "text-white/60 bg-white/5 border-white/10";
  }
}

function BacktestResults({ result }: { result: BacktestResult }) {
  const ret = result.total_return_pct;
  const pf = result.profit_factor;

  return (
    <div className="space-y-5" data-testid="backtest-results">
      <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
        <StatCard
          label="Total Return"
          value={`${ret >= 0 ? "+" : ""}${ret.toFixed(2)}%`}
          tone={ret >= 0 ? "text-emerald-400" : "text-red-400"}
        />
        <StatCard label="Max Drawdown" value={`-${result.max_drawdown_pct.toFixed(2)}%`} tone="text-red-400" />
        <StatCard label="Sharpe" value={result.sharpe.toFixed(2)} />
        <StatCard label="Win Rate" value={`${result.win_rate.toFixed(0)}%`} />
        <StatCard label="Profit Factor" value={pf > 0 ? pf.toFixed(2) : "-"} />
        <StatCard label="Trades" value={String(result.num_trades)} />
      </div>

      {result.equity_curve.length > 1 && (
        <div className="rounded-xl border border-white/10 bg-white/5 p-3">
          <p className="text-[9px] uppercase tracking-widest text-white/40 mb-2">Equity Curve</p>
          <EquityCurveChart points={result.equity_curve} up={ret >= 0} />
        </div>
      )}

      {result.trades.length > 0 && (
        <div className="rounded-xl border border-white/10 overflow-hidden">
          <table className="w-full text-xs">
            <thead className="bg-white/5 text-white/40 uppercase tracking-wider text-[9px]">
              <tr>
                <th className="text-left px-3 py-2 font-semibold">Entry</th>
                <th className="text-left px-3 py-2 font-semibold">Exit</th>
                <th className="text-right px-3 py-2 font-semibold">Qty</th>
                <th className="text-right px-3 py-2 font-semibold">P&amp;L</th>
                <th className="text-right px-3 py-2 font-semibold">Reason</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/5">
              {result.trades.map((t, i) => {
                const win = t.pnl_cents >= 0;
                return (
                  <tr key={i} className="text-white/70">
                    <td className="px-3 py-2 tabular-nums">${formatCents(t.entry_price_cents)}</td>
                    <td className="px-3 py-2 tabular-nums">${formatCents(t.exit_price_cents)}</td>
                    <td className="px-3 py-2 text-right tabular-nums">{t.qty}</td>
                    <td className={`px-3 py-2 text-right tabular-nums font-semibold ${win ? "text-emerald-400" : "text-red-400"}`}>
                      {win ? "+" : "-"}${formatCents(Math.abs(t.pnl_cents))}
                    </td>
                    <td className="px-3 py-2 text-right">
                      <span className={`px-1.5 py-0.5 rounded text-[9px] font-bold uppercase tracking-wider border ${reasonBadge(t.exit_reason)}`}>
                        {t.exit_reason.replace(/_/g, " ")}
                      </span>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}

function EquityCurveChart({ points, up }: { points: BacktestEquityPoint[]; up: boolean }) {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!ref.current) return;
    const lineColor = up ? "#34d399" : "#f87171";

    const chart = createChart(ref.current, {
      layout: {
        textColor: "rgba(255,255,255,0.6)",
        background: { type: ColorType.Solid, color: "transparent" },
        fontFamily: "var(--font-mono, ui-monospace)",
      },
      grid: {
        vertLines: { color: "rgba(255,255,255,0.04)" },
        horzLines: { color: "rgba(255,255,255,0.04)" },
      },
      rightPriceScale: { borderColor: "rgba(255,255,255,0.08)" },
      timeScale: { borderColor: "rgba(255,255,255,0.08)" },
      handleScroll: false,
      handleScale: false,
      width: ref.current.clientWidth,
      height: 220,
    });

    const series = chart.addSeries(AreaSeries, {
      lineColor,
      topColor: up ? "rgba(52,211,153,0.35)" : "rgba(248,113,113,0.35)",
      bottomColor: "rgba(0,0,0,0)",
      lineWidth: 2,
    });
    series.setData(points.map((p) => ({ time: p.time_unix as Time, value: p.equity_cents / 100 })));
    chart.timeScale().fitContent();

    const ro = new ResizeObserver(() => {
      if (ref.current) chart.applyOptions({ width: ref.current.clientWidth });
    });
    ro.observe(ref.current);

    return () => {
      ro.disconnect();
      chart.remove();
    };
  }, [points, up]);

  return <div ref={ref} className="w-full h-[220px]" />;
}
