'use client';
import SymbolChart from '@/components/ui/SymbolChart';
import TimeframeSelector from '@/components/ui/TimeframeSelector';
import GlassSurface from '@/components/primitives/GlassSurface';
import { TextEffect } from '@/components/primitives/TextEffect';
import { PageEnter, RevealStagger, RevealItem } from '@/components/primitives/Reveal';
import { SlidingNumber } from '@/components/primitives/SlidingNumber';
import {
  ArrowLeft,
  TrendingDown,
  TrendingUp,
  Info,
  Activity,
  BarChart3,
  Gauge,
  LineChart,
} from 'lucide-react';
import Link from 'next/link';
import { use, useEffect, useMemo, useRef, useState } from 'react';
import {
  Timeframe,
  RealBar,
  useSymbolBars,
  toChartBars,
  computeAnalytics,
  windowReturnPct,
  formatVolume,
} from '@/lib/marketData';
import { mergeLiveBar } from '@/lib/liveBars';
import { socketForSymbols } from '@/lib/socket';
import { useMarketOpen } from '@/components/ui/MarketStatusBadge';
import { motion, AnimatePresence } from 'motion/react';
import PixelHover from '@/components/ui/PixelHover';

export default function Symbol({
  params,
}: {
  params: Promise<{ symbol: string }>;
}) {
  const { symbol } = use(params);
  const [timeframe, setTimeframe] = useState<Timeframe>('3M');
  const [livePrice, setLivePrice] = useState<number | undefined>(undefined);
  const marketOpen = useMarketOpen();

  const { data: bars = [], isLoading: barsLoading } = useSymbolBars(symbol, timeframe);
  const { data: dailyBars = [] } = useSymbolBars(symbol, '1Y');

  const [liveBars, setLiveBars] = useState<RealBar[] | null>(null);
  useEffect(() => setLiveBars(null), [bars, timeframe]);

  const barsRef = useRef(bars);
  barsRef.current = bars;
  const tfRef = useRef(timeframe);
  tfRef.current = timeframe;

  const chartBars = useMemo(
    () => toChartBars(liveBars ?? bars, timeframe),
    [liveBars, bars, timeframe]
  );
  const analytics = useMemo(
    () => computeAnalytics(dailyBars, livePrice),
    [dailyBars, livePrice]
  );

  useEffect(() => {
    if (!symbol) return;
    const socket = socketForSymbols([symbol]);
    socket.addEventListener('message', (e) => {
      const msg = JSON.parse(e.data);
      for (const bar of msg.symbol_bar ?? []) {
        if (bar.symbol !== symbol || !(bar.close > 0)) continue;
        setLivePrice(bar.close);
        const wireBar: RealBar = {
          time: bar.time || new Date().toISOString(),
          open: bar.open ?? bar.close,
          high: bar.high ?? bar.close,
          low: bar.low ?? bar.close,
          close: bar.close,
          volume: bar.volume ?? 0,
        };
        setLiveBars((prev) =>
          mergeLiveBar(prev ?? barsRef.current, wireBar, tfRef.current)
        );
      }
    });
    return () => socket.close();
  }, [symbol]);

  const isGainer = analytics.change >= 0;
  const rangePosPct = Math.max(0, Math.min(1, analytics.range52Position)) * 100;

  const rsiTone =
    analytics.rsi >= 70
      ? { label: 'Overbought', color: 'text-rose-400', bar: 'bg-rose-400/80' }
      : analytics.rsi <= 30
        ? { label: 'Oversold', color: 'text-emerald-400', bar: 'bg-emerald-400/80' }
        : { label: 'Neutral', color: 'text-amber-300', bar: 'bg-amber-300/80' };

  const trendVsSma20 = analytics.last >= analytics.sma20 ? 'Above' : 'Below';
  const trendVsSma50 = analytics.last >= analytics.sma50 ? 'Above' : 'Below';

  const performanceWindows = useMemo(() => {
    const windows: { label: string; days: number }[] = [
      { label: '1D', days: 1 },
      { label: '1W', days: 5 },
      { label: '1M', days: 21 },
      { label: '3M', days: 63 },
      { label: '6M', days: 126 },
      { label: '1Y', days: 252 },
    ];
    return windows.map(({ label, days }) => ({
      label,
      pct: windowReturnPct(dailyBars, days),
    }));
  }, [dailyBars]);

  return (
    <PageEnter className="min-h-screen w-full bg-transparent text-white font-mono p-4 pt-32 md:p-8 md:pt-32 xl:p-12 xl:pt-32 overflow-y-auto pointer-events-auto z-0 relative">
      <RevealStagger className="mx-auto max-w-[1400px] flex flex-col space-y-8 pb-32 h-full" stagger={0.08}>
        <RevealItem>
          <header className="flex flex-col md:flex-row md:items-end justify-between gap-6 pb-6 border-b border-white/10">
            <div className="space-y-6">
              <motion.div
                initial={{ opacity: 0, x: -10 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ duration: 0.4, delay: 0.05 }}
              >
                <Link
                  href="/watchlist"
                  className="flex items-center gap-2 text-[10px] uppercase tracking-widest text-white/50 hover:text-white transition-colors w-fit border border-white/10 px-3 py-1.5 rounded-full bg-white/5 hover:bg-white/10 hover:gap-3"
                >
                  <ArrowLeft size={12} /> Back to Watchlists
                </Link>
              </motion.div>

              <div className="flex items-end gap-6">
                <TextEffect
                  as="h1"
                  preset="fade-in-blur"
                  per="char"
                  speedReveal={1.5}
                  className="text-4xl md:text-6xl font-black text-white tracking-tighter drop-shadow-md"
                >
                  {symbol}
                </TextEffect>
                <motion.div
                  initial={{ opacity: 0, y: 8 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.4, delay: 0.5 }}
                  className="hidden sm:flex flex-col pb-1"
                >
                  <h2 className="text-3xl font-light text-white/90 flex items-center">
                    <span className="text-white/60 mr-1">$</span>
                    <SlidingNumber value={+analytics.last.toFixed(2)} />
                  </h2>
                  <span
                    className={`text-sm font-bold flex items-center gap-1.5 ${isGainer ? 'text-emerald-400' : 'text-rose-400'
                      }`}
                  >
                    {isGainer ? <TrendingUp size={14} /> : <TrendingDown size={14} />}
                    {isGainer ? '+' : '-'}
                    <SlidingNumber value={+Math.abs(analytics.change).toFixed(2)} />
                    <span className="opacity-70">
                      ({isGainer ? '+' : '-'}
                      <SlidingNumber value={+Math.abs(analytics.changePct).toFixed(2)} />%)
                    </span>
                  </span>
                </motion.div>
              </div>
            </div>

            <div className="flex flex-wrap items-center gap-3">
              <motion.div
                initial={{ opacity: 0, scale: 0.95 }}
                animate={{ opacity: 1, scale: 1 }}
                transition={{ delay: 0.55, duration: 0.3 }}
                className="px-4 py-2 bg-black/40 border border-white/10 rounded-lg text-[10px] font-semibold tracking-widest uppercase backdrop-blur-md flex items-center gap-2 text-white/70"
                title={
                  marketOpen
                    ? 'US market open · 9:30 AM - 4:00 PM ET'
                    : 'US market closed · resumes 9:30 AM ET'
                }
              >
                {marketOpen ? (
                  <>
                    <motion.span
                      className="flex h-1.5 w-1.5 rounded-full bg-emerald-400"
                      animate={{
                        scale: [1, 1.5, 1],
                        boxShadow: [
                          '0 0 6px rgba(16,185,129,0.7)',
                          '0 0 14px rgba(16,185,129,0.95)',
                          '0 0 6px rgba(16,185,129,0.7)',
                        ],
                      }}
                      transition={{ duration: 1.8, repeat: Infinity, ease: 'easeInOut' }}
                    />
                    Live · IEX Feed
                  </>
                ) : (
                  <>
                    <span
                      className={`flex h-1.5 w-1.5 rounded-full ${marketOpen === null ? 'bg-white/40' : 'bg-amber-400/80'
                        }`}
                    />
                    {marketOpen === null ? 'IEX Feed' : 'Market Closed'}
                  </>
                )}
              </motion.div>
            </div>
          </header>
        </RevealItem>

        <section className="grid grid-cols-1 lg:grid-cols-4 gap-6 min-h-[560px]">
          <RevealItem className="col-span-1 lg:col-span-3">
            <GlassSurface
              displace={15}
              distortionScale={-150}
              opacity={0.6}
              borderRadius={32}
              className="flex overflow-hidden w-full relative h-full"
            >
              <div className="p-5 w-full h-full min-h-[520px] flex flex-col">
                <div className="flex items-center justify-between gap-4 mb-4 flex-wrap">
                  <div className="flex items-center gap-2 text-white/70 text-xs uppercase tracking-widest font-semibold">
                    <LineChart size={14} className="text-emerald-400" />
                    Price · {timeframe}
                  </div>
                  <TimeframeSelector value={timeframe} onChange={setTimeframe} />
                </div>
                <AnimatePresence mode="wait">
                  <motion.div
                    key={timeframe}
                    initial={{ opacity: 0, y: 8 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0, y: -4 }}
                    transition={{ duration: 0.35, ease: [0.22, 1, 0.36, 1] }}
                    className="flex-1 min-h-[440px]"
                  >
                    {chartBars.length > 0 ? (
                      <SymbolChart symbol={symbol} prices={chartBars} />
                    ) : (
                      <div className="h-full min-h-[440px] flex items-center justify-center text-white/40 text-sm">
                        {barsLoading ? 'Loading chart…' : 'No price history available for this range.'}
                      </div>
                    )}
                  </motion.div>
                </AnimatePresence>
              </div>
            </GlassSurface>
          </RevealItem>

          <RevealItem className="col-span-1 flex flex-col gap-6">
            <GlassSurface
              displace={20}
              distortionScale={-80}
              opacity={0.8}
              borderRadius={32}
              className="flex-1 flex flex-col h-full"
            >
              <div className="p-7 w-full flex flex-col gap-7 h-full">
                <div className="space-y-2">
                  <h3 className="text-base font-semibold tracking-tight text-white/90 flex items-center gap-2">
                    <Info size={14} />
                    Session Stats
                  </h3>
                </div>

                <div className="grid grid-cols-2 gap-y-5 gap-x-4">
                  <SlidingStat label="Open" value={+analytics.open.toFixed(2)} prefix="$" />
                  <SlidingStat label="Prev Close" value={+analytics.prevClose.toFixed(2)} prefix="$" />
                  <SlidingStat
                    label="High"
                    value={+analytics.high.toFixed(2)}
                    prefix="$"
                    valueClass="text-emerald-400"
                  />
                  <SlidingStat
                    label="Low"
                    value={+analytics.low.toFixed(2)}
                    prefix="$"
                    valueClass="text-rose-400"
                  />
                  <Stat label="Volume" value={formatVolume(analytics.volume)} />
                  <Stat label="Avg Vol" value={formatVolume(analytics.avgVolume)} />
                </div>

                <div>
                  <p className="text-[10px] uppercase tracking-widest text-white/40 mb-2">
                    52-Week Range
                  </p>
                  <div className="relative h-1.5 bg-white/5 rounded-full overflow-visible mb-1">
                    <motion.div
                      className="absolute top-1/2 -translate-y-1/2 h-2 w-2 rounded-full bg-white shadow-[0_0_10px_rgba(255,255,255,0.7)]"
                      initial={{ left: '0%', opacity: 0 }}
                      animate={{ left: `calc(${rangePosPct}% - 4px)`, opacity: 1 }}
                      transition={{ duration: 0.9, ease: [0.22, 1, 0.36, 1], delay: 0.3 }}
                    />
                  </div>
                  <div className="flex justify-between text-[10px] text-white/50 mt-1.5">
                    <span>${analytics.fiftyTwoLow.toFixed(2)}</span>
                    <span>${analytics.fiftyTwoHigh.toFixed(2)}</span>
                  </div>
                </div>

                <div className="mt-auto pt-5 border-t border-white/10 space-y-3">
                  <Link href="/orders" className="block group">
                    <PixelHover
                      variant='emerald'
                      className="w-full text-center px-4 py-3 hover:bg-emerald-500/20 hover:text-emerald-400 transition-colors border border-white/25 hover:border-emerald-500/30 rounded-xl text-xs font-bold uppercase tracking-widest backdrop-blur-md"
                    >
                      Execute Trade
                    </PixelHover>
                  </Link>
                  <p className="text-[10px] text-center text-white/40 leading-relaxed px-2">
                    Open the Terminal to place limit / stop-loss structures for {symbol}.
                  </p>
                </div>
              </div>
            </GlassSurface>
          </RevealItem>
        </section>

        <RevealStagger
          className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-6"
          stagger={0.08}
          delay={0.1}
        >
          <RevealItem className="col-span-1 md:col-span-2">
            <PanelCard>
              <PanelHeader icon={<Activity size={14} />} title="Performance" />
              <div className="grid grid-cols-3 sm:grid-cols-6 gap-3 mt-5">
                {performanceWindows.map((w, i) => {
                  const positive = w.pct >= 0;
                  return (
                    <motion.div
                      key={w.label}
                      initial={{ opacity: 0, y: 10 }}
                      animate={{ opacity: 1, y: 0 }}
                      transition={{ delay: 0.4 + i * 0.05, duration: 0.35 }}
                      className="rounded-xl border border-white/5 bg-white/2 hover:bg-white/5 hover:border-white/15 transition-colors px-3 py-3 flex flex-col items-center gap-1 cursor-default"
                    >
                      <span className="text-[10px] uppercase tracking-widest text-white/40">
                        {w.label}
                      </span>
                      <span
                        className={`text-sm font-bold ${positive ? 'text-emerald-400' : 'text-rose-400'
                          }`}
                      >
                        {positive ? '+' : ''}
                        {w.pct.toFixed(2)}%
                      </span>
                    </motion.div>
                  );
                })}
              </div>
            </PanelCard>
          </RevealItem>

          <RevealItem>
            <PanelCard>
              <PanelHeader icon={<Gauge size={14} />} title="Momentum · RSI(14)" />
              <div className="mt-6 flex flex-col gap-3">
                <div className="flex items-end justify-between">
                  <span className="text-3xl font-bold tracking-tight flex items-center">
                    <SlidingNumber value={+analytics.rsi.toFixed(1)} />
                  </span>
                  <span className={`text-xs font-bold uppercase tracking-widest ${rsiTone.color}`}>
                    {rsiTone.label}
                  </span>
                </div>
                <div className="relative h-1.5 bg-white/5 rounded-full overflow-hidden">
                  <motion.div
                    className={`absolute top-0 left-0 h-full ${rsiTone.bar} rounded-full`}
                    initial={{ width: 0 }}
                    animate={{ width: `${analytics.rsi}%` }}
                    transition={{ duration: 0.9, ease: [0.22, 1, 0.36, 1], delay: 0.3 }}
                  />
                  <div className="absolute top-0 h-full w-px bg-white/15" style={{ left: '30%' }} />
                  <div className="absolute top-0 h-full w-px bg-white/15" style={{ left: '70%' }} />
                </div>
                <div className="flex justify-between text-[10px] text-white/40 uppercase tracking-widest">
                  <span>Oversold</span>
                  <span>Overbought</span>
                </div>
              </div>
            </PanelCard>
          </RevealItem>

          <RevealItem>
            <PanelCard>
              <PanelHeader icon={<BarChart3 size={14} />} title="Technicals" />
              <div className="mt-5 space-y-3.5">
                <Row
                  label="SMA 20"
                  value={`$${analytics.sma20.toFixed(2)}`}
                  badge={trendVsSma20}
                  badgeTone={trendVsSma20 === 'Above' ? 'pos' : 'neg'}
                />
                <Row
                  label="SMA 50"
                  value={`$${analytics.sma50.toFixed(2)}`}
                  badge={trendVsSma50}
                  badgeTone={trendVsSma50 === 'Above' ? 'pos' : 'neg'}
                />
                <Row label="Volatility" value={`${analytics.volatility.toFixed(1)}%`} subtle />
              </div>
            </PanelCard>
          </RevealItem>

          <RevealItem className="col-span-1 md:col-span-2">
            <PanelCard>
              <PanelHeader icon={<Info size={14} className="text-white/70" />} title={`About ${symbol}`} />
              <TextEffect
                as="p"
                preset="fade"
                per="word"
                delay={0.3}
                speedReveal={1.4}
                className="mt-4 text-xs leading-relaxed text-white/60"
              >
                {`${symbol} is currently trading ${isGainer ? 'up' : 'down'} ${analytics.changePct.toFixed(2)}% versus the prior session close. The asset is sitting at the ${rangePosPct.toFixed(0)}th percentile of its 52-week range, with momentum currently reading ${rsiTone.label.toLowerCase()} on the RSI.`}
              </TextEffect>
              <p className="mt-3 text-[10px] uppercase tracking-widest text-white/30">
                Metrics derived from one year of daily market history (split-adjusted).
              </p>
            </PanelCard>
          </RevealItem>
        </RevealStagger>
      </RevealStagger>
    </PageEnter>
  );
}

function PanelCard({ children }: { children: React.ReactNode }) {
  return (
    <motion.div
      transition={{ duration: 0.25, ease: [0.22, 1, 0.36, 1] }}
      className="h-full"
    >
      <GlassSurface borderRadius={24} opacity={0.7} className="h-full">
        <div className="p-6 w-full h-full">{children}</div>
      </GlassSurface>
    </motion.div>
  );
}

function PanelHeader({ icon, title }: { icon: React.ReactNode; title: string }) {
  return (
    <div className="flex items-center gap-2 text-[11px] font-semibold uppercase tracking-widest text-white/70">
      {icon}
      {title}
    </div>
  );
}

function Stat({
  label,
  value,
  valueClass = '',
}: {
  label: string;
  value: string;
  valueClass?: string;
}) {
  return (
    <div>
      <p className="text-[10px] uppercase tracking-widest text-white/40 mb-1">{label}</p>
      <p className={`font-semibold text-sm ${valueClass}`}>{value}</p>
    </div>
  );
}

function SlidingStat({
  label,
  value,
  prefix = '',
  valueClass = '',
}: {
  label: string;
  value: number;
  prefix?: string;
  valueClass?: string;
}) {
  return (
    <div>
      <p className="text-[10px] uppercase tracking-widest text-white/40 mb-1">{label}</p>
      <p className={`font-semibold text-sm flex items-center ${valueClass}`}>
        {prefix && <span className="opacity-80 mr-0.5">{prefix}</span>}
        <SlidingNumber value={value} />
      </p>
    </div>
  );
}

function Row({
  label,
  value,
  badge,
  badgeTone,
  subtle = false,
}: {
  label: string;
  value: string;
  badge?: string;
  badgeTone?: 'pos' | 'neg';
  subtle?: boolean;
}) {
  return (
    <div className="flex items-center justify-between">
      <span className="text-xs text-white/50">{label}</span>
      <div className="flex items-center gap-2">
        <span className={`text-sm font-semibold ${subtle ? 'text-white/80' : 'text-white'}`}>
          {value}
        </span>
        {badge && (
          <span
            className={`text-[9px] font-bold uppercase tracking-widest px-1.5 py-0.5 rounded-md border ${badgeTone === 'pos'
              ? 'text-emerald-400 border-emerald-400/30 bg-emerald-400/10'
              : 'text-rose-400 border-rose-400/30 bg-rose-400/10'
              }`}
          >
            {badge}
          </span>
        )}
      </div>
    </div>
  );
}
