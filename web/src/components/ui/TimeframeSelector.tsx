"use client";

import { TIMEFRAMES, Timeframe } from "@/lib/marketData";
import { motion } from "motion/react";
import { useId } from "react";

interface TimeframeSelectorProps {
  value: Timeframe;
  onChange: (tf: Timeframe) => void;
  size?: "sm" | "md";
    options?: readonly Timeframe[];
  className?: string;
    stopPropagation?: boolean;
}

export default function TimeframeSelector({
  value,
  onChange,
  size = "md",
  options = TIMEFRAMES,
  className = "",
  stopPropagation = false,
}: TimeframeSelectorProps) {
  const padding = size === "sm" ? "px-2 py-1 text-[10px]" : "px-3 py-1.5 text-xs";
  const layoutId = useId();

  return (
    <div
      className={`inline-flex items-center gap-0.5 rounded-full border border-white/10 bg-black/30 backdrop-blur-md p-0.5 ${className}`}
      onClick={stopPropagation ? (e) => e.preventDefault() : undefined}
    >
      {options.map((tf) => {
        const active = tf === value;
        return (
          <motion.button
            key={tf}
            type="button"
            whileTap={{ scale: 0.92 }}
            onClick={(e) => {
              if (stopPropagation) {
                e.preventDefault();
                e.stopPropagation();
              }
              onChange(tf);
            }}
            className={`relative ${padding} font-mono font-semibold tracking-wider rounded-full transition-colors ${
              active ? "text-white" : "text-white/40 hover:text-white/80"
            }`}
          >
            {active && (
              <motion.div
                layoutId={`tf-pill-${layoutId}`}
                className="absolute inset-0 bg-white/15 rounded-full shadow-[inset_0_0_0_1px_rgba(255,255,255,0.15)]"
                transition={{ type: "spring", stiffness: 400, damping: 32 }}
              />
            )}
            <span className="relative z-10">{tf}</span>
          </motion.button>
        );
      })}
    </div>
  );
}
