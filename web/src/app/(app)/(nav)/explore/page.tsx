'use client';

import Link from "next/link";
import { useEffect, useMemo, useState } from "react";
import { useLatestNews, useTopMovers } from "@/services/explore/queries";
import { ArrowUpRight, Hash, TrendingDown, TrendingUp } from "lucide-react";
import GlassSurface from "@/components/primitives/GlassSurface";
import { TextEffect } from "@/components/primitives/TextEffect";
import { PageEnter, RevealStagger, RevealItem } from "@/components/primitives/Reveal";
import { SlidingNumber } from "@/components/primitives/SlidingNumber";
import MarketStatusBadge from "@/components/ui/MarketStatusBadge";
import { socketForSymbols } from "@/lib/socket";
import { motion } from "motion/react";

const MAX_LIVE_MOVERS = 12;

type TopMover = {
  symbol: string;
  name: string;
  type: "gainer" | "loser";
  price: number;
  changePercent: number;
};

type NewsItem = {
  id: string | number;
  imageUrl: string;
  symbols: string[];
  timestamp: string;
  headline: string;
  summary: string;
  source: string;
  url: string;
};

export default function Explore() {
  const { data: newsData } = useLatestNews();
  const { data: moversData } = useTopMovers();

  const newsItems: NewsItem[] = (newsData ?? []) as NewsItem[];
  const movers: TopMover[] = useMemo(() => (moversData ?? []) as TopMover[], [moversData]);

  const [closePrices, setClosePrices] = useState<Record<string, number>>({});
  const moverSymbolsKey = useMemo(
    () =>
      movers
        .slice(0, MAX_LIVE_MOVERS)
        .map((m) => m.symbol)
        .sort()
        .join(","),
    [movers]
  );

  useEffect(() => {
    if (!moverSymbolsKey) return;
    const socket = socketForSymbols(moverSymbolsKey.split(","));
    socket.addEventListener("message", (e) => {
      const msg = JSON.parse(e.data);
      if (!msg.symbol_bar) return;
      const updates: Record<string, number> = {};
      for (const bar of msg.symbol_bar) {
        if (bar.close > 0) updates[bar.symbol] = bar.close;
      }
      setClosePrices((prev) => ({ ...prev, ...updates }));
    });
    return () => socket.close();
  }, [moverSymbolsKey]);

  return (
    <PageEnter className="min-h-screen w-full bg-transparent text-white font-mono p-4 pt-32 md:p-8 md:pt-32 xl:p-12 xl:pt-32 overflow-y-auto pointer-events-auto z-0 relative">
      <RevealStagger className="mx-auto max-w-7xl space-y-12 pb-32" stagger={0.08}>

        <RevealItem>
          <header className="flex flex-col md:flex-row md:items-end justify-between gap-6 pb-6 border-b border-white/10">
            <div className="space-y-3">
              <TextEffect
                as="h1"
                preset="fade-in-blur"
                per="word"
                className="text-4xl md:text-6xl font-light text-white tracking-tighter drop-shadow-md flex items-center gap-4"
              >
                Explore Markets
              </TextEffect>
              <motion.div
                initial={{ opacity: 0, y: 6 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.4, duration: 0.4 }}
                className="text-sm text-white/50 font-medium tracking-wide uppercase"
              >
                <MarketStatusBadge />
              </motion.div>
            </div>
          </header>
        </RevealItem>

        <RevealItem className="space-y-6">
          <div className="flex items-center gap-2 text-white/80">
            <h2 className="text-xl font-medium tracking-tight">Today&apos;s Top Movers</h2>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
            {movers.map((mover: TopMover, idx: number) => {
              const isGainer = mover.type === 'gainer';
              const livePrice = closePrices[mover.symbol] ?? mover.price;
              return (
                <motion.div
                  key={mover.symbol + idx}
                  initial={{ opacity: 0, y: 14, scale: 0.97 }}
                  animate={{ opacity: 1, y: 0, scale: 1 }}
                  transition={{ delay: 0.2 + idx * 0.04, duration: 0.4, ease: [0.22, 1, 0.36, 1] }}
                >
                  <Link href={`/watchlist/${mover.symbol}`} className="block h-full">
                    <GlassSurface
                      displace={15}
                      distortionScale={-80}
                      opacity={0.65}
                      borderRadius={20}
                      className="h-full hover:border-white/25 transition-colors"
                    >
                      <div className="p-4 w-full flex flex-col gap-3 min-w-0">
                        <div className="flex items-center justify-between gap-2 min-w-0">
                          <div className="min-w-0">
                            <h3 className="text-base font-bold tracking-tight text-white">{mover.symbol}</h3>
                            <p className="text-[10px] text-white/45 truncate uppercase font-semibold">{mover.name}</p>
                          </div>
                          <div className={`p-1.5 rounded-lg shrink-0 backdrop-blur-md border ${isGainer ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20' : 'bg-red-500/10 text-red-400 border-red-500/20'}`}>
                            {isGainer ? <TrendingUp size={16} /> : <TrendingDown size={16} />}
                          </div>
                        </div>
                        <div className="flex items-end justify-between gap-2">
                          <span className="text-xl font-light text-white/90 tabular-nums flex items-center">
                            <span className="opacity-70 mr-0.5">$</span>
                            <SlidingNumber value={+livePrice.toFixed(2)} />
                          </span>
                          <span className={`text-xs font-bold tabular-nums shrink-0 ${isGainer ? 'text-emerald-400' : 'text-red-400'}`}>
                            {isGainer ? '+' : ''}{mover.changePercent.toFixed(2)}%
                          </span>
                        </div>
                      </div>
                    </GlassSurface>
                  </Link>
                </motion.div>
              );
            })}
            {movers.length === 0 && (
              <div className="col-span-full py-10 text-center text-white/40 text-sm">
                Loading market movers…
              </div>
            )}
          </div>
        </RevealItem>

        <RevealItem className="space-y-6">
          <div className="flex items-center gap-2 text-white/80">
            <h2 className="text-xl font-medium tracking-tight">Financial News</h2>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
            {newsItems.map((news: NewsItem, idx: number) => (
              <motion.div
                key={`${news.id}-${idx}`}
                initial={{ opacity: 0, y: 22 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.35 + idx * 0.08, duration: 0.5, ease: [0.22, 1, 0.36, 1] }}
                className={`col-span-1 ${idx === 0 ? 'md:col-span-2' : ''}`}
              >
                <GlassSurface
                  displace={25}
                  distortionScale={-100}
                  opacity={0.7}
                  borderRadius={24}
                  className="flex flex-col h-full group overflow-hidden hover:border-white/25 transition-colors"
                >
                  <div className="p-6 md:p-8 flex flex-col h-full min-w-0">
                    <div className="flex items-center gap-2 flex-wrap mb-4">
                      {news.symbols.slice(0, 4).map((sym: string) => (
                        <span key={sym} className="px-2.5 py-1 rounded-md text-[10px] font-bold bg-white/20 border border-white/20 tracking-widest text-white shadow-sm backdrop-blur-md">
                          <Hash size={10} className="inline mr-0.5 opacity-50" />{sym}
                        </span>
                      ))}
                      {news.symbols.length > 4 && (
                        <span className="px-2 py-1 rounded-md text-[10px] font-bold bg-white/10 border border-white/10 text-white/50">
                          +{news.symbols.length - 4}
                        </span>
                      )}
                      <span className="ml-auto text-xs text-white/60 font-semibold uppercase tracking-wider shrink-0">{news.timestamp}</span>
                    </div>

                    <h3 className={`font-semibold text-white/90 leading-tight group-hover:text-white transition-colors drop-shadow-sm wrap-break-word line-clamp-2 ${idx === 0 ? 'text-2xl md:text-4xl' : 'text-xl md:text-2xl'}`}>
                      {news.headline}
                    </h3>

                    <p className={`text-white/60 mt-4 leading-relaxed wrap-break-word line-clamp-3 ${idx === 0 ? 'text-base max-w-3xl' : 'text-sm'}`}>
                      {news.summary}
                    </p>

                    <div className="flex items-center justify-between mt-auto pt-6">
                      <span className="text-sm text-emerald-400 font-bold uppercase tracking-wider">{news.source}</span>
                      <a
                        href={news.url}
                        className="flex items-center gap-1.5 text-xs font-semibold uppercase tracking-widest text-white/40 hover:text-white hover:gap-2.5 transition-all bg-white/5 hover:bg-white/10 px-4 py-2 rounded-full border border-white/10"
                      >
                        Read Report <ArrowUpRight size={14} className="opacity-70" />
                      </a>
                    </div>
                  </div>
                </GlassSurface>
              </motion.div>
            ))}
          </div>
        </RevealItem>

      </RevealStagger>
    </PageEnter>
  );
}
