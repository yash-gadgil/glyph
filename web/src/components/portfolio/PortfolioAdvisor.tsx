'use client';

import { useEffect } from "react";
import { Sparkles, RefreshCw, AlertTriangle } from "lucide-react";
import { motion } from "motion/react";
import GlassSurface from "@/components/primitives/GlassSurface";
import PixelHover from "@/components/ui/PixelHover";
import { useAnalyzePortfolio } from "@/services/advisor/stream";

export default function PortfolioAdvisor() {
  const { text, status, start } = useAnalyzePortfolio();
  const isStreaming = status === "streaming";

  useEffect(() => {
    start();
  }, [start]);

  return (
    <section>
      <div className="mb-6 flex items-center justify-between gap-4">
        <h3 className="text-xl font-medium text-neutral-200 tracking-tight flex items-center gap-2">
          <Sparkles size={18} style={{ color: "#5600a2" }} />
          AI Portfolio Analysis
        </h3>
        <PixelHover
          gap={3}
          speed={40}
          colors="#a06cd5,#7d34c4,#5600a2"
          active={isStreaming}
          className="group rounded-lg border border-white/10 bg-white/5 hover:bg-white/10 hover:border-[#5600a2]/60 transition-colors"
        >
          <button
            onClick={start}
            disabled={isStreaming}
            className="flex items-center gap-2 px-4 py-2 rounded-lg text-white bg-transparent text-sm font-medium transition-colors group-hover:text-[#c9a6f0] disabled:cursor-not-allowed"
          >
            <RefreshCw size={14} className={isStreaming ? "animate-spin" : ""} />
            {status === "idle" ? "Generate analysis" : isStreaming ? "Analyzing…" : "Regenerate"}
          </button>
        </PixelHover>
      </div>

      <GlassSurface borderRadius={16} order="start" alignItems="stretch" flexDirection="col" innerClassName="p-6">
        {status === "error" ? (
          <div className="flex items-start gap-3 text-red-400 text-sm">
            <AlertTriangle size={18} className="mt-0.5 shrink-0" />
            <p>The analysis could not be generated right now. Try again in a moment.</p>
          </div>
        ) : status === "idle" ? (
          <div className="py-10 text-center text-neutral-500 text-sm">
            Generate a written review of your current allocation, risks, and the kinds of strategies that suit this book.
          </div>
        ) : (
          <div className="text-sm leading-relaxed text-neutral-300 whitespace-pre-wrap">
            {isStreaming && !text && (
              <span className="text-neutral-500">Reading your positions…</span>
            )}
            {text}
            {isStreaming && (
              <motion.span
                className="inline-block w-2 h-4 ml-0.5 -mb-0.5"
                style={{ backgroundColor: "#5600a2" }}
                animate={{ opacity: [1, 0.2, 1] }}
                transition={{ duration: 1, repeat: Infinity }}
              />
            )}
          </div>
        )}

        {status === "done" && (
          <p className="mt-6 pt-4 border-t border-neutral-800/60 text-xs text-neutral-600">
            Simulated account, educational use only. Not financial advice.
          </p>
        )}
      </GlassSurface>
    </section>
  );
}
