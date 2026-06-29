'use client';

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import GlassSurface from "@/components/primitives/GlassSurface";
import {
  CustomStrategy,
  RuleGroup,
  formatGroup,
  loadCustomStrategies,
  createCustomStrategy,
  deleteCustomStrategyById,
} from "@/lib/strategies";
import {
  Activity,
  AlertTriangle,
  Bot,
  ChevronDown,
  ChevronRight,
  CircleDot,
  Cpu,
  FlaskConical,
  Layers,
  Plus,
  ShieldCheck,
  Sparkles,
  Square,
  Target,
  TrendingDown,
  TrendingUp,
  Trash2,
  Wand2,
  Zap,
} from "lucide-react";
import PixelHover from "@/components/ui/PixelHover";
import ModalPortal from "@/components/ui/ModalPortal";
import SymbolCombobox from "@/components/ui/SymbolCombobox";
import { TextEffect } from "@/components/primitives/TextEffect";
import { PageEnter, RevealStagger, RevealItem } from "@/components/primitives/Reveal";
import { motion, AnimatePresence } from "motion/react";
import {
  generateStrategy,
  resumeStrategyGeneration,
  isStrategyGenerationRunning,
  type GeneratedStrategy,
  type BacktestSummary,
} from "@/services/advisor/strategy";
import { useDeployments, type Deployment } from "@/services/strategies/queries";
import { useDeployStrategy, useStopDeployment, useDeleteDeployment } from "@/services/strategies/mutations";
import { formatCents } from "@/lib/utils";
import MarketStatusBadge from "@/components/ui/MarketStatusBadge";
import BacktestModal from "@/components/strategies/BacktestModal";


type RiskLevel = "low" | "medium" | "high";

interface StrategyPreset {
  key: string;
  name: string;
  tagline: string;
  description: string;
  risk: RiskLevel;
  icon: React.ReactNode;
  tags: string[];
  entry: RuleGroup;
  exit: RuleGroup;
  stopLossPct: number;
  takeProfitPct: number;
}

const rule = (
  lhsKind: string,
  lhsParams: Record<string, number>,
  op: string,
  rhs: { value?: number; kind?: string; params?: Record<string, number> }
) => ({
  id: `${lhsKind}-${op}-${Math.random().toString(36).slice(2, 8)}`,
  lhs: { kind: lhsKind, params: lhsParams },
  op,
  rhs:
    rhs.value !== undefined
      ? { kind: "value", value: rhs.value }
      : { kind: "indicator", indicator: { kind: rhs.kind!, params: rhs.params ?? {} } },
}) as RuleGroup["rules"][number];

const PRESETS: StrategyPreset[] = [
  {
    key: "rsi_dip_buyer",
    name: "RSI Dip Buyer",
    tagline: "Buy oversold dips, sell the recovery",
    description:
      "Enters when RSI(14) drops below 30, a classic oversold signal, and exits when momentum recovers above 60. ATR-agnostic, works on liquid large-caps. Stop-loss caps downside if the dip keeps dipping.",
    risk: "medium",
    icon: <TrendingDown size={22} />,
    tags: ["Mean Reversion", "RSI", "Intraday"],
    entry: { combinator: "AND", rules: [rule("rsi", { period: 14 }, "<", { value: 30 })] },
    exit: { combinator: "OR", rules: [rule("rsi", { period: 14 }, ">", { value: 60 })] },
    stopLossPct: 2,
    takeProfitPct: 3,
  },
  {
    key: "momentum_breakout",
    name: "Momentum Breakout",
    tagline: "Ride moves through the upper Bollinger band",
    description:
      "Enters when price crosses above the upper Bollinger band (20, 2σ), volatility expansion in your favour, and exits when price falls back through the middle band. Best in trending tape.",
    risk: "high",
    icon: <TrendingUp size={22} />,
    tags: ["Momentum", "Breakout", "Bollinger"],
    entry: {
      combinator: "AND",
      rules: [rule("price", {}, "crosses_above", { kind: "bbands_upper", params: { period: 20, stddev: 2 } })],
    },
    exit: {
      combinator: "OR",
      rules: [rule("price", {}, "crosses_below", { kind: "bbands_middle", params: { period: 20 } })],
    },
    stopLossPct: 2,
    takeProfitPct: 5,
  },
  {
    key: "macd_trend_rider",
    name: "MACD Trend Rider",
    tagline: "Follow MACD line / signal crossovers",
    description:
      "Enters on the MACD line crossing above its signal line (12/26/9) and exits on the cross back below. The classic trend-following crossover, with a stop-loss for failed signals.",
    risk: "medium",
    icon: <Target size={22} />,
    tags: ["Trend Following", "MACD", "Crossover"],
    entry: {
      combinator: "AND",
      rules: [
        rule("macd_line", { fast: 12, slow: 26 }, "crosses_above", {
          kind: "macd_signal",
          params: { fast: 12, slow: 26, signal: 9 },
        }),
      ],
    },
    exit: {
      combinator: "OR",
      rules: [
        rule("macd_line", { fast: 12, slow: 26 }, "crosses_below", {
          kind: "macd_signal",
          params: { fast: 12, slow: 26, signal: 9 },
        }),
      ],
    },
    stopLossPct: 3,
    takeProfitPct: 6,
  },
  {
    key: "band_reversion",
    name: "Band Reversion",
    tagline: "Buy washouts below the lower band, exit at VWAP",
    description:
      "Enters when price breaks below the lower Bollinger band while RSI confirms oversold, and exits when price reclaims the session VWAP. Tight stop, modest target, low risk by construction.",
    risk: "low",
    icon: <Activity size={22} />,
    tags: ["VWAP", "Mean Reversion", "Bollinger"],
    entry: {
      combinator: "AND",
      rules: [
        rule("price", {}, "<", { kind: "bbands_lower", params: { period: 20, stddev: 2 } }),
        rule("rsi", { period: 14 }, "<", { value: 35 }),
      ],
    },
    exit: {
      combinator: "OR",
      rules: [rule("price", {}, "crosses_above", { kind: "vwap", params: {} })],
    },
    stopLossPct: 1.5,
    takeProfitPct: 2.5,
  },
];


function riskColor(r: RiskLevel) {
  return r === "low"
    ? "text-emerald-400 bg-emerald-500/10 border-emerald-500/20"
    : r === "medium"
      ? "text-amber-400 bg-amber-500/10 border-amber-500/20"
      : "text-red-400 bg-red-500/10 border-red-500/20";
}

function statusColor(s: Deployment["status"]) {
  return s === "running"
    ? "text-emerald-400 bg-emerald-500/10 border-emerald-500/20"
    : "text-white/40 bg-white/5 border-white/10";
}


function DeployModal({
  name,
  tagline,
  rulesPreview,
  busy,
  onClose,
  onDeploy,
  apiError,
}: {
  name: string;
  tagline: string;
  rulesPreview: { entry: string; exit: string };
  busy: boolean;
  onClose: () => void;
  onDeploy: (symbol: string, positionSizeCents: number) => void;
  apiError: string;
}) {
  const [symbol, setSymbol] = useState("AAPL");
  const [posSize, setPosSize] = useState("5000");
  const [error, setError] = useState("");

  function handleDeploy() {
    if (!symbol.trim()) { setError("Symbol is required"); return; }
    const size = parseFloat(posSize);
    if (!size || size <= 0) { setError("Position size must be a positive number"); return; }
    setError("");
    onDeploy(symbol.toUpperCase().trim(), Math.round(size * 100));
  }

  const shownError = error || apiError;

  return (
    <ModalPortal>
      <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
        <div className="w-full max-w-lg rounded-2xl border border-white/10 bg-neutral-950/95 shadow-2xl overflow-hidden">
          <div className="px-6 py-5 border-b border-white/10 flex items-start justify-between gap-4">
            <div>
              <p className="text-xs text-white/40 uppercase tracking-widest mb-1 font-semibold">Deploy Strategy</p>
              <h2 className="text-xl font-bold text-white">{name}</h2>
              <p className="text-xs text-white/50 mt-1">{tagline}</p>
            </div>
            <button
              onClick={onClose}
              className="p-1.5 rounded-lg text-white/40 hover:text-white hover:bg-white/10 transition-colors mt-0.5 shrink-0"
            >
              ✕
            </button>
          </div>

          <div className="px-6 py-5 space-y-5 max-h-[65vh] overflow-y-auto">
            <div className="rounded-xl border border-white/10 bg-white/5 p-4 space-y-3">
              <p className="text-[10px] font-semibold text-white/40 uppercase tracking-widest">Strategy Rules</p>
              <div className="space-y-2 text-xs">
                <div className="flex items-start gap-2">
                  <span className="px-1.5 py-0.5 rounded text-[9px] font-bold uppercase tracking-wider text-emerald-400 bg-emerald-500/10 border border-emerald-500/20 shrink-0">Buy</span>
                  <span className="text-white/70 wrap-break-word leading-relaxed">{rulesPreview.entry || "-"}</span>
                </div>
                <div className="flex items-start gap-2">
                  <span className="px-1.5 py-0.5 rounded text-[9px] font-bold uppercase tracking-wider text-red-400 bg-red-500/10 border border-red-500/20 shrink-0">Sell</span>
                  <span className="text-white/70 wrap-break-word leading-relaxed">{rulesPreview.exit || "-"}</span>
                </div>
              </div>
            </div>

            <div className="space-y-1.5">
              <label className="text-xs font-semibold text-white/50 uppercase tracking-wider">Symbol</label>
              <SymbolCombobox
                value={symbol}
                onChange={(v) => setSymbol(v.toUpperCase())}
                placeholder="AAPL"
                inputClassName="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-sm text-white placeholder:text-white/30 focus:outline-none focus:border-emerald-500/50 focus:ring-1 focus:ring-emerald-500/30 transition-all"
              />
            </div>

            <div className="space-y-1.5">
              <label className="text-xs font-semibold text-white/50 uppercase tracking-wider">Position Size ($)</label>
              <input
                type="number"
                value={posSize}
                onChange={(e) => setPosSize(e.target.value)}
                placeholder="5000"
                className="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-sm text-white placeholder:text-white/30 focus:outline-none focus:border-emerald-500/50 focus:ring-1 focus:ring-emerald-500/30 transition-all"
              />
              <p className="text-[10px] text-white/30 leading-relaxed">
                Cash committed per entry. The engine checks rules once a minute
                during market hours and places real paper orders.
              </p>
            </div>

            {shownError && (
              <div className="p-3 bg-red-500/10 border border-red-500/20 rounded-xl text-red-400 text-xs font-medium flex items-center gap-2">
                <AlertTriangle size={13} /> {shownError}
              </div>
            )}
          </div>

          <div className="px-6 py-4 border-t border-white/10 flex items-center gap-3">
            <button
              onClick={onClose}
              className="flex-1 py-3 rounded-xl text-sm font-semibold uppercase tracking-wider text-white/50 border border-white/10 hover:bg-white/5 transition-all"
            >
              Cancel
            </button>
            <PixelHover
              variant="emerald"
              className="group flex-1 rounded-xl border border-white/10 bg-white/5 hover:bg-white/10 hover:border-emerald-500/40 transition-all active:scale-[0.98]"
            >
              <button
                onClick={handleDeploy}
                disabled={busy}
                className="w-full py-3 rounded-xl text-sm font-bold uppercase tracking-wider text-white bg-transparent group-hover:text-emerald-300 transition-colors disabled:opacity-50"
              >
                {busy ? "Deploying…" : "Deploy"}
              </button>
            </PixelHover>
          </div>
        </div>
      </div>
    </ModalPortal>
  );
}


function GenerateModal({
  busy,
  onClose,
  onGenerate,
}: {
  busy: boolean;
  onClose: () => void;
  onGenerate: (symbol: string) => void;
}) {
  const [symbol, setSymbol] = useState("AAPL");
  const [error, setError] = useState("");

  function handleGenerate() {
    if (!symbol.trim()) { setError("Symbol is required"); return; }
    setError("");
    onGenerate(symbol.toUpperCase().trim());
  }

  return (
    <ModalPortal>
      <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
        <div className="w-full max-w-lg rounded-2xl border border-white/10 bg-neutral-950/95 shadow-2xl overflow-hidden">
          <div className="px-6 py-5 border-b border-white/10 flex items-start justify-between gap-4">
            <div>
              <p className="text-xs uppercase tracking-widest mb-1 font-semibold" style={{ color: "#8b3fd6" }}>Generate with AI</p>
              <h2 className="text-xl font-bold text-white">Author a strategy</h2>
              <p className="text-xs text-white/50 mt-1">Pick a stock. The model reads its current trend, momentum and volatility, then designs and backtests a strategy that fits.</p>
            </div>
            <button
              onClick={onClose}
              className="p-1.5 rounded-lg text-white/40 hover:text-white hover:bg-white/10 transition-colors mt-0.5 shrink-0"
            >
              ✕
            </button>
          </div>

          <div className="px-6 py-5 space-y-5">
            <div className="space-y-1.5">
              <label className="text-xs font-semibold text-white/50 uppercase tracking-wider">Symbol</label>
              <SymbolCombobox
                value={symbol}
                onChange={(v) => setSymbol(v.toUpperCase())}
                placeholder="AAPL"
                inputClassName="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-sm text-white placeholder:text-white/30 focus:outline-none focus:border-[#5600a2]/60 focus:ring-1 focus:ring-[#5600a2]/40 transition-all"
              />
            </div>

            {error && (
              <div className="p-3 bg-red-500/10 border border-red-500/20 rounded-xl text-red-400 text-xs font-medium flex items-center gap-2">
                <AlertTriangle size={13} /> {error}
              </div>
            )}
          </div>

          <div className="px-6 py-4 border-t border-white/10 flex items-center gap-3">
            <button
              onClick={onClose}
              className="flex-1 py-3 rounded-xl text-sm font-semibold uppercase tracking-wider text-white/50 border border-white/10 hover:bg-white/5 transition-all"
            >
              Cancel
            </button>
            <PixelHover
              gap={3}
              speed={40}
              colors="#a06cd5,#7d34c4,#5600a2"
              className="group flex-1 rounded-xl border border-white/10 bg-white/5 hover:bg-white/10 hover:border-[#5600a2]/60 transition-all active:scale-[0.98]"
            >
              <button
                onClick={handleGenerate}
                disabled={busy}
                className="w-full py-3 rounded-xl text-sm font-bold uppercase tracking-wider text-white bg-transparent group-hover:text-[#c9a6f0] transition-colors disabled:opacity-50 inline-flex items-center justify-center gap-2"
              >
                <Sparkles size={14} />
                {busy ? "Generating…" : "Generate"}
              </button>
            </PixelHover>
          </div>
        </div>
      </div>
    </ModalPortal>
  );
}


type DeployTarget =
  | { kind: "custom"; strategy: CustomStrategy }
  | { kind: "preset"; preset: StrategyPreset };

function backtestConfig(t: DeployTarget): unknown {
  if (t.kind === "custom") return t.strategy;
  const p = t.preset;
  return {
    name: p.name,
    entry: p.entry,
    exit: p.exit,
    stopLossPct: p.stopLossPct,
    takeProfitPct: p.takeProfitPct,
  };
}

export default function Strategies() {
  const [customStrategies, setCustomStrategies] = useState<CustomStrategy[]>([]);
  const [deployTarget, setDeployTarget] = useState<DeployTarget | null>(null);
  const [backtestTarget, setBacktestTarget] = useState<DeployTarget | null>(null);
  const [expandedCard, setExpandedCard] = useState<string | null>(null);
  const [deployError, setDeployError] = useState("");
  const [presetBusy, setPresetBusy] = useState(false);
  const [genBusy, setGenBusy] = useState(false);
  const [genError, setGenError] = useState("");
  const [genRationale, setGenRationale] = useState("");
  const [genBacktest, setGenBacktest] = useState<BacktestSummary | null>(null);
  const [genModalOpen, setGenModalOpen] = useState(false);

  const { data: deployments = [] } = useDeployments();
  const deployMutation = useDeployStrategy();
  const stopMutation = useStopDeployment();
  const deleteDeploymentMutation = useDeleteDeployment();
  const [deploymentActionError, setDeploymentActionError] = useState("");

  useEffect(() => {
    loadCustomStrategies()
      .then(setCustomStrategies)
      .catch(() => setCustomStrategies([]));
  }, []);

  useEffect(() => {
    let active = true;
    isStrategyGenerationRunning().then((running) => {
      if (!running || !active) return;
      setGenBusy(true);
      setGenError("");
      resumeStrategyGeneration()
        .then((generated) => {
          if (active) return applyGenerated(generated);
        })
        .catch(() => {
          if (active) setGenError("Could not finish the strategy that was generating. Try again.");
        })
        .finally(() => {
          if (active) setGenBusy(false);
        });
    });
    return () => {
      active = false;
    };
  }, []);

  async function applyGenerated(generated: GeneratedStrategy) {
    const cs: CustomStrategy = {
      ...generated.config,
      name: generated.name,
      description: generated.rationale,
      createdAt: new Date().toISOString(),
    };
    await createCustomStrategy(cs);
    const list = await loadCustomStrategies();
    setCustomStrategies(list);
    setGenRationale(generated.rationale);
    setGenBacktest(generated.backtest ?? null);
  }

  async function handleGenerate(symbol: string) {
    if (genBusy) return;
    setGenModalOpen(false);
    setGenBusy(true);
    setGenError("");
    setGenRationale("");
    setGenBacktest(null);
    try {
      const generated = await generateStrategy(symbol);
      await applyGenerated(generated);
    } catch {
      setGenError("Could not generate a strategy right now. Try again in a moment.");
    } finally {
      setGenBusy(false);
    }
  }

  function deleteCustomStrategy(id: string) {
    if (!window.confirm("Delete this custom strategy?")) return;
    setCustomStrategies((prev) => prev.filter((c) => c.id !== id));
    deleteCustomStrategyById(id).catch(() => {
      loadCustomStrategies().then(setCustomStrategies).catch(() => { });
    });
  }

  async function handleDeploy(symbol: string, positionSizeCents: number) {
    if (!deployTarget) return;
    setDeployError("");

    try {
      let strategyId: string;
      if (deployTarget.kind === "custom") {
        strategyId = deployTarget.strategy.id;
      } else {
        const p = deployTarget.preset;
        const existing = customStrategies.find((c) => c.name === p.name);
        if (existing) {
          strategyId = existing.id;
        } else {
          setPresetBusy(true);
          const created = await createCustomStrategy({
            id: "",
            name: p.name,
            description: p.tagline,
            risk: p.risk,
            tags: p.tags,
            entry: p.entry,
            exit: p.exit,
            stopLossPct: p.stopLossPct,
            takeProfitPct: p.takeProfitPct,
            createdAt: new Date().toISOString(),
          });
          strategyId = created.id;
          loadCustomStrategies().then(setCustomStrategies).catch(() => { });
        }
      }

      await deployMutation.mutateAsync({ strategyId, symbol, positionSizeCents });
      setDeployTarget(null);
    } catch (err) {
      setDeployError(err instanceof Error ? err.message : "Failed to deploy strategy");
    } finally {
      setPresetBusy(false);
    }
  }

  function handleStop(id: string) {
    stopMutation.mutate(id);
  }

  function handleRestart(d: Deployment) {
    setDeploymentActionError("");
    deployMutation
      .mutateAsync({
        strategyId: d.strategy_id,
        symbol: d.symbol,
        positionSizeCents: d.position_size_cents,
      })
      .catch((err) =>
        setDeploymentActionError(
          err instanceof Error ? err.message : "Failed to restart deployment"
        )
      );
  }

  function handleRemoveDeployment(id: string) {
    setDeploymentActionError("");
    deleteDeploymentMutation.mutate(id, {
      onError: (err) =>
        setDeploymentActionError(
          err instanceof Error ? err.message : "Failed to remove deployment"
        ),
    });
  }

  const visibleDeployments = useMemo(() => {
    const byStrategySymbol = new Map<string, Deployment>();
    for (const d of deployments) {
      const key = `${d.strategy_id}::${d.symbol}`;
      const existing = byStrategySymbol.get(key);
      if (!existing || (existing.status !== "running" && d.status === "running")) {
        byStrategySymbol.set(key, d);
      }
    }
    return Array.from(byStrategySymbol.values());
  }, [deployments]);

  const running = visibleDeployments.filter((d) => d.status === "running").length;

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
                className="text-4xl md:text-6xl font-light text-white tracking-tighter drop-shadow-md"
              >
                Strategies
              </TextEffect>
              <motion.p
                initial={{ opacity: 0, y: 6 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.4, duration: 0.4 }}
                className="text-sm text-white/50 flex items-center gap-2 font-medium tracking-wide uppercase"
              >
                <motion.span
                  className="flex h-2 w-2 rounded-full bg-emerald-500"
                  animate={{
                    scale: [1, 1.4, 1],
                    boxShadow: [
                      '0 0 8px rgba(16,185,129,0.8)',
                      '0 0 16px rgba(16,185,129,0.95)',
                      '0 0 8px rgba(16,185,129,0.8)',
                    ],
                  }}
                  transition={{ duration: 2, repeat: Infinity, ease: 'easeInOut' }}
                />
                {running > 0 ? `${running} automation${running > 1 ? "s" : ""} running` : "No active automations"}
              </motion.p>
            </div>

            <div className="flex items-center gap-3 flex-wrap">
              <motion.div
                initial={{ opacity: 0, scale: 0.95 }}
                animate={{ opacity: 1, scale: 1 }}
                transition={{ delay: 0.46, duration: 0.3 }}
                className="flex items-center"
              >
                <MarketStatusBadge variant="compact" />
              </motion.div>
              <motion.div
                initial={{ opacity: 0, scale: 0.95 }}
                animate={{ opacity: 1, scale: 1 }}
                transition={{ delay: 0.5, duration: 0.3 }}
                className="flex items-center gap-2 px-4 py-2 rounded-xl border border-white/10 bg-white/5 text-xs text-white/50 font-medium"
              >
                <Bot size={14} className="text-emerald-400" />
                <span>{visibleDeployments.length} deployed</span>
              </motion.div>
              <motion.div
                initial={{ opacity: 0, scale: 0.95 }}
                animate={{ opacity: 1, scale: 1 }}
                transition={{ delay: 0.58, duration: 0.3 }}
                className="flex items-center gap-2 px-4 py-2 rounded-xl border border-white/10 bg-white/5 text-xs text-white/50 font-medium"
              >
                <Layers size={14} className="text-white/40" />
                <span>{PRESETS.length + customStrategies.length} available</span>
              </motion.div>
              <motion.div
                initial={{ opacity: 0, scale: 0.95 }}
                animate={{ opacity: 1, scale: 1 }}
                transition={{ delay: 0.66, duration: 0.3 }}
              >
                <PixelHover
                  variant="emerald"
                  className="group rounded-xl border border-white/10 bg-white/5 hover:bg-white/10 hover:border-emerald-500/40 transition-colors"
                >
                  <Link
                    href="/strategies/new"
                    className="inline-flex items-center gap-2 px-4 py-2 rounded-xl text-xs font-bold uppercase tracking-wider text-white bg-transparent group-hover:text-emerald-300 transition-colors"
                  >
                    Create Strategy
                  </Link>
                </PixelHover>
              </motion.div>

              <motion.div
                initial={{ opacity: 0, scale: 0.95 }}
                animate={{ opacity: 1, scale: 1 }}
                transition={{ delay: 0.72, duration: 0.3 }}
              >
                <PixelHover
                  gap={3}
                  speed={40}
                  colors="#a06cd5,#7d34c4,#5600a2"
                  className="group rounded-xl border border-white/10 bg-white/5 hover:bg-white/10 hover:border-[#5600a2]/60 transition-colors"
                >
                  <button
                    onClick={() => setGenModalOpen(true)}
                    disabled={genBusy}
                    className="inline-flex items-center gap-2 px-4 py-2 rounded-xl text-xs font-bold uppercase tracking-wider text-white bg-transparent transition-colors group-hover:text-[#c9a6f0] disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {genBusy ? "Generating…" : "Generate with AI"}
                  </button>
                </PixelHover>
              </motion.div>
            </div>
          </header>
        </RevealItem>

        {(genBusy || genRationale || genError) && (
          <RevealItem>
            <GlassSurface borderRadius={16} order="start" alignItems="stretch" flexDirection="col" innerClassName="p-4">
              {genError ? (
                <div className="flex items-start gap-3 text-red-400 text-sm">
                  <AlertTriangle size={16} className="mt-0.5 shrink-0" />
                  <p>{genError}</p>
                </div>
              ) : genBusy ? (
                <div className="flex items-start gap-3 text-sm text-neutral-300">
                  <Sparkles size={16} className="mt-0.5 shrink-0 animate-pulse" style={{ color: "#5600a2" }} />
                  <p><span className="font-medium" style={{ color: "#8b3fd6" }}>Authoring a strategy… </span>the model is drafting, validating and backtesting rules. This survives a refresh.</p>
                </div>
              ) : (
                <div className="space-y-3">
                  <div className="flex items-start gap-3 text-sm text-neutral-300">
                    <Sparkles size={16} className="mt-0.5 shrink-0" style={{ color: "#5600a2" }} />
                    <p><span className="font-medium" style={{ color: "#8b3fd6" }}>Added a strategy: </span>{genRationale}</p>
                  </div>
                  {genBacktest && genBacktest.num_trades > 0 && (
                    <div className="flex flex-wrap gap-2 pl-7">
                      <span className="px-2 py-1 rounded-md text-[10px] font-semibold bg-white/5 border border-white/10 text-white/60 tracking-wider">
                        Return {genBacktest.total_return_pct.toFixed(1)}%
                      </span>
                      <span className="px-2 py-1 rounded-md text-[10px] font-semibold bg-white/5 border border-white/10 text-white/60 tracking-wider">
                        Max DD {genBacktest.max_drawdown_pct.toFixed(1)}%
                      </span>
                      <span className="px-2 py-1 rounded-md text-[10px] font-semibold bg-white/5 border border-white/10 text-white/60 tracking-wider">
                        Sharpe {genBacktest.sharpe.toFixed(2)}
                      </span>
                      <span className="px-2 py-1 rounded-md text-[10px] font-semibold bg-white/5 border border-white/10 text-white/60 tracking-wider">
                        Win {(genBacktest.win_rate * 100).toFixed(0)}%
                      </span>
                      <span className="px-2 py-1 rounded-md text-[10px] font-semibold bg-white/5 border border-white/10 text-white/60 tracking-wider">
                        {genBacktest.num_trades} trades
                      </span>
                    </div>
                  )}
                </div>
              )}
            </GlassSurface>
          </RevealItem>
        )}

        {visibleDeployments.length > 0 && (
          <motion.section
            initial={{ opacity: 0, y: 16 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, ease: [0.22, 1, 0.36, 1] }}
            className="space-y-4"
          >
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-semibold tracking-tight text-white flex items-center gap-2">
                <Cpu size={18} className="text-emerald-400" />
                Active Automations
              </h2>
              <span className="text-xs text-white/30 font-medium">{visibleDeployments.length} total</span>
            </div>

            {deploymentActionError && (
              <div className="p-3 bg-red-500/10 border border-red-500/20 rounded-xl text-red-400 text-xs font-medium flex items-center gap-2">
                <AlertTriangle size={13} /> {deploymentActionError}
              </div>
            )}

            <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
              {visibleDeployments.map((d, idx) => (
                <motion.div
                  key={d.id}
                  initial={{ opacity: 0, y: 14, scale: 0.96 }}
                  animate={{ opacity: 1, y: 0, scale: 1 }}
                  transition={{ duration: 0.45, ease: [0.22, 1, 0.36, 1], delay: 0.1 + idx * 0.06 }}
                  className="rounded-2xl border border-white/10 bg-black/30 backdrop-blur-md p-5 space-y-4 shadow-xl shadow-black/20"
                >
                  <div className="flex items-start justify-between gap-3">
                    <div className="flex items-center gap-3">
                      <div className="p-2 rounded-lg bg-white/5 border border-white/10 text-white/60">
                        <Zap size={18} />
                      </div>
                      <div>
                        <p className="font-bold text-white text-sm">{d.strategy_name || "Strategy"}</p>
                        <p className="text-xs text-white/50 font-semibold tracking-wider">{d.symbol}</p>
                      </div>
                    </div>

                    <span className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md text-[10px] font-bold uppercase tracking-wider border ${statusColor(d.status)}`}>
                      <CircleDot size={8} className={d.status === "running" ? "animate-pulse" : ""} />
                      {d.status}
                    </span>
                  </div>

                  <div className="grid grid-cols-3 gap-3">
                    <div className="bg-white/5 rounded-xl p-3 text-center border border-white/5">
                      <p className="text-[10px] text-white/40 uppercase tracking-wider font-medium mb-1">Position</p>
                      {d.in_position ? (
                        <p className="text-sm font-bold text-emerald-400">
                          {d.qty} @ ${formatCents(d.entry_price_cents)}
                        </p>
                      ) : (
                        <p className="text-sm font-bold text-white/50">Flat</p>
                      )}
                    </div>
                    <div className="bg-white/5 rounded-xl p-3 text-center border border-white/5">
                      <p className="text-[10px] text-white/40 uppercase tracking-wider font-medium mb-1">Size</p>
                      <p className="text-sm font-bold text-white/80">${formatCents(d.position_size_cents)}</p>
                    </div>
                    <Link
                      href={`/strategies/${d.strategy_id}`}
                      className="bg-white/5 rounded-xl p-3 text-center border border-white/5 hover:bg-white/10 hover:border-emerald-500/30 transition-colors group"
                    >
                      <p className="text-[10px] text-white/40 uppercase tracking-wider font-medium mb-1">Trades</p>
                      <p className="text-sm font-bold text-white/70 group-hover:text-emerald-300 transition-colors">View →</p>
                    </Link>
                  </div>

                  <div className="flex items-center gap-2 pt-1">
                    {d.status === "running" && (
                      <PixelHover
                        variant="red"
                        className="group rounded-lg border border-white/10 hover:bg-red-500/10 hover:border-red-500/20 transition-colors"
                      >
                        <motion.button
                          whileTap={{ scale: 0.95 }}
                          onClick={() => handleStop(d.id)}
                          disabled={stopMutation.isPending}
                          className="flex items-center gap-1.5 px-3 py-2 rounded-lg text-xs font-bold uppercase tracking-wider text-white/40 bg-transparent group-hover:text-red-400 transition-colors disabled:opacity-50"
                        >
                          <Square size={12} />
                          Stop
                        </motion.button>
                      </PixelHover>
                    )}
                    {d.status !== "running" && (
                      <>
                        <PixelHover
                          variant="emerald"
                          className="group rounded-lg border border-white/10 hover:bg-emerald-500/10 hover:border-emerald-500/20 transition-colors"
                        >
                          <motion.button
                            whileTap={{ scale: 0.95 }}
                            onClick={() => handleRestart(d)}
                            disabled={deployMutation.isPending}
                            className="flex items-center gap-1.5 px-3 py-2 rounded-lg text-xs font-bold uppercase tracking-wider text-white/40 bg-transparent group-hover:text-emerald-400 transition-colors disabled:opacity-50"
                          >
                            Start again
                          </motion.button>
                        </PixelHover>
                        <PixelHover
                          variant="red"
                          className="group rounded-lg border border-white/10 hover:bg-red-500/10 hover:border-red-500/20 transition-colors"
                        >
                          <motion.button
                            whileTap={{ scale: 0.95 }}
                            onClick={() => handleRemoveDeployment(d.id)}
                            disabled={deleteDeploymentMutation.isPending}
                            className="flex items-center gap-1.5 px-3 py-2 rounded-lg text-xs font-bold uppercase tracking-wider text-white/40 bg-transparent group-hover:text-red-400 transition-colors disabled:opacity-50"
                          >
                            <Trash2 size={12} />
                            Remove
                          </motion.button>
                        </PixelHover>
                      </>
                    )}
                    {d.in_position && d.status !== "running" && (
                      <p className="text-[10px] text-amber-400/80 font-medium">
                        Stopped with an open position, close it from Orders.
                      </p>
                    )}
                    <p className="ml-auto text-[10px] text-white/25 font-medium">
                      {new Date(d.created_at).toLocaleDateString()}
                    </p>
                  </div>
                </motion.div>
              ))}
            </div>
          </motion.section>
        )}

        {customStrategies.length > 0 && (
          <motion.section
            initial={{ opacity: 0, y: 16 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, ease: [0.22, 1, 0.36, 1] }}
            className="space-y-6"
          >
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-semibold tracking-tight text-white flex items-center gap-2">
                <Wand2 size={18} className="text-emerald-400" />
                Your Strategies
              </h2>
              <span className="text-xs text-white/30 font-medium uppercase tracking-wider">
                {customStrategies.length} custom
              </span>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              {customStrategies.map((c, idx) => {
                const isExpanded = expandedCard === c.id;
                return (
                  <motion.div
                    key={c.id}
                    initial={{ opacity: 0, y: 14, scale: 0.96 }}
                    animate={{ opacity: 1, y: 0, scale: 1 }}
                    transition={{ duration: 0.45, ease: [0.22, 1, 0.36, 1], delay: 0.1 + idx * 0.06 }}
                    className="rounded-2xl border border-emerald-500/20 bg-emerald-500/3 backdrop-blur-md hover:border-emerald-500/40 transition-colors shadow-xl shadow-black/20 group"
                  >
                    <div className="p-6 space-y-5">
                      <div className="flex items-start justify-between gap-4">
                        <div className="flex items-center gap-4">
                          <div className="p-3 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 shrink-0">
                            <Wand2 size={22} />
                          </div>
                          <div>
                            <div className="flex items-center gap-2 flex-wrap">
                              <h3 className="text-lg font-bold text-white tracking-tight">{c.name}</h3>
                              <span className="px-1.5 py-0.5 rounded text-[9px] font-bold uppercase tracking-wider text-emerald-400 bg-emerald-500/10 border border-emerald-500/20">
                                Custom
                              </span>
                            </div>
                            {c.description && (
                              <p className="text-xs text-white/50 mt-0.5 leading-relaxed line-clamp-2">{c.description}</p>
                            )}
                          </div>
                        </div>
                        <div className="flex items-center gap-2 shrink-0">
                          <span className={`inline-flex items-center gap-1 px-2.5 py-1 rounded-md text-[10px] font-bold uppercase tracking-wider border ${riskColor(c.risk)}`}>
                            <ShieldCheck size={10} />
                            {c.risk}
                          </span>
                          <PixelHover
                            variant="red"
                            className="group rounded-lg hover:bg-red-500/10 transition-all"
                          >
                            <button
                              onClick={() => deleteCustomStrategy(c.id)}
                              className="p-2 rounded-lg text-white/30 bg-transparent group-hover:text-red-400 transition-colors"
                              title="Delete strategy"
                            >
                              <Trash2 size={14} />
                            </button>
                          </PixelHover>
                        </div>
                      </div>

                      <div className="space-y-2">
                        <div className="flex items-start gap-2 text-xs">
                          <span className="px-1.5 py-0.5 rounded text-[9px] font-bold uppercase tracking-wider text-emerald-400 bg-emerald-500/10 border border-emerald-500/20 shrink-0 mt-0.5">Buy</span>
                          <span className="text-white/70 wrap-break-word leading-relaxed font-medium">{formatGroup(c.entry)}</span>
                        </div>
                        <div className="flex items-start gap-2 text-xs">
                          <span className="px-1.5 py-0.5 rounded text-[9px] font-bold uppercase tracking-wider text-red-400 bg-red-500/10 border border-red-500/20 shrink-0 mt-0.5">Sell</span>
                          <span className="text-white/70 wrap-break-word leading-relaxed font-medium">{formatGroup(c.exit)}</span>
                        </div>
                      </div>

                      {c.tags.length > 0 && (
                        <div className="flex flex-wrap gap-2">
                          {c.tags.map((tag) => (
                            <span key={tag} className="px-2.5 py-1 rounded-md text-[10px] font-semibold bg-white/5 border border-white/10 text-white/50 tracking-wider">
                              {tag}
                            </span>
                          ))}
                        </div>
                      )}

                      <button
                        onClick={() => setExpandedCard(isExpanded ? null : c.id)}
                        className="flex items-center gap-1.5 text-xs text-white/40 hover:text-white/70 transition-colors font-medium"
                      >
                        {isExpanded ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
                        {isExpanded ? "Less detail" : "More detail"}
                      </button>

                      <AnimatePresence initial={false}>
                        {isExpanded && (
                          <motion.div
                            initial={{ opacity: 0, height: 0, y: -6 }}
                            animate={{ opacity: 1, height: 'auto', y: 0 }}
                            exit={{ opacity: 0, height: 0, y: -6 }}
                            transition={{ duration: 0.25, ease: [0.22, 1, 0.36, 1] }}
                            className="rounded-xl border border-white/10 bg-white/3 p-4 space-y-2 overflow-hidden"
                          >
                            <div className="flex justify-between text-xs">
                              <span className="text-white/40">Stop Loss</span>
                              <span className="text-white/70 font-medium">{c.stopLossPct}%</span>
                            </div>
                            <div className="flex justify-between text-xs">
                              <span className="text-white/40">Take Profit</span>
                              <span className="text-white/70 font-medium">{c.takeProfitPct}%</span>
                            </div>
                            <div className="flex justify-between text-xs">
                              <span className="text-white/40">Created</span>
                              <span className="text-white/70 font-medium">{new Date(c.createdAt).toLocaleDateString()}</span>
                            </div>
                          </motion.div>
                        )}
                      </AnimatePresence>

                      <div className="flex items-stretch gap-2">
                        <PixelHover
                          variant="mono"
                          className="group/backtest flex-1 rounded-xl border border-white/10 bg-white/5 hover:bg-white/10 hover:border-white/25 transition-colors"
                        >
                          <motion.button
                            whileTap={{ scale: 0.97 }}
                            onClick={() => setBacktestTarget({ kind: "custom", strategy: c })}
                            className="w-full flex items-center justify-center gap-2 py-3 rounded-xl text-sm font-bold uppercase tracking-wider text-white/60 bg-transparent group-hover/backtest:text-white transition-colors"
                          >
                            <FlaskConical size={14} />
                            Backtest
                          </motion.button>
                        </PixelHover>
                        <PixelHover
                          variant="emerald"
                          className="group flex-1 rounded-xl border border-white/10 bg-white/5 hover:bg-white/10 hover:border-emerald-500/40 transition-colors"
                        >
                          <motion.button
                            whileTap={{ scale: 0.97 }}
                            onClick={() => {
                              setDeployError("");
                              setDeployTarget({ kind: "custom", strategy: c });
                            }}
                            className="w-full flex items-center justify-center gap-2 py-3 rounded-xl text-sm font-bold uppercase tracking-wider text-white bg-transparent group-hover:text-emerald-300 transition-colors"
                          >
                            <Plus size={15} />
                            Deploy
                          </motion.button>
                        </PixelHover>
                      </div>
                    </div>
                  </motion.div>
                );
              })}
            </div>
          </motion.section>
        )}

        <motion.section
          initial={{ opacity: 0, y: 16 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, ease: [0.22, 1, 0.36, 1], delay: 0.05 }}
          className="space-y-6"
        >
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold tracking-tight text-white flex items-center gap-2">
              <FlaskConical size={18} className="text-white/60" />
              Strategy Library
            </h2>
            <span className="text-xs text-white/30 font-medium uppercase tracking-wider">{PRESETS.length} presets</span>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {PRESETS.map((t, idx) => {
              const isExpanded = expandedCard === t.key;
              return (
                <motion.div
                  key={t.key}
                  initial={{ opacity: 0, y: 16, scale: 0.96 }}
                  animate={{ opacity: 1, y: 0, scale: 1 }}
                  transition={{ duration: 0.45, ease: [0.22, 1, 0.36, 1], delay: 0.2 + idx * 0.05 }}
                >
                  <GlassSurface
                    displace={15}
                    distortionScale={-80}
                    opacity={0.65}
                    borderRadius={24}
                    className="group cursor-pointer hover:border-white/25 transition-colors"
                  >
                    <div className="p-6 space-y-5">
                      <div className="flex items-start justify-between gap-4">
                        <div className="flex items-center gap-4">
                          <div className="p-3 rounded-xl bg-white/5 border border-white/10 text-white/70 group-hover:text-white group-hover:border-white/20 transition-all shrink-0">
                            {t.icon}
                          </div>
                          <div>
                            <h3 className="text-lg font-bold text-white tracking-tight">{t.name}</h3>
                            <p className="text-xs text-white/50 mt-0.5 leading-relaxed">{t.tagline}</p>
                          </div>
                        </div>
                        <span className={`inline-flex items-center gap-1 px-2.5 py-1 rounded-md text-[10px] font-bold uppercase tracking-wider border shrink-0 ${riskColor(t.risk)}`}>
                          <ShieldCheck size={10} />
                          {t.risk}
                        </span>
                      </div>

                      <div className="space-y-2">
                        <div className="flex items-start gap-2 text-xs">
                          <span className="px-1.5 py-0.5 rounded text-[9px] font-bold uppercase tracking-wider text-emerald-400 bg-emerald-500/10 border border-emerald-500/20 shrink-0 mt-0.5">Buy</span>
                          <span className="text-white/70 wrap-break-word leading-relaxed font-medium">{formatGroup(t.entry)}</span>
                        </div>
                        <div className="flex items-start gap-2 text-xs">
                          <span className="px-1.5 py-0.5 rounded text-[9px] font-bold uppercase tracking-wider text-red-400 bg-red-500/10 border border-red-500/20 shrink-0 mt-0.5">Sell</span>
                          <span className="text-white/70 wrap-break-word leading-relaxed font-medium">
                            {formatGroup(t.exit)} · SL {t.stopLossPct}% · TP {t.takeProfitPct}%
                          </span>
                        </div>
                      </div>

                      <div className="flex flex-wrap gap-2">
                        {t.tags.map((tag) => (
                          <span key={tag} className="px-2.5 py-1 rounded-md text-[10px] font-semibold bg-white/5 border border-white/10 text-white/50 tracking-wider">
                            {tag}
                          </span>
                        ))}
                      </div>

                      <button
                        onClick={() => setExpandedCard(isExpanded ? null : t.key)}
                        className="flex items-center gap-1.5 text-xs text-white/40 hover:text-white/70 transition-colors font-medium"
                      >
                        {isExpanded ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
                        {isExpanded ? "Less detail" : "More detail"}
                      </button>

                      <AnimatePresence initial={false}>
                        {isExpanded && (
                          <motion.div
                            initial={{ opacity: 0, height: 0, y: -6 }}
                            animate={{ opacity: 1, height: 'auto', y: 0 }}
                            exit={{ opacity: 0, height: 0, y: -6 }}
                            transition={{ duration: 0.25, ease: [0.22, 1, 0.36, 1] }}
                            className="space-y-3 overflow-hidden"
                          >
                            <p className="text-xs text-white/50 leading-relaxed">{t.description}</p>
                          </motion.div>
                        )}
                      </AnimatePresence>

                      <div className="flex items-stretch gap-2">
                        <PixelHover
                          variant="mono"
                          className="group/backtest flex-1 rounded-xl border border-white/10 bg-white/5 hover:bg-white/10 hover:border-white/25 transition-colors"
                        >
                          <motion.button
                            whileTap={{ scale: 0.97 }}
                            onClick={() => setBacktestTarget({ kind: "preset", preset: t })}
                            className="w-full flex items-center justify-center gap-2 py-3 rounded-xl text-sm font-bold uppercase tracking-wider text-white/60 bg-transparent group-hover/backtest:text-white transition-colors"
                          >
                            <FlaskConical size={14} />
                            Backtest
                          </motion.button>
                        </PixelHover>
                        <PixelHover
                          variant="emerald"
                          className="group/deploy flex-1 rounded-xl bg-white/5 border border-white/10 hover:bg-white/10 hover:border-emerald-500/40 transition-colors"
                        >
                          <motion.button
                            whileTap={{ scale: 0.97 }}
                            onClick={() => {
                              setDeployError("");
                              setDeployTarget({ kind: "preset", preset: t });
                            }}
                            className="w-full flex items-center justify-center gap-2 py-3 rounded-xl text-sm font-bold uppercase tracking-wider bg-transparent text-white/60 group-hover/deploy:text-emerald-300 transition-colors"
                          >
                            <Plus size={15} />
                            Deploy
                          </motion.button>
                        </PixelHover>
                      </div>
                    </div>
                  </GlassSurface>
                </motion.div>
              );
            })}
          </div>
        </motion.section>

      </RevealStagger>

      {genModalOpen && (
        <GenerateModal
          busy={genBusy}
          onClose={() => setGenModalOpen(false)}
          onGenerate={handleGenerate}
        />
      )}

      {deployTarget && (
        <DeployModal
          name={deployTarget.kind === "custom" ? deployTarget.strategy.name : deployTarget.preset.name}
          tagline={
            deployTarget.kind === "custom"
              ? deployTarget.strategy.description || "Custom rule-based strategy"
              : deployTarget.preset.tagline
          }
          rulesPreview={{
            entry: formatGroup(deployTarget.kind === "custom" ? deployTarget.strategy.entry : deployTarget.preset.entry),
            exit: formatGroup(deployTarget.kind === "custom" ? deployTarget.strategy.exit : deployTarget.preset.exit),
          }}
          busy={deployMutation.isPending || presetBusy}
          apiError={deployError}
          onClose={() => {
            setDeployTarget(null);
            setDeployError("");
          }}
          onDeploy={handleDeploy}
        />
      )}

      {backtestTarget && (
        <BacktestModal
          name={backtestTarget.kind === "custom" ? backtestTarget.strategy.name : backtestTarget.preset.name}
          tagline={
            backtestTarget.kind === "custom"
              ? backtestTarget.strategy.description || "Custom rule-based strategy"
              : backtestTarget.preset.tagline
          }
          config={backtestConfig(backtestTarget)}
          onClose={() => setBacktestTarget(null)}
        />
      )}
    </PageEnter>
  );
}
