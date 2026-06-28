'use client';
import GlassButton from "@/components/ui/GlassButton";
import { getAccount, getTrades } from "@/services/account/queries";
import { useResetAccount, useDeleteAccount } from "@/services/account/mutations";
import { useSignout } from "@/services/auth/mutations";
import { User, Wallet, LogOut, History, RotateCcw, Trash2 } from "lucide-react";
import { GlassCard } from "@/components/ui/GlassCard";
import { TextEffect } from "@/components/primitives/TextEffect";
import { PageEnter, RevealStagger, RevealItem } from "@/components/primitives/Reveal";
import { SlidingNumber } from "@/components/primitives/SlidingNumber";
import { formatCents, dollarsFromCents, timeAgo } from "@/lib/utils";
import { motion } from "motion/react";

type Fill = {
  trade_id: string;
  order_id: string;
  symbol: string;
  side: string;
  qty: number;
  price_cents: number;
  liquidity: string;
  executed_at: string;
};

function ProfileDetail({ label, value }: { label: string, value: string }) {
  return (
    <div className="flex flex-col gap-y-1 mb-4 last:mb-0">
      <span className="text-xs text-white/40 uppercase tracking-wider">{label}</span>
      <span className="text-base text-white/80 break-all">{value}</span>
    </div>
  );
}

export default function Account() {
  const signout = useSignout();
  const { data: account, isLoading } = getAccount();
  const { data: tradesData } = getTrades();
  const resetAccount = useResetAccount();
  const deleteAccount = useDeleteAccount();

  const userName: string = account?.user_name || "Trader";
  const email: string = account?.email || "-";
  const initials = userName
    .split(" ")
    .map((p: string) => p.charAt(0))
    .join("")
    .slice(0, 2)
    .toUpperCase();

  const fills: Fill[] = tradesData?.fills ?? [];

  const handleReset = () => {
    if (!window.confirm("Reset your paper account back to $100,000? This wipes positions, open orders' reservations, and trade history.")) return;
    resetAccount.mutate();
  };

  const handleDelete = () => {
    if (!window.confirm("Permanently delete your account? This removes your profile, positions, orders, trade history, and all strategies. This cannot be undone.")) return;
    deleteAccount.mutate();
  };

  return (
    <PageEnter className="relative min-h-screen w-full flex flex-col items-center justify-start font-mono p-4 pt-32 pb-32 md:p-8 md:pt-32 xl:p-12 xl:pt-32 z-0">
      <RevealStagger className="w-full max-w-5xl flex flex-col" stagger={0.08}>
        <RevealItem className="mb-8 flex justify-between items-end">
          <TextEffect as="h1" preset="fade-in-blur" per="word" className="text-3xl sm:text-4xl font-semibold text-white/90">
            Account
          </TextEffect>
        </RevealItem>

        <RevealStagger className="w-full grid grid-cols-1 lg:grid-cols-2 gap-6 lg:gap-8" stagger={0.08} delay={0.1}>
          <RevealItem>
            <motion.div transition={{ duration: 0.25 }}>
              <GlassCard title="Profile" icon={User}>
                <div className="flex items-center gap-x-6 mb-8">
                  <motion.div
                    initial={{ scale: 0.8, opacity: 0 }}
                    animate={{ scale: 1, opacity: 1 }}
                    transition={{ delay: 0.3, duration: 0.4, ease: [0.22, 1, 0.36, 1] }}
                    className="w-16 h-16 rounded-full bg-white/5 border border-white/10 flex items-center justify-center text-xl text-white/60"
                  >
                    {initials || "-"}
                  </motion.div>
                  <div className="flex flex-col">
                    <span className="text-lg font-medium text-white/90">{isLoading ? "Loading…" : userName}</span>
                    <span className="text-sm text-white/50">Paper Trading Account</span>
                  </div>
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
                  <ProfileDetail label="Email Address" value={email} />
                  <ProfileDetail label="Base Currency" value={`${account?.currency || "USD"} ($)`} />
                  <ProfileDetail label="Equity" value={`$${formatCents(account?.equity_cents)}`} />
                  <ProfileDetail label="Buying Power" value={`$${formatCents(account?.buying_power_cents)}`} />
                </div>
              </GlassCard>
            </motion.div>
          </RevealItem>

          <RevealItem>
            <motion.div transition={{ duration: 0.25 }}>
              <GlassCard title="Paper Funds" icon={Wallet} flexCol>
                <div className="mb-6">
                  <span className="text-xs text-white/40 uppercase tracking-wider">Cash Balance</span>
                  <p className="text-3xl text-white/90 font-light tracking-tight flex items-center mt-1">
                    <span className="text-white/50 mr-1">$</span>
                    <SlidingNumber value={+dollarsFromCents(account?.cash_balance_cents).toFixed(2)} />
                  </p>
                  <p className="text-xs text-white/40 mt-1">
                    ${formatCents(account?.reserved_cash_cents)} reserved for open orders
                  </p>
                </div>

                <div className="flex flex-col gap-y-3">
                  <p className="text-xs text-white/40 leading-relaxed">
                    Every paper account starts with a fixed $100,000, the same
                    balance for everyone, so performance is comparable and
                    fills stay meaningful.
                  </p>

                  <div className="flex items-start justify-between pt-4 mt-2 border-t border-white/10">
                    <div className="flex flex-col gap-y-1 max-w-[65%]">
                      <span className="text-sm font-medium text-white/80">Reset Account</span>
                      <span className="text-xs text-white/40 leading-snug">Restore the $100,000 starting balance and wipe positions and trade history.</span>
                    </div>
                    <GlassButton
                      text={resetAccount.isPending ? "Resetting…" : "Reset"}
                      icon={<RotateCcw size={14} />}
                      onClick={handleReset}
                      disabled={resetAccount.isPending}
                      className="border-red-500/30 text-red-400"
                    />
                  </div>
                </div>
              </GlassCard>
            </motion.div>
          </RevealItem>

          <RevealItem>
            <motion.div transition={{ duration: 0.25 }}>
              <GlassCard title="Recent Trades" icon={History}>
                {fills.length > 0 ? (
                  <div className="flex flex-col divide-y divide-white/5">
                    {fills.slice(0, 6).map((f) => (
                      <div key={`${f.trade_id}-${f.order_id}`} className="flex items-center justify-between py-3 first:pt-0 last:pb-0">
                        <div className="flex items-center gap-3">
                          <span className={`text-xs font-bold uppercase px-2 py-0.5 rounded-md border ${f.side === "buy"
                            ? "text-emerald-400 border-emerald-500/20 bg-emerald-500/10"
                            : "text-red-400 border-red-500/20 bg-red-500/10"}`}>
                            {f.side}
                          </span>
                          <span className="text-sm text-white/80 font-semibold">{f.symbol}</span>
                          <span className="text-xs text-white/40">{f.qty} × ${formatCents(f.price_cents)}</span>
                        </div>
                        <span className="text-xs text-white/40">{timeAgo(f.executed_at)}</span>
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="text-sm text-white/40 py-6 text-center">No trades yet, fills appear here as your orders execute.</p>
                )}
              </GlassCard>
            </motion.div>
          </RevealItem>

          <RevealItem>
            <motion.div transition={{ duration: 0.25 }}>
              <GlassCard title="Account Actions" icon={LogOut}>
                <div className="flex flex-col gap-y-6 h-full justify-between">
                  <div className="flex items-start justify-between">
                    <div className="flex flex-col gap-y-1 max-w-[65%]">
                      <span className="text-sm font-medium text-white/80 shrink-0">Sign Out</span>
                      <span className="text-xs text-white/40 leading-snug">Securely sign out of your account on this device.</span>
                    </div>
                    <GlassButton onClick={() => signout.mutate()} text="Sign out" icon={<LogOut size={14} />} />
                  </div>

                  <div className="flex items-start justify-between pt-4 mt-2 border-t border-white/10">
                    <div className="flex flex-col gap-y-1 max-w-[65%]">
                      <span className="text-sm font-medium text-white/80 shrink-0">Delete Account</span>
                      <span className="text-xs text-white/40 leading-snug">Permanently delete your account, including all positions, trades, and strategies.</span>
                    </div>
                    <GlassButton
                      text={deleteAccount.isPending ? "Deleting…" : "Delete"}
                      icon={<Trash2 size={14} />}
                      onClick={handleDelete}
                      disabled={deleteAccount.isPending}
                      className="border-red-500/30 text-red-400"
                    />
                  </div>
                </div>
              </GlassCard>
            </motion.div>
          </RevealItem>
        </RevealStagger>
      </RevealStagger>
    </PageEnter>
  );
}
