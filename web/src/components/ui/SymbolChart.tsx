'use client';

import { CandlestickSeries, ColorType, createChart, IChartApi, ISeriesApi, Time } from "lightweight-charts";
import { useEffect, useRef } from "react";

interface SymbolChartProps {
  symbol: string;
  prices: {
    time: string;
    open: number;
    high: number;
    close: number;
    low: number;
  }[];
}

function toLwcTime(time: string): Time {
  if (time.includes("T")) {
    return Math.floor(new Date(time).getTime() / 1000) as Time;
  }
  if (time.includes(" ")) {
    return Math.floor(new Date(time.replace(" ", "T") + "Z").getTime() / 1000) as Time;
  }
  return time as Time;
}

export default function SymbolChart({ symbol, prices }: SymbolChartProps) {
  const chartContainerRef = useRef<HTMLDivElement>(null);
  const tooltipRef = useRef<HTMLDivElement>(null);
  const chartInstanceRef = useRef<IChartApi | null>(null);
  const seriesRef = useRef<ISeriesApi<"Candlestick"> | null>(null);

  useEffect(() => {
    if (!chartContainerRef.current) return;

    const chart = createChart(chartContainerRef.current, {
      layout: {
        textColor: "rgba(255, 255, 255, 0.7)",
        background: { type: ColorType.Solid, color: "transparent" },
        fontFamily: "var(--font-mono, ui-monospace)",
      },
      grid: {
        vertLines: { color: "rgba(255, 255, 255, 0.04)" },
        horzLines: { color: "rgba(255, 255, 255, 0.04)" },
      },
      rightPriceScale: { borderColor: "rgba(255,255,255,0.08)" },
      timeScale: {
        borderColor: "rgba(255,255,255,0.08)",
        lockVisibleTimeRangeOnResize: true,
      },
      crosshair: {
        vertLine: { color: "rgba(255,255,255,0.25)", labelBackgroundColor: "#0a0a0a" },
        horzLine: { color: "rgba(255,255,255,0.25)", labelBackgroundColor: "#0a0a0a" },
      },
      handleScroll: {
        mouseWheel: false,
        pressedMouseMove: false,
        horzTouchDrag: false,
        vertTouchDrag: false,
      },
      handleScale: {
        axisPressedMouseMove: false,
        axisDoubleClickReset: false,
        mouseWheel: false,
        pinch: false,
      },
      kineticScroll: { touch: false, mouse: false },
      width: chartContainerRef.current.clientWidth,
      height: chartContainerRef.current.clientHeight,
    });
    chartInstanceRef.current = chart;

    const candlestickSeries = chart.addSeries(CandlestickSeries, {
      upColor: "#34d399",
      downColor: "#f87171",
      borderVisible: false,
      wickUpColor: "#34d399",
      wickDownColor: "#f87171",
    });
    seriesRef.current = candlestickSeries;

    candlestickSeries.setData(prices.map((p) => ({ ...p, time: toLwcTime(p.time) })));
    chart.timeScale().fitContent();

    const tooltip = tooltipRef.current;

    chart.subscribeCrosshairMove((param) => {
      if (!tooltip || !chartContainerRef.current) return;

      if (
        param.point === undefined ||
        !param.time ||
        param.point.x < 0 ||
        param.point.x > chartContainerRef.current.clientWidth ||
        param.point.y < 0 ||
        param.point.y > chartContainerRef.current.clientHeight
      ) {
        tooltip.style.display = "none";
        return;
      }

      const data = param.seriesData.get(candlestickSeries) as
        | { open: number; high: number; low: number; close: number }
        | undefined;
      if (data && "open" in data) {
        const change = data.close - data.open;
        const pct = data.open ? (change / data.open) * 100 : 0;
        const sign = change >= 0 ? "+" : "";
        const colorClass = change >= 0 ? "text-emerald-400" : "text-rose-400";
        tooltip.style.display = "block";
        tooltip.innerHTML = `
          <div class="font-bold mb-2 text-base flex items-center justify-between gap-3">
            <span>${symbol}</span>
            <span class="${colorClass} text-xs">${sign}${change.toFixed(2)} (${sign}${pct.toFixed(2)}%)</span>
          </div>
          <div class="grid grid-cols-[auto_auto] gap-x-3 gap-y-1 text-xs">
            <span class="text-neutral-400">O</span> <span>${data.open.toFixed(2)}</span>
            <span class="text-neutral-400">H</span> <span class="text-emerald-400">${data.high.toFixed(2)}</span>
            <span class="text-neutral-400">L</span> <span class="text-rose-400">${data.low.toFixed(2)}</span>
            <span class="text-neutral-400">C</span> <span>${data.close.toFixed(2)}</span>
          </div>
        `;

        const margin = 15;
        let left = param.point.x + margin;
        let top = param.point.y + margin;
        const tooltipRect = tooltip.getBoundingClientRect();
        if (left + (tooltipRect.width || 160) > chartContainerRef.current.clientWidth) {
          left = param.point.x - (tooltipRect.width || 160) - margin;
        }
        if (top + (tooltipRect.height || 120) > chartContainerRef.current.clientHeight) {
          top = param.point.y - (tooltipRect.height || 120) - margin;
        }
        tooltip.style.left = left + "px";
        tooltip.style.top = top + "px";
      } else {
        tooltip.style.display = "none";
      }
    });

    const handleResize = () => {
      if (chartContainerRef.current) {
        chart.applyOptions({
          width: chartContainerRef.current.clientWidth,
          height: chartContainerRef.current.clientHeight,
        });
        chart.timeScale().fitContent();
      }
    };
    const resizeObserver = new ResizeObserver(handleResize);
    resizeObserver.observe(chartContainerRef.current);

    return () => {
      resizeObserver.disconnect();
      chart.remove();
      chartInstanceRef.current = null;
      seriesRef.current = null;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [symbol]);

  useEffect(() => {
    if (seriesRef.current && chartInstanceRef.current) {
      seriesRef.current.setData(prices.map((p) => ({ ...p, time: toLwcTime(p.time) })));
      chartInstanceRef.current.timeScale().fitContent();
    }
  }, [prices]);

  return (
    <div className="relative w-full h-full min-h-[400px] flex flex-col">
      <div
        ref={tooltipRef}
        className="absolute z-50 text-white font-mono pointer-events-none drop-shadow-md bg-neutral-900/95 border border-white/10 px-3 py-2 rounded-lg shadow-lg"
        style={{ display: "none" }}
      />
      <div ref={chartContainerRef} className="flex-1 w-full" />
    </div>
  );
}
