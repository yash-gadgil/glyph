'use client';

import { use, useMemo } from "react";
import Link from "next/link";
import GlassSurface from "@/components/primitives/GlassSurface";
import { TextEffect } from "@/components/primitives/TextEffect";
import { PageEnter, RevealStagger, RevealItem } from "@/components/primitives/Reveal";
import { formatCents, timeAgo } from "@/lib/utils";
import {
  useDeployments,
  useStrategyTrades,
} from "@/services/strategies/queries";
import { useStopDeployment } from "@/services/strategies/mutations";
import {
  ArrowLeft,
  CircleDot,
  Cpu,
  History,
  Square,
  TrendingDown,
  TrendingUp,
} from "lucide-react";
import { motion } from "motion/react";

export default function StrategyDetail({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id: strategyId } = use(params);

  const { data: deployments = [] } = useDeployments();
  const { data: fills = [], isLoading: fillsLoading } = useStrategyTrades(strategyId);
  const stopMutation = useStopDeployment();

  const deployment = useMemo(() => {
    const forStrategy = deployments.filter((d) => d.strategy_id === strategyId);
    return forStrategy.find((d) => d.status === "running") ?? forStrategy[0];
  }, [deployments, strategyId]);

  const stats = useMemo(() => {
    let bought = 0, sold = 0, boughtQty = 0, soldQty = 0;
    for (const f of fills) {
      const value = f.qty * f.price_cents;
      if (f.side === "buy") { bought += value; boughtQty += f.qty; }
      else { sold += value; soldQty += f.qty; }
    }
    const closedQty = Math.min(boughtQty, soldQty);
    const avgBuy = boughtQty > 0 ? bought / boughtQty : 0;
    const avgSell = soldQty > 0 ? sold / soldQty : 0;
    const realized = closedQty > 0 ? Math.round((avgSell - avgBuy) * closedQty) : 0;
    return { trades: fills.length, realized };
  }, [fills]);

  const name = deployment?.strategy_name || "Strategy";

  return (
    <PageEnter className="min-h-screen w-full bg-transparent text-white font-mono p-4 pt-32 md:p-8 md:pt-32 xl:p-12 xl:pt-32 overflow-y-auto pointer-events-auto z-0 relative">
      <RevealStagger className="mx-auto max-w-5xl space-y-8 pb-32" stagger={0.08}>
        <RevealItem>
          <header className="space-y-6 pb-6 border-b border-white/10">
            <Link
              href="/strategies"
              className="flex items-center gap-2 text-[10px] uppercase tracking-widest text-white/50 hover:text-white transition-colors w-fit border border-white/10 px-3 py-1.5 rounded-full bg-white/5 hover:bg-white/10 hover:gap-3"
            >
              <ArrowLeft size={12} /> Back to Strategies
            </Link>

            <div className="flex flex-col md:flex-row md:items-end justify-between gap-4">
              <div>
                <TextEffect
                  as="h1"
                  preset="fade-in-blur"
                  per="word"
                  className="text-3xl md:text-5xl font-light text-white tracking-tighter drop-shadow-md"
                >
                  {name}
                </TextEffect>
                {deployment && (
                  <p className="text-sm text-white/50 mt-2 flex items-center gap-3">
                    <span className="font-semibold tracking-wider">{deployment.symbol}</span>

                  </p>
                )}
              </div>

              {deployment?.status === "running" && (
                <motion.button
                  whileTap={{ scale: 0.95 }}
                  onClick={() => stopMutation.mutate(deployment.id)}
                  disabled={stopMutation.isPending}
                  className="flex items-center gap-2 px-4 py-2.5 rounded-xl text-xs font-bold uppercase tracking-wider text-red-400 border border-red-500/20 bg-red-500/10 hover:bg-red-500/20 transition-colors disabled:opacity-50 self-start"
                >
                  <Square size={12} />
                  {stopMutation.isPending ? "Stopping…" : "Stop Automation"}
                </motion.button>
              )}
            </div>
          </header>
        </RevealItem>

        <RevealItem>
          <section className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <StatCard
              label="Status"
              value={deployment ? deployment.status : "not deployed"}
              tone={deployment?.status === "running" ? "pos" : "muted"}
            />
            <StatCard
              label="Position"
              value={
                deployment?.in_position
                  ? `${deployment.qty} @ $${formatCents(deployment.entry_price_cents)}`
                  : "Flat"
              }
              tone={deployment?.in_position ? "pos" : "muted"}
            />
            <StatCard label="Trades" value={String(stats.trades)} tone="muted" />
            <StatCard
              label="Realized P&L"
              value={`${stats.realized >= 0 ? "+" : "-"}$${formatCents(Math.abs(stats.realized))}`}
              tone={stats.realized >= 0 ? "pos" : "neg"}
            />
          </section>
        </RevealItem>

        <RevealItem>
          <section className="space-y-4">
            <h2 className="text-lg font-semibold tracking-tight text-white flex items-center gap-2">
              <History size={18} />
              Trades
            </h2>

            <GlassSurface borderRadius={16} order="start" alignItems="stretch" flexDirection="col">
              <div className="overflow-x-auto w-full">
                <table className="w-full text-left text-sm whitespace-nowrap">
                  <thead className="border-b border-white/10 text-white/40">
                    <tr>
                      <th className="px-6 py-4 font-semibold text-xs tracking-wider uppercase">Side</th>
                      <th className="px-6 py-4 font-semibold text-xs tracking-wider uppercase">Symbol</th>
                      <th className="px-6 py-4 font-semibold text-xs tracking-wider uppercase text-right">Qty</th>
                      <th className="px-6 py-4 font-semibold text-xs tracking-wider uppercase text-right">Price</th>
                      <th className="px-6 py-4 font-semibold text-xs tracking-wider uppercase text-right">Value</th>
                      <th className="px-6 py-4 font-semibold text-xs tracking-wider uppercase text-right">Executed</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-white/5">
                    {fills.length > 0 ? (
                      fills.map((f, idx) => (
                        <motion.tr
                          key={`${f.trade_id}-${f.order_id}`}
                          initial={{ opacity: 0, x: -12 }}
                          animate={{ opacity: 1, x: 0 }}
                          transition={{ delay: 0.2 + idx * 0.04, duration: 0.35 }}
                          className="hover:bg-white/5 transition-colors"
                        >
                          <td className="px-6 py-4">
                            <span className={`inline-flex items-center gap-1.5 text-xs font-bold uppercase px-2 py-1 rounded-md border ${f.side === "buy"
                              ? "text-emerald-400 border-emerald-500/20 bg-emerald-500/10"
                              : "text-red-400 border-red-500/20 bg-red-500/10"
                              }`}>
                              {f.side === "buy" ? <TrendingUp size={11} /> : <TrendingDown size={11} />}
                              {f.side}
                            </span>
                          </td>
                          <td className="px-6 py-4 font-semibold text-white/80">{f.symbol}</td>
                          <td className="px-6 py-4 text-right text-white/70">{f.qty}</td>
                          <td className="px-6 py-4 text-right text-white/70">${formatCents(f.price_cents)}</td>
                          <td className="px-6 py-4 text-right text-white/70">${formatCents(f.qty * f.price_cents)}</td>
                          <td className="px-6 py-4 text-right text-white/40 text-xs">{timeAgo(f.executed_at)}</td>
                        </motion.tr>
                      ))
                    ) : (
                      <tr>
                        <td colSpan={6} className="px-6 py-16 text-center text-white/40 text-sm">
                          {fillsLoading
                            ? "Loading trades…"
                            : "No trades yet, the engine places orders as the strategy's rules fire (checked once a minute during market hours)."}
                        </td>
                      </tr>
                    )}
                  </tbody>
                </table>
              </div>
            </GlassSurface>
          </section>
        </RevealItem>

        {!deployment && (
          <RevealItem>
            <div className="flex items-start gap-3 px-5 py-4 rounded-xl border border-white/5 bg-white/3 text-xs text-white/40 font-medium leading-relaxed">
              <Cpu size={14} className="shrink-0 mt-0.5 text-white/30" />
              This strategy isn&apos;t deployed right now. Deploy it from the
              Strategies page to start automated trading; past trades stay
              recorded here.
            </div>
          </RevealItem>
        )}
      </RevealStagger>
    </PageEnter>
  );
}

function StatCard({ label, value, tone }: { label: string; value: string; tone: "pos" | "neg" | "muted" }) {
  const color =
    tone === "pos" ? "text-emerald-400" : tone === "neg" ? "text-red-400" : "text-white/80";
  return (
    <GlassSurface borderRadius={16} order="start" alignItems="stretch" flexDirection="col" innerClassName="p-5">
      <p className="text-[10px] uppercase tracking-widest text-white/40 mb-2 font-semibold">{label}</p>
      <p className={`text-lg font-bold tracking-tight capitalize ${color}`}>{value}</p>
    </GlassSurface>
  );
}
