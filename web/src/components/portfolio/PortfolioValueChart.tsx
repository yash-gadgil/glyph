"use client";

import { useEffect, useRef, useState } from "react";
import { AreaSeries, ColorType, createChart, type Time } from "lightweight-charts";
import {
  usePortfolioHistory,
  type PortfolioHistoryPoint,
} from "@/services/portfolio/queries";

const RANGES: { label: string; hours: number }[] = [
  { label: "1D", hours: 24 },
  { label: "1W", hours: 168 },
  { label: "1M", hours: 720 },
];

export default function PortfolioValueChart() {
  const [hours, setHours] = useState(24);
  const { data } = usePortfolioHistory(hours);

  const points: PortfolioHistoryPoint[] = data?.points ?? [];
  const series = (() => {
    const seen = new Map<number, number>();
    for (const p of points) seen.set(p.time_unix, p.equity_cents);
    return Array.from(seen.entries())
      .sort((a, b) => a[0] - b[0])
      .map(([t, equity]) => ({ time: t as Time, value: equity / 100 }));
  })();

  const up =
    series.length >= 2 ? series[series.length - 1].value >= series[0].value : true;

  return (
    <div className="rounded-2xl border border-white/10 bg-black/30 backdrop-blur-md p-5 space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-sm font-bold uppercase tracking-wider text-white/70">
          Account Value
        </h2>
        <div className="flex gap-1">
          {RANGES.map((r) => (
            <button
              key={r.label}
              onClick={() => setHours(r.hours)}
              className={`px-2.5 py-1 rounded-md text-[10px] font-bold uppercase tracking-wider border transition-colors ${
                hours === r.hours
                  ? "text-emerald-300 bg-emerald-500/10 border-emerald-500/30"
                  : "text-white/40 border-white/10 hover:bg-white/5"
              }`}
            >
              {r.label}
            </button>
          ))}
        </div>
      </div>

      {series.length < 2 ? (
        <div className="flex h-[220px] items-center justify-center text-xs text-white/40">
          Collecting account history…
        </div>
      ) : (
        <AreaChart series={series} up={up} />
      )}
    </div>
  );
}

function AreaChart({
  series,
  up,
}: {
  series: { time: Time; value: number }[];
  up: boolean;
}) {
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
      timeScale: { borderColor: "rgba(255,255,255,0.08)", timeVisible: true },
      handleScroll: false,
      handleScale: false,
      width: ref.current.clientWidth,
      height: 220,
    });

    const s = chart.addSeries(AreaSeries, {
      lineColor,
      topColor: up ? "rgba(52,211,153,0.35)" : "rgba(248,113,113,0.35)",
      bottomColor: "rgba(0,0,0,0)",
      lineWidth: 2,
    });
    s.setData(series);
    chart.timeScale().fitContent();

    const ro = new ResizeObserver(() => {
      if (ref.current) chart.applyOptions({ width: ref.current.clientWidth });
    });
    ro.observe(ref.current);

    return () => {
      ro.disconnect();
      chart.remove();
    };
  }, [series, up]);

  return <div ref={ref} className="w-full h-[220px]" />;
}
