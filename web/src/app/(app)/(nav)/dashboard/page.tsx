'use client';

import { getAccount } from "@/services/account/queries";
import { useOrders } from "@/services/orders/queries";
import { getWatchlists } from "@/services/watchlists/queries";
import { useLatestNews } from "@/services/explore/queries";
import { ArrowRight, ArrowUpRight, BarChart3, FileText, Newspaper } from "lucide-react";
import Link from "next/link";
import GlassSurface from "@/components/primitives/GlassSurface";
import { TextEffect } from "@/components/primitives/TextEffect";
import { PageEnter, RevealStagger, RevealItem } from "@/components/primitives/Reveal";
import { SlidingNumber } from "@/components/primitives/SlidingNumber";
import { motion } from "motion/react";

type DashboardOrder = {
  side: string | number;
  symbol: string;
  order_type?: string;
  qty: string | number;
};

type DashboardWatchlist = {
  id: string | number;
  name: string;
};

type DashboardNewsItem = {
  id: string | number;
  url: string;
  symbols: string[];
  headline: string;
  summary: string;
  source: string;
  timestamp: string;
};

export default function Dashboard() {
  const { data: accountData } = getAccount();
  const { data: ordersData } = useOrders();
  const { data: watchlistsData } = getWatchlists();
  const { data: newsData } = useLatestNews();

  const userId = accountData?.user_name || "Trader";
  const orders: DashboardOrder[] = Array.isArray(ordersData)
    ? (ordersData as DashboardOrder[])
    : Array.isArray(ordersData?.orders)
      ? (ordersData.orders as DashboardOrder[])
      : [];
  const watchlists: DashboardWatchlist[] = Array.isArray(watchlistsData?.w_metadata)
    ? (watchlistsData.w_metadata as DashboardWatchlist[])
    : [];
  const newsItems: DashboardNewsItem[] = Array.isArray(newsData) ? (newsData as DashboardNewsItem[]) : [];

  const equity = (accountData?.equity_cents ?? 0) / 100;
  const buyingPower = (accountData?.buying_power_cents ?? 0) / 100;
  const openOrders = orders.filter(
    (o: DashboardOrder & { status?: string | number }) =>
      o && ((o as { status?: string | number }).status === undefined ||
        (o as { status?: string | number }).status === "OPEN" ||
        (o as { status?: string | number }).status === 1 ||
        (o as { status?: string | number }).status === 2)
  );

  return (
    <PageEnter className="min-h-screen w-full bg-transparent text-white font-mono p-4 pt-32 md:p-8 md:pt-32 xl:p-12 xl:pt-32 overflow-y-auto pointer-events-auto z-0 relative">
      <RevealStagger className="mx-auto max-w-7xl space-y-10 pb-32" stagger={0.08}>

        <RevealItem>
          <header className="flex flex-col md:flex-row md:items-end justify-between gap-6 pb-6 border-b border-white/10">
            <div className="space-y-3">
              <TextEffect
                as="h1"
                preset="fade-in-blur"
                per="word"
                className="text-4xl md:text-6xl font-light text-white tracking-tighter drop-shadow-md"
              >
                Welcome back.
              </TextEffect>
              <motion.p
                initial={{ opacity: 0, y: 6 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.4, duration: 0.4 }}
                className="text-sm text-white/50 flex items-center gap-2 font-medium tracking-wide uppercase"
              >
                {userId}
              </motion.p>
            </div>
          </header>
        </RevealItem>

        <RevealStagger className="grid grid-cols-1 md:grid-cols-3 gap-6" stagger={0.07} delay={0.1}>
          <RevealItem>
            <StatCard
              label="Account Value"
              valueNode={
                <span className="flex items-center">
                  <span className="text-white/60 mr-1">$</span>
                  <SlidingNumber value={+equity.toFixed(2)} />
                </span>
              }
              footer={
                <p className="text-xs font-medium text-white/40">
                  Cash + market value of holdings
                </p>
              }
            />
          </RevealItem>

          <RevealItem>
            <StatCard
              label="Active Orders"
              valueNode={
                <span className="flex items-baseline gap-3">
                  <SlidingNumber value={openOrders.length} />
                  <span className="text-white/40 text-sm font-medium pb-2 uppercase tracking-wider">
                    Working
                  </span>
                </span>
              }
            />
          </RevealItem>

          <RevealItem>
            <StatCard
              label="Buying Power"
              valueNode={
                <span className="flex items-center">
                  <span className="text-white/60 mr-1">$</span>
                  <SlidingNumber value={+buyingPower.toFixed(2)} />
                </span>
              }
              footer={
                <p className="text-xs font-medium text-white/40">
                  Cash not reserved by open orders
                </p>
              }
            />
          </RevealItem>
        </RevealStagger>

        <RevealStagger className="grid grid-cols-1 lg:grid-cols-3 gap-8" stagger={0.08} delay={0.15}>

          <RevealItem className="col-span-1">
            <motion.div transition={{ duration: 0.25 }} className="h-full">
              <GlassSurface displace={20} distortionScale={-120} opacity={0.8} borderRadius={32} className="flex items-start flex-1 min-h-[280px] md:min-h-[380px] h-full">
                <div className="p-5 md:p-8 w-full h-full flex flex-col">
                  <div className="flex justify-between items-center mb-6 md:mb-8">
                    <h3 className="text-xl font-medium text-white flex items-center gap-3">
                      Recent Activity
                    </h3>
                    <Link href="/orders" className="text-xs text-white/50 hover:text-white transition-colors flex items-center gap-1.5 px-3 py-1.5 rounded-full bg-white/5 hover:bg-white/10 hover:gap-2 uppercase tracking-widest">
                      Manage <ArrowRight size={12} />
                    </Link>
                  </div>

                  <div className="flex-1 flex flex-col justify-center space-y-4">
                    {orders.length > 0 ? (
                      orders.slice(0, 4).map((o, i: number) => {
                        const isBuy = o.side === "BUY" || o.side === 0;
                        return (
                          <motion.div
                            key={i}
                            initial={{ opacity: 0, x: -12 }}
                            animate={{ opacity: 1, x: 0 }}
                            transition={{ delay: 0.4 + i * 0.07, duration: 0.4 }}
                            className="flex justify-between items-center p-4 rounded-2xl bg-white/5 border border-white/10 hover:bg-white/10 transition-colors backdrop-blur-sm shadow-sm"
                          >
                            <div className="flex items-center gap-5">
                              <div className={`p-3 rounded-xl shadow-inner ${isBuy ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' : 'bg-orange-500/10 text-orange-400 border border-orange-500/20'}`}>
                                <ArrowUpRight size={18} className={isBuy ? 'rotate-90' : ''} />
                              </div>
                              <div className="space-y-1">
                                <p className="font-bold text-lg leading-none">{o.symbol}</p>
                                <p className="text-xs text-white/50 uppercase tracking-widest">{o.order_type || 'MARKET'} • {o.qty} shares</p>
                              </div>
                            </div>
                          </motion.div>
                        )
                      })
                    ) : (
                      <div className="flex flex-col items-center justify-center p-12 bg-black/20 rounded-2xl border border-white/5 border-dashed">
                        <FileText size={32} className="text-white/20 mb-4" />
                        <p className="text-white/50 font-medium">No recent orders found</p>
                        <p className="text-white/30 text-xs mt-2 text-center max-w-[200px]">Head over to the trade terminal to place your first order.</p>
                      </div>
                    )}
                  </div>
                </div>
              </GlassSurface>
            </motion.div>
          </RevealItem>

          <RevealItem className="col-span-1 lg:col-span-2">
            <motion.div transition={{ duration: 0.25 }} className="h-full">
              <GlassSurface displace={20} distortionScale={-120} opacity={0.8} borderRadius={32} className="flex items-start flex-1 min-h-[280px] md:min-h-[380px] h-full">
                <div className="p-5 md:p-8 w-full h-full flex flex-col">
                  <div className="flex justify-between items-center mb-6 md:mb-8">
                    <h3 className="text-xl font-medium text-white flex items-center gap-3">
                      Watchlists Tracker
                    </h3>
                    <Link href="/watchlist" className="text-xs text-white/50 hover:text-white transition-colors flex items-center gap-1.5 px-3 py-1.5 rounded-full bg-white/5 hover:bg-white/10 hover:gap-2 uppercase tracking-widest">
                      Explore <ArrowRight size={12} />
                    </Link>
                  </div>

                  <div className="flex-1 flex flex-col justify-center">
                    {watchlists.length > 0 ? (
                      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                        {watchlists.slice(0, 4).map((w, i) => (
                          <motion.div
                            key={w.id}
                            initial={{ opacity: 0, y: 12 }}
                            animate={{ opacity: 1, y: 0 }}
                            transition={{ delay: 0.45 + i * 0.07, duration: 0.4 }}
                          >
                            <Link href="/watchlist" className="block p-5 rounded-2xl bg-white/5 border border-white/10 hover:bg-white/10 hover:border-white/25 transition-all group overflow-hidden relative shadow-sm">
                              <div className="flex flex-col h-full justify-between gap-6 z-10 relative">
                                <p className="font-bold text-lg capitalize tracking-wide text-white/90 group-hover:text-white transition-colors">
                                  {w.name}
                                </p>
                                <div className="flex items-center justify-between">
                                  <p className="text-xs text-white/40 uppercase tracking-widest font-semibold">View List</p>
                                  <div className="p-1.5 rounded-full bg-white/5 border border-white/10 translate-x-4 opacity-0 group-hover:translate-x-0 group-hover:opacity-100 transition-all duration-300">
                                    <ArrowRight size={14} className="text-white" />
                                  </div>
                                </div>
                              </div>
                            </Link>
                          </motion.div>
                        ))}
                      </div>
                    ) : (
                      <div className="flex flex-col items-center justify-center p-12 bg-black/20 rounded-2xl border border-white/5 border-dashed">
                        <BarChart3 size={32} className="text-white/20 mb-4" />
                        <p className="text-white/50 font-medium">No watchlists tracked</p>
                        <p className="text-white/30 text-xs mt-2 text-center max-w-[200px]">Create your first personalized watchlist to track market movements.</p>
                      </div>
                    )}
                  </div>
                </div>
              </GlassSurface>
            </motion.div>
          </RevealItem>

        </RevealStagger>

        <RevealItem>
          <GlassSurface displace={20} distortionScale={-120} opacity={0.8} borderRadius={32} className="flex items-start w-full">
            <div className="p-5 md:p-8 w-full">
              <div className="flex justify-between items-center mb-6 md:mb-8">
                <h3 className="text-xl font-medium text-white flex items-center gap-3">
                  Market News
                </h3>
                <Link href="/explore" className="text-xs text-white/50 hover:text-white transition-colors flex items-center gap-1.5 px-3 py-1.5 rounded-full bg-white/5 hover:bg-white/10 hover:gap-2 uppercase tracking-widest">
                  Explore All <ArrowRight size={12} />
                </Link>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                {newsItems.length > 0 ? (
                  newsItems.slice(0, 3).map((news, i) => (
                    <motion.a
                      key={news.id}
                      href={news.url}
                      initial={{ opacity: 0, y: 14 }}
                      animate={{ opacity: 1, y: 0 }}
                      transition={{ delay: 0.5 + i * 0.08, duration: 0.45 }}
                      className="flex flex-col gap-4 p-5 rounded-2xl bg-white/5 border border-white/10 hover:bg-white/10 hover:border-white/25 transition-colors backdrop-blur-sm shadow-sm group"
                    >
                      <div className="flex items-center gap-2 flex-wrap">
                        {news.symbols.map((sym: string) => (
                          <span key={sym} className="px-2 py-0.5 rounded-md text-[10px] font-bold bg-white/10 border border-white/10 tracking-widest text-white/70">
                            {sym}
                          </span>
                        ))}
                      </div>
                      <h4 className="font-semibold text-lg text-white/90 leading-snug group-hover:text-white transition-colors">{news.headline}</h4>
                      <p className="text-xs text-white/50 line-clamp-2">{news.summary}</p>
                      <div className="flex items-center justify-between mt-auto pt-2 border-t border-white/5">
                        <span className="text-xs text-emerald-400 font-bold uppercase tracking-wider">{news.source}</span>
                        <span className="text-xs text-white/30">{news.timestamp}</span>
                      </div>
                    </motion.a>
                  ))
                ) : (
                  <div className="col-span-3 flex flex-col items-center justify-center p-12 bg-black/20 rounded-2xl border border-white/5 border-dashed">
                    <Newspaper size={32} className="text-white/20 mb-4" />
                    <p className="text-white/50 font-medium">No recent news available</p>
                  </div>
                )}
              </div>
            </div>
          </GlassSurface>
        </RevealItem>
      </RevealStagger>
    </PageEnter>
  );
}

function StatCard({
  label,
  valueNode,
  footer,
}: {
  label: string;
  valueNode: React.ReactNode;
  footer?: React.ReactNode;
}) {
  return (
    <motion.div transition={{ duration: 0.25 }}>
      <GlassSurface displace={10} distortionScale={-80} opacity={0.65} borderRadius={24} className="relative overflow-hidden group transition-colors duration-500 hover:border-white/25">
        <div className="p-5 sm:p-8 w-full relative z-10 flex flex-col justify-between h-full min-h-28 sm:min-h-40">
          <div className="flex items-center gap-2 text-white/50 mb-2">
            <h3 className="text-sm sm:text-xl font-semibold uppercase tracking-widest">{label}</h3>
          </div>
          <div className="flex flex-col gap-1 mt-3 sm:mt-4">
            <h2 className="text-3xl sm:text-4xl lg:text-5xl font-light tracking-tighter">
              {valueNode}
            </h2>
            {footer}
          </div>
        </div>
      </GlassSurface>
    </motion.div>
  );
}
