"use client";

import { useEffect, useState } from "react";
import { motion } from "motion/react";
import {
  isUSMarketOpen,
  nextMarketOpen,
  formatCountdown,
} from "@/lib/marketHours";

function useNow(pollMs = 30_000): Date | null {
  const [now, setNow] = useState<Date | null>(null);
  useEffect(() => {
    const tick = () => setNow(new Date());
    tick();
    const id = setInterval(tick, pollMs);
    return () => clearInterval(id);
  }, [pollMs]);
  return now;
}

export function useMarketOpen(pollMs = 30_000): boolean | null {
  const now = useNow(pollMs);
  return now ? isUSMarketOpen(now) : null;
}

interface MarketStatusBadgeProps {
    variant?: "badge" | "compact";
  className?: string;
}

export default function MarketStatusBadge({
  variant = "badge",
  className = "",
}: MarketStatusBadgeProps) {
  const now = useNow();
  const open = now ? isUSMarketOpen(now) : null;
  const isOpen = open === true;

  let countdown = "";
  let localOpen = "";
  if (now && !isOpen) {
    const next = nextMarketOpen(now);
    countdown = formatCountdown(next.getTime() - now.getTime());
    localOpen = next.toLocaleTimeString([], { hour: "numeric", minute: "2-digit" });
  }
  const closedTitle = `Opens 9:30 AM ET · ${localOpen} your time${
    countdown ? ` · in ${countdown}` : ""
  }`;

  const dotColor =
    open === null ? "bg-white/40" : isOpen ? "bg-emerald-500" : "bg-amber-400/80";

  const dot =
    isOpen ? (
      <motion.span
        className="flex h-2 w-2 rounded-full bg-emerald-500"
        animate={{
          scale: [1, 1.4, 1],
          boxShadow: [
            "0 0 8px rgba(16,185,129,0.8)",
            "0 0 16px rgba(16,185,129,0.95)",
            "0 0 8px rgba(16,185,129,0.8)",
          ],
        }}
        transition={{ duration: 2, repeat: Infinity, ease: "easeInOut" }}
      />
    ) : (
      <span className={`flex h-2 w-2 rounded-full ${dotColor}`} />
    );

  if (variant === "compact") {
    const compactLabel =
      open === null ? "Markets" : isOpen ? "Live" : countdown ? `Opens ${countdown}` : "Closed";
    return (
      <div
        className={`flex items-center gap-1.5 px-2.5 py-1 rounded-full border border-white/10 bg-white/5 text-[10px] font-semibold uppercase tracking-widest text-white/70 ${className}`}
        title={isOpen ? "US market open · 9:30 AM - 4:00 PM ET" : closedTitle}
      >
        <span className={`flex h-1.5 w-1.5 rounded-full ${dotColor}`} />
        <span>{compactLabel}</span>
      </div>
    );
  }

  const label =
    open === null
      ? "Checking market status…"
      : isOpen
        ? "Live · market open"
        : `Market closed · opens in ${countdown} (${localOpen} local)`;

  return (
    <span className={`flex items-center gap-2 ${className}`}>
      {dot}
      {label}
    </span>
  );
}
