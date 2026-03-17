"use client";

import { Sparklines, SparklinesLine } from "react-sparklines-ts";
import GlassSurface from "../primitives/GlassSurface";
import { useMemo, useState } from "react";
import { X } from "lucide-react";
import Link from "next/link";
import { Skeleton } from "../primitives/Skeleton";
import TimeframeSelector from "./TimeframeSelector";
import { Timeframe, useSymbolBars, formatVolume } from "@/lib/marketData";
import { SlidingNumber } from "../primitives/SlidingNumber";
import { motion, AnimatePresence } from "motion/react";

const MOCK_COMPANIES: Record<string, string> = {
  AAPL: "Apple Inc.",
  MSFT: "Microsoft Corp.",
  GOOGL: "Alphabet Inc.",
  AMZN: "Amazon.com",
  TSLA: "Tesla Inc.",
  NVDA: "NVIDIA Corp.",
  META: "Meta Platforms",
  BTC: "Bitcoin",
  ETH: "Ethereum",
  POOL: "Pool Corporation",
  LOKV: "Lockheed Martin",
  JOJO: "Jojo Holdings",
  DUNK: "Dunkin' Brands",
};

function getMockCompany(symbol: string) {
  return MOCK_COMPANIES[symbol.toUpperCase()] || "Equities";
}

const CARD_TIMEFRAMES = ["1D", "1W", "1M", "1Y"] as const satisfies readonly Timeframe[];

interface SymbolCardProps {
  symbol: string;
  livePrice?: number;
  onRemove?: (symbol: string) => void;
}

export default function SymbolCard({ symbol, livePrice, onRemove }: SymbolCardProps) {
  const [timeframe, setTimeframe] = useState<Timeframe>("1M");

  const { data: bars = [], isLoading } = useSymbolBars(symbol, timeframe);

  const { data: anchorBars = [] } = useSymbolBars(symbol, "1D");

  const closes = useMemo(() => bars.map((b) => b.close), [bars]);

  const chartData = useMemo(() => {
    if (closes.length === 0) return closes;
    return livePrice != null ? [...closes, livePrice] : closes;
  }, [closes, livePrice]);

  const anchorPrice = anchorBars[anchorBars.length - 1]?.close;
  const latestPrice = livePrice ?? anchorPrice ?? closes[closes.length - 1];
  const firstPrice = closes[0];

  const change = latestPrice != null && firstPrice != null ? latestPrice - firstPrice : 0;
  const percentChange = firstPrice != null && firstPrice !== 0 ? (change / firstPrice) * 100 : 0;
  const isPositive = change >= 0;

  const colorClass = isPositive ? "text-emerald-400" : "text-rose-400";
  const lineColor = isPositive ? "#34d399" : "#fb7185";

  const companyName = getMockCompany(symbol);
  const cardVolume = useMemo(() => {
    if (bars.length === 0) return "-";
    const total = bars.reduce((sum, b) => sum + b.volume, 0);
    return formatVolume(total / bars.length);
  }, [bars]);

  const hasChart = chartData.length > 1;

  return (
    <Link href={`watchlist/${symbol}`} className="block outline-none select-none group">
      <motion.div
        transition={{ duration: 0.3, ease: [0.22, 1, 0.36, 1] }}
        className="rounded-2xl"
      >
        <GlassSurface
          displace={15}
          distortionScale={-150}
          redOffset={5}
          greenOffset={15}
          blueOffset={25}
          brightness={60}
          opacity={0.8}
          mixBlendMode="overlay"
          flexDirection="col"
          alignItems="stretch"
          order="between"
          className="w-[320px] h-[220px] hover:cursor-pointer transition-colors duration-300 hover:border-white/25 hover:bg-black/50 rounded-2xl"
          innerClassName="p-5"
        >
          <div className="flex flex-col w-full pointer-events-none px-1">
            <div className="flex justify-between items-end w-full">
              <span className="font-mono font-bold text-white/95 text-2xl tracking-tighter leading-none drop-shadow-md">
                {symbol}
              </span>
              <span className="font-mono text-xl font-bold text-white/95 leading-none drop-shadow-md flex items-center">
                {latestPrice != null ? (
                  <>
                    <span className="opacity-80">$</span>
                    <SlidingNumber value={+latestPrice.toFixed(2)} />
                  </>
                ) : (
                  "-"
                )}
              </span>
            </div>
            <div className="flex justify-between items-baseline w-full mt-2">
              <span className="text-[10px] text-white/50 font-medium tracking-widest uppercase truncate max-w-[140px]">
                {companyName}
              </span>
              <AnimatePresence mode="wait">
                <motion.span
                  key={`${timeframe}-${isPositive}`}
                  initial={{ opacity: 0, y: 4 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, y: -4 }}
                  transition={{ duration: 0.2 }}
                  className={`font-mono text-xs font-semibold leading-none ${colorClass}`}
                >
                  {isPositive ? '+' : ''}{percentChange.toFixed(2)}%
                </motion.span>
              </AnimatePresence>
            </div>
          </div>

          <div className="w-full h-[50px] my-3 opacity-80 pointer-events-none">
            <AnimatePresence mode="wait">
              <motion.div
                key={`${symbol}-${timeframe}-${chartData.length}`}
                initial={{ opacity: 0, scale: 0.98 }}
                animate={{ opacity: 1, scale: 1 }}
                exit={{ opacity: 0 }}
                transition={{ duration: 0.35, ease: [0.22, 1, 0.36, 1] }}
                className="w-full h-full"
              >
                {hasChart ? (
                  <Sparklines data={chartData} margin={2} width={280} height={50}>
                    <SparklinesLine color={lineColor} style={{ strokeWidth: 2, fill: "none" }} />
                  </Sparklines>
                ) : isLoading ? (
                  <div className="w-full h-full flex items-center justify-center">
                    <Skeleton className="w-full h-0.5 rounded-full opacity-20" />
                  </div>
                ) : (
                  <div className="w-full h-full flex items-center justify-center text-[10px] uppercase tracking-widest text-white/30">
                    No data for {timeframe}
                  </div>
                )}
              </motion.div>
            </AnimatePresence>
          </div>

          <div className="w-full flex justify-center pointer-events-auto">
            <TimeframeSelector
              value={timeframe}
              onChange={setTimeframe}
              options={CARD_TIMEFRAMES}
              size="sm"
              stopPropagation
            />
          </div>

          <div className="flex justify-between items-center w-full mt-auto px-1 pt-3">
            <span className="text-[10px] font-medium text-white/30 tracking-widest uppercase pointer-events-none">
              Vol {cardVolume}
            </span>
            {onRemove && (
              <button
                onClick={(e) => {
                  e.preventDefault();
                  e.stopPropagation();
                  onRemove(symbol);
                }}
                className="text-[10px] font-medium text-white/30 hover:text-rose-300 tracking-widest uppercase flex items-center gap-1 transition-colors pointer-events-auto"
                title="Remove from watchlist"
              >
                <X size={10} strokeWidth={3} /> Remove
              </button>
            )}
          </div>
        </GlassSurface>
      </motion.div>
    </Link>
  );
}
