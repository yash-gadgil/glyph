'use client';

import { getPortfolio, getHoldings } from "@/services/portfolio/queries";
import { TrendingUp, TrendingDown, Clock, Activity, Briefcase, PieChart } from "lucide-react";
import GlassSurface from "@/components/primitives/GlassSurface";
import { TextEffect } from "@/components/primitives/TextEffect";
import { PageEnter, RevealStagger, RevealItem } from "@/components/primitives/Reveal";
import { SlidingNumber } from "@/components/primitives/SlidingNumber";
import PortfolioValueChart from "@/components/portfolio/PortfolioValueChart";
import { formatCents, dollarsFromCents } from "@/lib/utils";
import { motion } from "motion/react";

type Holding = {
  symbol: string;
  qty: number;
  avg_price_cents: number;
  cost_basis_cents: number;
  last_price_cents: number;
  market_value_cents: number;
  unrealized_pnl_cents: number;
  realized_pnl_cents: number;
};

export default function Portfolio() {
  const { data: portfolioData, isLoading: isPortfolioLoading, error: portfolioError } = getPortfolio();
  const { data: holdingsData, isLoading: isHoldingsLoading, error: holdingsError } = getHoldings();

  const isLoading = isPortfolioLoading || isHoldingsLoading;
  const hasError = portfolioError || holdingsError;

  if (hasError) {
    return (
      <div className="flex min-h-screen w-full items-center justify-center bg-neutral-950 p-8">
        <div className="rounded-xl border border-red-500/20 bg-red-500/10 p-6 text-red-400 font-mono text-sm max-w-lg shadow-xl shadow-red-900/10">
          <p className="mb-2 font-bold text-red-500 flex items-center gap-2">
            <Activity size={18} /> Error Retrieving Portfolio
          </p>
          <p className="mb-1 opacity-80">{portfolioError instanceof Error ? portfolioError.message : ''}</p>
          <p className="opacity-80">{holdingsError instanceof Error ? holdingsError.message : ''}</p>
        </div>
      </div>
    );
  }

  const portfolio = portfolioData || {};
  const holdings: Holding[] = holdingsData?.holdings || [];

  const cashCents: number = portfolio.cash_balance_cents ?? 0;
  const reservedCents: number = portfolio.reserved_cash_cents ?? 0;
  const buyingPowerCents: number = portfolio.buying_power_cents ?? (cashCents - reservedCents);
  const marketValueCents: number = holdingsData?.total_market_value_cents ?? 0;
  const unrealizedCents: number = holdingsData?.total_unrealized_pnl_cents ?? 0;
  const realizedCents: number = holdingsData?.total_realized_pnl_cents ?? 0;
  const costBasisCents: number = holdingsData?.total_cost_basis_cents ?? 0;

  const totalValueCents = cashCents + marketValueCents;
  const pnlIsPositive = unrealizedCents >= 0;
  const pnlPct = costBasisCents > 0 ? (unrealizedCents / costBasisCents) * 100 : 0;

  const openHoldings = holdings.filter((h) => h.qty !== 0);

  return (
    <PageEnter className="min-h-screen w-full text-neutral-300 font-mono p-4 pt-32 md:p-8 md:pt-32 xl:p-12 xl:pt-32 overflow-y-auto">
      <RevealStagger className="mx-auto max-w-7xl space-y-8 pb-32" stagger={0.08}>

        <RevealItem>
          <header className="flex flex-col sm:flex-row sm:items-end justify-between gap-4">
            <div className="space-y-1">
              <TextEffect
                as="h1"
                preset="fade-in-blur"
                per="word"
                className="text-2xl md:text-3xl font-medium text-neutral-100 tracking-tight"
              >
                Portfolio Overview
              </TextEffect>
            </div>
          </header>
        </RevealItem>

        <RevealItem>
          <section className="grid grid-cols-1 lg:grid-cols-4 gap-6">
            <GlassSurface borderRadius={16} order="start" alignItems="stretch" flexDirection="col" className="col-span-1 lg:col-span-3" innerClassName="p-6">
              <div className="mb-6 flex flex-col sm:flex-row sm:justify-between sm:items-start gap-4">
                <div>
                  <p className="text-sm text-neutral-400 font-medium mb-1">Total Account Value</p>
                  <div className="flex flex-col sm:flex-row sm:items-baseline gap-3">
                    <h2 className="text-4xl md:text-5xl font-light text-white tracking-tighter flex items-center">
                      <span className="text-white/70 mr-1">$</span>
                      <SlidingNumber value={+dollarsFromCents(totalValueCents).toFixed(2)} />
                    </h2>
                    <div className={`flex items-center text-sm font-medium px-2.5 py-1 rounded-md bg-opacity-10 backdrop-blur-sm border ${pnlIsPositive ? 'text-emerald-400 bg-emerald-500/10 border-emerald-500/20' : 'text-red-400 bg-red-500/10 border-red-500/20'}`}>
                      {pnlIsPositive ? <TrendingUp size={14} className="mr-1.5" /> : <TrendingDown size={14} className="mr-1.5" />}
                      {pnlIsPositive ? '+' : ''}{formatCents(unrealizedCents)} ({pnlIsPositive ? '+' : ''}{pnlPct.toFixed(2)}%) unrealized
                    </div>
                  </div>
                </div>
              </div>

              <div className="mt-6">
                <p className="text-xs uppercase tracking-wider text-neutral-500 mb-4">Allocation</p>
                {openHoldings.length > 0 || cashCents > 0 ? (
                  <div className="space-y-3">
                    <AllocationBar
                      label="Cash"
                      valueCents={cashCents}
                      totalCents={totalValueCents}
                      muted
                      index={0}
                    />
                    {openHoldings
                      .slice()
                      .sort((a, b) => b.market_value_cents - a.market_value_cents)
                      .map((h, i) => (
                        <AllocationBar
                          key={h.symbol}
                          label={h.symbol}
                          valueCents={h.market_value_cents}
                          totalCents={totalValueCents}
                          positive={h.unrealized_pnl_cents >= 0}
                          index={i + 1}
                        />
                      ))}
                  </div>
                ) : (
                  <div className="py-12 text-center text-neutral-500 text-sm">
                    {isLoading ? 'Loading…' : 'No positions yet, place your first order to put your $100,000 paper balance to work.'}
                  </div>
                )}
              </div>
            </GlassSurface>

            <div className="col-span-1 flex flex-col gap-4">
              <motion.div transition={{ duration: 0.25 }} className="flex-1">
                <GlassSurface borderRadius={16} order="center" alignItems="stretch" flexDirection="col" className="h-full" innerClassName="p-6 justify-center">
                  <div className="flex items-center gap-3 mb-4">
                    <p className="text-xl font-medium text-neutral-400">Cash Balance</p>
                  </div>
                  <p className="text-3xl text-neutral-100 font-light tracking-tight flex items-center">
                    <span className="text-neutral-300/80 mr-1">$</span>
                    <SlidingNumber value={+dollarsFromCents(cashCents).toFixed(2)} />
                  </p>
                </GlassSurface>
              </motion.div>

              <motion.div transition={{ duration: 0.25 }} className="flex-1">
                <GlassSurface borderRadius={16} order="center" alignItems="stretch" flexDirection="col" className="h-full" innerClassName="p-6 justify-center">
                  <div className="flex items-center gap-3 mb-4">
                    <p className="text-xl font-medium text-neutral-400">Market Value</p>
                  </div>
                  <p className="text-3xl text-neutral-100 font-light tracking-tight flex items-center">
                    <span className="text-neutral-300/80 mr-1">$</span>
                    <SlidingNumber value={+dollarsFromCents(marketValueCents).toFixed(2)} />
                  </p>
                </GlassSurface>
              </motion.div>
            </div>
          </section>
        </RevealItem>

        <RevealItem>
          <section className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <StatCard
              icon={<Briefcase size={16} />}
              label="Reserved Cash"
              value={`$${formatCents(reservedCents)}`}
              index={0}
            />
            <StatCard
              icon={<Clock size={16} />}
              label="Buying Power"
              value={`$${formatCents(buyingPowerCents)}`}
              index={1}
            />
            <StatCard
              icon={<Activity size={16} />}
              label="Realized PnL"
              value={`${realizedCents >= 0 ? '+' : '-'}$${formatCents(Math.abs(realizedCents))}`}
              index={2}
            />
            <StatCard
              icon={<PieChart size={16} />}
              label="Currency"
              value={(portfolio.currency || 'USD').toUpperCase()}
              index={3}
            />
          </section>
        </RevealItem>

        <RevealItem>
          <PortfolioValueChart />
        </RevealItem>

        <RevealItem>
          <section>
            <div className="mb-6 flex items-center justify-between">
              <h3 className="text-xl font-medium text-neutral-200 tracking-tight">Current Positions</h3>
              <div className="flex gap-2">
                <span className="px-3 py-1 rounded-full bg-neutral-900 border border-neutral-800 text-xs text-neutral-400">
                  {openHoldings.length} {openHoldings.length === 1 ? "Position" : "Positions"}
                </span>
              </div>
            </div>

            <GlassSurface borderRadius={16} order="start" alignItems="stretch" flexDirection="col">
              <div className="overflow-x-auto w-full">
                <table className="w-full text-left text-sm whitespace-nowrap">
                  <thead className="bg-neutral-950 border-b border-neutral-800 text-neutral-500">
                    <tr>
                      <th className="px-6 py-5 font-semibold text-xs tracking-wider uppercase">Asset</th>
                      <th className="px-6 py-5 font-semibold text-xs tracking-wider uppercase text-right">Shares</th>
                      <th className="px-6 py-5 font-semibold text-xs tracking-wider uppercase text-right">Avg Price</th>
                      <th className="px-6 py-5 font-semibold text-xs tracking-wider uppercase text-right">Last Price</th>
                      <th className="px-6 py-5 font-semibold text-xs tracking-wider uppercase text-right">Market Value</th>
                      <th className="px-6 py-5 font-semibold text-xs tracking-wider uppercase text-right leading-tight">Unrealized<br />PnL</th>
                      <th className="px-6 py-5 font-semibold text-xs tracking-wider uppercase text-right leading-tight">Realized<br />PnL</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-neutral-800/50">
                    {openHoldings.length > 0 ? (
                      openHoldings.map((h, idx) => {
                        const unrealized = h.unrealized_pnl_cents;
                        const realized = h.realized_pnl_cents;
                        return (
                          <motion.tr
                            key={h.symbol}
                            initial={{ opacity: 0, x: -16 }}
                            animate={{ opacity: 1, x: 0 }}
                            transition={{ delay: 0.35 + idx * 0.05, duration: 0.4, ease: [0.22, 1, 0.36, 1] }}
                            className="group hover:bg-neutral-800/40 transition-colors cursor-default"
                          >
                            <td className="px-6 py-4">
                              <div className="flex items-center gap-4">
                                <div className="h-10 w-10 rounded-full bg-neutral-800 flex items-center justify-center text-sm font-bold text-neutral-300 border border-neutral-700 shadow-inner group-hover:border-neutral-500 transition-colors">
                                  {h.symbol.charAt(0)}
                                </div>
                                <div>
                                  <p className="font-semibold text-neutral-200">{h.symbol}</p>
                                  <p className="text-xs text-neutral-500 mt-0.5">Equity</p>
                                </div>
                              </div>
                            </td>
                            <td className="px-6 py-4 text-right font-medium text-neutral-300 align-middle">{h.qty}</td>
                            <td className="px-6 py-4 text-right text-neutral-400 align-middle">${formatCents(h.avg_price_cents)}</td>
                            <td className="px-6 py-4 text-right text-neutral-300 align-middle">${formatCents(h.last_price_cents)}</td>
                            <td className="px-6 py-4 text-right text-neutral-300 align-middle">${formatCents(h.market_value_cents)}</td>
                            <td className={`px-6 py-4 text-right font-medium align-middle ${unrealized >= 0 ? 'text-emerald-400' : 'text-red-400'}`}>
                              {unrealized >= 0 ? '+' : '-'}${formatCents(Math.abs(unrealized))}
                            </td>
                            <td className={`px-6 py-4 text-right font-medium align-middle ${realized >= 0 ? 'text-emerald-400' : 'text-red-400'}`}>
                              {realized >= 0 ? '+' : '-'}${formatCents(Math.abs(realized))}
                            </td>
                          </motion.tr>
                        );
                      })
                    ) : (
                      <tr>
                        <td colSpan={7} className="px-6 py-16 text-center text-neutral-500 text-sm">
                          {isLoading ? 'Loading positions…' : 'No open positions, fills land here as your orders execute.'}
                        </td>
                      </tr>
                    )}
                  </tbody>
                </table>
              </div>
            </GlassSurface>
          </section>
        </RevealItem>

      </RevealStagger>
    </PageEnter>
  );
}

function AllocationBar({ label, valueCents, totalCents, positive = true, muted = false, index = 0 }: {
  label: string;
  valueCents: number;
  totalCents: number;
  positive?: boolean;
  muted?: boolean;
  index?: number;
}) {
  const pct = totalCents > 0 ? (valueCents / totalCents) * 100 : 0;
  const barColor = muted
    ? 'from-neutral-600 to-neutral-500'
    : positive
      ? 'from-emerald-600 to-emerald-400'
      : 'from-red-600 to-red-400';

  return (
    <div className="flex items-center gap-4">
      <span className="w-16 shrink-0 text-xs font-semibold text-neutral-300">{label}</span>
      <div className="flex-1 h-2.5 rounded-full bg-neutral-950 overflow-hidden border border-neutral-800">
        <motion.div
          className={`h-full bg-linear-to-r ${barColor} rounded-full`}
          initial={{ width: 0 }}
          animate={{ width: `${pct}%` }}
          transition={{ duration: 0.8, ease: [0.22, 1, 0.36, 1], delay: 0.2 + index * 0.05 }}
        />
      </div>
      <span className="w-24 sm:w-28 shrink-0 text-right text-xs text-neutral-400 tabular-nums">${formatCents(valueCents)}</span>
      <span className="w-12 shrink-0 text-right text-xs text-neutral-500">{pct.toFixed(1)}%</span>
    </div>
  );
}

function StatCard({ icon, label, value, className = "", index = 0 }: { icon: React.ReactNode, label: string, value: string, className?: string, index?: number }) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 14 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: 0.3 + index * 0.06, duration: 0.4, ease: [0.22, 1, 0.36, 1] }}
    >
      <GlassSurface borderRadius={16} order="start" alignItems="stretch" flexDirection="col" className={`transition-colors hover:border-white/15 ${className}`} innerClassName="p-5">
        <div className="flex items-center gap-2 mb-4 text-neutral-400 bg-neutral-950/50 w-fit px-3 py-1.5 rounded-lg border border-neutral-800 shadow-inner">
          <span className="text-neutral-500">{icon}</span>
          <p className="text-xs font-semibold tracking-wide uppercase">{label}</p>
        </div>
        <p className="text-2xl text-neutral-100 font-light tracking-tight wrap-break-word">{value}</p>
      </GlassSurface>
    </motion.div>
  )
}
