'use client';

import { useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import {
  INDICATORS,
  IndicatorKind,
  IndicatorRef,
  OPERATORS,
  Operator,
  Rule,
  RuleGroup,
  RuleRHS,
  RiskLevel,
  CustomStrategy,
  defaultIndicatorRef,
  formatGroup,
  createCustomStrategy,
  newExitRule,
  newRule,
  compatibleIndicators,
  valueBoundsFor,
  coerceRuleForLhs,
} from "@/lib/strategies";
import {
  AlertTriangle,
  ArrowLeft,
  ChevronDown,
  Eye,
  Plus,
  ShieldCheck,
  Sparkles,
  Trash2,
} from "lucide-react";
import PixelHover from "@/components/ui/PixelHover";
import { TextEffect } from "@/components/primitives/TextEffect";
import { PageEnter, RevealStagger, RevealItem } from "@/components/primitives/Reveal";
import { motion } from "motion/react";

function uid() {
  return Math.random().toString(36).slice(2, 10);
}


function SectionCard({
  title,
  subtitle,
  accent,
  children,
}: {
  title: string;
  subtitle?: string;
  accent?: "emerald" | "red" | "white";
  children: React.ReactNode;
}) {
  const accentColor =
    accent === "emerald"
      ? "text-emerald-400"
      : accent === "red"
        ? "text-red-400"
        : "text-white/70";
  return (
    <div className="rounded-2xl border border-white/10 bg-black/30 backdrop-blur-md p-6 space-y-5 shadow-xl shadow-black/20">
      <div className="flex items-start justify-between gap-4 pb-4 border-b border-white/10">
        <div>
          <h2 className={`text-sm font-bold uppercase tracking-widest ${accentColor}`}>
            {title}
          </h2>
          {subtitle && <p className="text-xs text-white/40 mt-1">{subtitle}</p>}
        </div>
      </div>
      {children}
    </div>
  );
}

function FieldLabel({ children }: { children: React.ReactNode }) {
  return (
    <label className="text-[10px] font-semibold text-white/40 uppercase tracking-widest">
      {children}
    </label>
  );
}


function IndicatorPicker({
  value,
  onChange,
  compact = false,
  only,
}: {
  value: IndicatorRef;
  onChange: (v: IndicatorRef) => void;
  compact?: boolean;
  only?: IndicatorKind[];
}) {
  const spec = INDICATORS[value.kind];

  const grouped = useMemo(() => {
    const allowed = only ? new Set(only) : null;
    const g: Record<string, { kind: IndicatorKind; label: string }[]> = {};
    (Object.keys(INDICATORS) as IndicatorKind[]).forEach((k) => {
      if (allowed && !allowed.has(k)) return;
      const s = INDICATORS[k];
      (g[s.category] ||= []).push({ kind: k, label: s.label });
    });
    return g;
  }, [only]);

  const categoryLabels: Record<string, string> = {
    price: "Price",
    trend: "Trend",
    momentum: "Momentum",
    volatility: "Volatility",
    volume: "Volume",
  };

  return (
    <div className={`flex items-center gap-2 ${compact ? "" : "flex-wrap"}`}>
      <div className="relative">
        <select
          value={value.kind}
          onChange={(e) =>
            onChange(defaultIndicatorRef(e.target.value as IndicatorKind))
          }
          className="appearance-none bg-white/5 border border-white/10 rounded-lg pl-3 pr-7 py-2 text-xs font-bold text-white cursor-pointer hover:bg-white/10 transition-colors focus:outline-none focus:border-emerald-500/50"
        >
          {Object.entries(grouped).map(([category, items]) => (
            <optgroup key={category} label={categoryLabels[category] ?? category}>
              {items.map((i) => (
                <option key={i.kind} value={i.kind}>
                  {i.label}
                </option>
              ))}
            </optgroup>
          ))}
        </select>
        <ChevronDown
          size={12}
          className="absolute right-2 top-1/2 -translate-y-1/2 text-white/40 pointer-events-none"
        />
      </div>

      {spec.params.map((p) => (
        <div
          key={p.key}
          className="flex items-center gap-1 bg-white/5 border border-white/10 rounded-lg px-2 py-1.5"
        >
          <span className="text-[10px] text-white/40 uppercase tracking-wider font-semibold">
            {p.label}
          </span>
          <input
            type="number"
            value={value.params[p.key] ?? p.default}
            min={p.min}
            max={p.max}
            onChange={(e) =>
              onChange({
                ...value,
                params: {
                  ...value.params,
                  [p.key]: Number(e.target.value),
                },
              })
            }
            className="w-12 bg-transparent text-sm text-white text-center focus:outline-none"
          />
        </div>
      ))}
    </div>
  );
}


function RHSPicker({
  value,
  onChange,
  lhsKind,
}: {
  value: RuleRHS;
  onChange: (v: RuleRHS) => void;
  lhsKind: IndicatorKind;
}) {
  const bounds = valueBoundsFor(lhsKind);
  return (
    <div className="flex items-center gap-2 flex-wrap">
      <div className="flex rounded-lg bg-white/5 border border-white/10 p-0.5">
        <button
          onClick={() =>
            onChange(
              value.kind === "value" ? value : { kind: "value", value: 0 }
            )
          }
          className={`px-2.5 py-1 rounded-md text-[10px] font-bold uppercase tracking-wider transition-all ${value.kind === "value"
            ? "bg-white/15 text-white"
            : "text-white/40 hover:text-white/70"
            }`}
        >
          Value
        </button>
        <button
          onClick={() =>
            onChange(
              value.kind === "indicator"
                ? value
                : { kind: "indicator", indicator: defaultIndicatorRef(lhsKind) }
            )
          }
          className={`px-2.5 py-1 rounded-md text-[10px] font-bold uppercase tracking-wider transition-all ${value.kind === "indicator"
            ? "bg-white/15 text-white"
            : "text-white/40 hover:text-white/70"
            }`}
        >
          Indicator
        </button>
      </div>

      {value.kind === "value" ? (
        <input
          type="number"
          value={value.value}
          min={bounds.min}
          max={bounds.max}
          onChange={(e) =>
            onChange({ kind: "value", value: Number(e.target.value) })
          }
          className="w-24 bg-white/5 border border-white/10 rounded-lg px-3 py-2 text-sm text-white font-semibold focus:outline-none focus:border-emerald-500/50"
        />
      ) : (
        <IndicatorPicker
          value={value.indicator}
          onChange={(ind) =>
            onChange({ kind: "indicator", indicator: ind })
          }
          only={compatibleIndicators(lhsKind)}
        />
      )}
    </div>
  );
}


function OperatorPicker({
  value,
  onChange,
}: {
  value: Operator;
  onChange: (v: Operator) => void;
}) {
  return (
    <div className="relative">
      <select
        value={value}
        onChange={(e) => onChange(e.target.value as Operator)}
        className="appearance-none bg-white/5 border border-white/10 rounded-lg pl-3 pr-8 py-2 text-xs font-bold text-white cursor-pointer hover:bg-white/10 transition-colors focus:outline-none focus:border-emerald-500/50"
      >
        {OPERATORS.map((o) => (
          <option key={o.value} value={o.value}>
            {o.label}
          </option>
        ))}
      </select>
      <ChevronDown
        size={12}
        className="absolute right-2 top-1/2 -translate-y-1/2 text-white/40 pointer-events-none"
      />
    </div>
  );
}


function RuleRow({
  rule,
  onChange,
  onDelete,
  index,
  combinator,
  showCombinator,
}: {
  rule: Rule;
  onChange: (r: Rule) => void;
  onDelete: () => void;
  index: number;
  combinator: "AND" | "OR";
  showCombinator: boolean;
}) {
  return (
    <div className="space-y-2">
      {showCombinator && (
        <div className="flex items-center gap-2 pl-4">
          <div className="h-4 w-px bg-white/10" />
          <span className="px-2 py-0.5 rounded-md text-[10px] font-bold uppercase tracking-widest text-white/40 bg-white/5 border border-white/10">
            {combinator}
          </span>
          <div className="h-4 w-px bg-white/10" />
        </div>
      )}

      <div className="flex items-start gap-3 p-4 rounded-xl bg-white/3 border border-white/10 hover:border-white/20 transition-colors">
        <div className="flex items-center justify-center w-6 h-6 rounded-md bg-white/5 border border-white/10 text-[10px] font-bold text-white/50 shrink-0 mt-1.5">
          {index + 1}
        </div>

        <div className="flex-1 flex items-center gap-3 flex-wrap">
          <IndicatorPicker
            value={rule.lhs}
            onChange={(v) => onChange(coerceRuleForLhs({ ...rule, lhs: v }))}
          />
          <OperatorPicker
            value={rule.op}
            onChange={(v) => onChange({ ...rule, op: v })}
          />
          <RHSPicker
            value={rule.rhs}
            onChange={(v) => onChange({ ...rule, rhs: v })}
            lhsKind={rule.lhs.kind}
          />
        </div>

        <PixelHover variant="red" className="group rounded-lg shrink-0 hover:bg-red-500/10 transition-all">
          <button
            onClick={onDelete}
            className="p-2 rounded-lg text-white/30 bg-transparent group-hover:text-red-400 transition-colors"
            title="Remove rule"
          >
            <Trash2 size={14} />
          </button>
        </PixelHover>
      </div>
    </div>
  );
}


function RuleGroupEditor({
  group,
  onChange,
  accent,
  emptyBuilder,
}: {
  group: RuleGroup;
  onChange: (g: RuleGroup) => void;
  accent: "emerald" | "red";
  emptyBuilder: () => Rule;
}) {
  function updateRule(id: string, next: Rule) {
    onChange({
      ...group,
      rules: group.rules.map((r) => (r.id === id ? next : r)),
    });
  }
  function deleteRule(id: string) {
    onChange({ ...group, rules: group.rules.filter((r) => r.id !== id) });
  }
  function addRule() {
    onChange({ ...group, rules: [...group.rules, emptyBuilder()] });
  }

  const buttonAccent =
    accent === "emerald"
      ? "hover:bg-emerald-500/10 hover:text-emerald-400 hover:border-emerald-500/30"
      : "hover:bg-red-500/10 hover:text-red-400 hover:border-red-500/30";

  return (
    <div className="space-y-4">
      {group.rules.length > 1 && (
        <div className="flex items-center gap-2">
          <FieldLabel>Combine rules with</FieldLabel>
          <div className="flex rounded-lg bg-white/5 border border-white/10 p-0.5">
            {(["AND", "OR"] as const).map((c) => (
              <button
                key={c}
                onClick={() => onChange({ ...group, combinator: c })}
                className={`px-3 py-1 rounded-md text-[10px] font-bold uppercase tracking-wider transition-all ${group.combinator === c
                  ? "bg-white/15 text-white"
                  : "text-white/40 hover:text-white/70"
                  }`}
              >
                {c}
              </button>
            ))}
          </div>
        </div>
      )}

      {group.rules.length === 0 ? (
        <div className="rounded-xl border border-dashed border-white/10 p-8 text-center">
          <p className="text-xs text-white/40 mb-3">No rules yet</p>
          <PixelHover
            variant={accent === "emerald" ? "emerald" : "red"}
            className={`inline-block rounded-lg border border-white/10 transition-all ${buttonAccent}`}
          >
            <button
              onClick={addRule}
              className="inline-flex items-center gap-2 px-4 py-2 rounded-lg text-xs font-bold uppercase tracking-wider text-white/70 bg-transparent"
            >
              <Plus size={14} />
              Add Rule
            </button>
          </PixelHover>
        </div>
      ) : (
        <>
          <div className="space-y-2">
            {group.rules.map((rule, idx) => (
              <RuleRow
                key={rule.id}
                rule={rule}
                index={idx}
                combinator={group.combinator}
                showCombinator={idx > 0}
                onChange={(next) => updateRule(rule.id, next)}
                onDelete={() => deleteRule(rule.id)}
              />
            ))}
          </div>

          <PixelHover
            variant={accent === "emerald" ? "emerald" : "red"}
            className={`inline-block rounded-lg border border-white/10 transition-all ${buttonAccent}`}
          >
            <button
              onClick={addRule}
              className="inline-flex items-center gap-2 px-4 py-2 rounded-lg text-xs font-bold uppercase tracking-wider text-white/70 bg-transparent"
            >
              <Plus size={14} />
              Add Rule
            </button>
          </PixelHover>
        </>
      )}
    </div>
  );
}


export default function StrategyBuilder() {
  const router = useRouter();

  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [risk, setRisk] = useState<RiskLevel>("medium");
  const [tagsInput, setTagsInput] = useState("");

  const [entry, setEntry] = useState<RuleGroup>({
    combinator: "AND",
    rules: [newRule()],
  });
  const [exit, setExit] = useState<RuleGroup>({
    combinator: "OR",
    rules: [newExitRule()],
  });

  const [stopLossPct, setStopLossPct] = useState("2");
  const [takeProfitPct, setTakeProfitPct] = useState("4");

  const [error, setError] = useState("");
  const [saving, setSaving] = useState(false);

  function validate(): string | null {
    if (!name.trim()) return "Strategy name is required";
    if (entry.rules.length === 0) return "Add at least one entry (buy) rule";
    for (const r of [...entry.rules, ...exit.rules]) {
      if (r.rhs.kind === "value" && Number.isNaN(r.rhs.value)) {
        return "Rule values must be numbers";
      }
    }
    const sl = parseFloat(stopLossPct);
    const tp = parseFloat(takeProfitPct);
    if (isNaN(sl) || sl < 0) return "Stop-loss must be a non-negative number";
    if (isNaN(tp) || tp < 0) return "Take-profit must be a non-negative number";
    return null;
  }

  function handleSave() {
    setError("");
    const err = validate();
    if (err) {
      setError(err);
      return;
    }
    setSaving(true);

    const strategy: CustomStrategy = {
      id: `custom_${uid()}`,
      name: name.trim(),
      description: description.trim() || "Custom strategy",
      risk,
      tags: tagsInput
        .split(",")
        .map((t) => t.trim())
        .filter(Boolean),
      entry,
      exit,
      stopLossPct: parseFloat(stopLossPct),
      takeProfitPct: parseFloat(takeProfitPct),
      createdAt: new Date().toISOString(),
    };

    createCustomStrategy(strategy)
      .then(() => router.push("/strategies"))
      .catch((e: Error) => {
        setSaving(false);
        setError(e?.message || "Failed to save strategy");
      });
  }

  const riskColor =
    risk === "low"
      ? "text-emerald-400 bg-emerald-500/10 border-emerald-500/20"
      : risk === "medium"
        ? "text-amber-400 bg-amber-500/10 border-amber-500/20"
        : "text-red-400 bg-red-500/10 border-red-500/20";

  return (
    <PageEnter className="min-h-screen w-full bg-transparent text-white font-mono p-4 pt-32 md:p-8 md:pt-32 xl:p-12 xl:pt-32 overflow-y-auto pointer-events-auto z-0 relative">
      <RevealStagger className="mx-auto max-w-5xl space-y-8 pb-32" stagger={0.08}>

        <RevealItem>
          <div className="space-y-6">
            <motion.div initial={{ opacity: 0, x: -10 }} animate={{ opacity: 1, x: 0 }} transition={{ duration: 0.4 }}>
              <Link
                href="/strategies"
                className="inline-flex items-center gap-1.5 text-xs text-white/40 hover:text-white/70 hover:gap-2.5 transition-all font-medium uppercase tracking-wider"
              >
                <ArrowLeft size={12} />
                Back to Strategies
              </Link>
            </motion.div>

            <header className="flex flex-col md:flex-row md:items-end justify-between gap-6 pb-6 border-b border-white/10">
              <div className="space-y-3">
                <TextEffect
                  as="h1"
                  preset="fade-in-blur"
                  per="word"
                  className="text-4xl md:text-5xl font-light text-white tracking-tighter drop-shadow-md flex items-center gap-4"
                >
                  Strategy Builder
                </TextEffect>
                <motion.p
                  initial={{ opacity: 0, y: 6 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: 0.4, duration: 0.4 }}
                  className="text-sm text-white/50 font-medium tracking-wide uppercase"
                >
                  Compose rules over technical indicators
                </motion.p>
              </div>
            </header>
          </div>
        </RevealItem>

        <RevealItem>
          <SectionCard title="Strategy Info" subtitle="Name, description, risk profile">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
              <div className="space-y-1.5 md:col-span-2">
                <FieldLabel>Name</FieldLabel>
                <input
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="My RSI Reversal"
                  className="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-sm text-white placeholder:text-white/30 focus:outline-none focus:border-emerald-500/50 focus:ring-1 focus:ring-emerald-500/30 transition-all"
                />
              </div>

              <div className="space-y-1.5 md:col-span-2">
                <FieldLabel>Description</FieldLabel>
                <textarea
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  placeholder="What does this strategy do and when does it work best?"
                  rows={2}
                  className="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-sm text-white placeholder:text-white/30 focus:outline-none focus:border-emerald-500/50 focus:ring-1 focus:ring-emerald-500/30 transition-all resize-none"
                />
              </div>

              <div className="space-y-1.5">
                <FieldLabel>Risk Level</FieldLabel>
                <div className="grid grid-cols-3 gap-2">
                  {(["low", "medium", "high"] as RiskLevel[]).map((r) => {
                    const variant = r === "low" ? "emerald" : r === "medium" ? "amber" : "red";
                    const hoverBorder = r === "low" ? "hover:border-emerald-500/40" : r === "medium" ? "hover:border-amber-500/40" : "hover:border-red-500/40";
                    const selectedCls = risk === r
                      ? r === "low"
                        ? "bg-emerald-500/20 text-emerald-400 border-emerald-500/30"
                        : r === "medium"
                          ? "bg-amber-500/20 text-amber-400 border-amber-500/30"
                          : "bg-red-500/20 text-red-400 border-red-500/30"
                      : `bg-white/5 text-white/40 border-white/10 hover:bg-white/10 ${hoverBorder}`;
                    return (
                      <PixelHover key={r} variant={variant} active={risk === r} className={`rounded-xl border transition-all ${selectedCls}`}>
                        <button
                          onClick={() => setRisk(r)}
                          className="w-full py-2.5 rounded-xl text-[10px] font-bold uppercase tracking-wider bg-transparent"
                        >
                          <ShieldCheck size={12} className="inline mr-1 -mt-0.5" />
                          {r}
                        </button>
                      </PixelHover>
                    );
                  })}
                </div>
              </div>

              <div className="space-y-1.5">
                <FieldLabel>Tags (comma-separated)</FieldLabel>
                <input
                  value={tagsInput}
                  onChange={(e) => setTagsInput(e.target.value)}
                  placeholder="Mean Reversion, Intraday"
                  className="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-sm text-white placeholder:text-white/30 focus:outline-none focus:border-emerald-500/50 focus:ring-1 focus:ring-emerald-500/30 transition-all"
                />
              </div>
            </div>
          </SectionCard>
        </RevealItem>

        <RevealItem>
          <SectionCard
            title="Buy When"
            subtitle="Open a long position when these rules evaluate true"
            accent="emerald"
          >
            <RuleGroupEditor
              group={entry}
              onChange={setEntry}
              accent="emerald"
              emptyBuilder={newRule}
            />
          </SectionCard>
        </RevealItem>

        <RevealItem>
          <SectionCard
            title="Sell When"
            subtitle="Close the position when any rule here is triggered"
            accent="red"
          >
            <RuleGroupEditor
              group={exit}
              onChange={setExit}
              accent="red"
              emptyBuilder={newExitRule}
            />
          </SectionCard>
        </RevealItem>

        <RevealItem>
          <SectionCard title="Risk Management" subtitle="Protective exits applied to every trade">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
              <div className="space-y-1.5">
                <FieldLabel>Stop Loss (%)</FieldLabel>
                <input
                  type="number"
                  value={stopLossPct}
                  onChange={(e) => setStopLossPct(e.target.value)}
                  placeholder="2"
                  min="0"
                  step="0.1"
                  className="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-sm text-white placeholder:text-white/30 focus:outline-none focus:border-emerald-500/50 focus:ring-1 focus:ring-emerald-500/30 transition-all"
                />
              </div>
              <div className="space-y-1.5">
                <FieldLabel>Take Profit (%)</FieldLabel>
                <input
                  type="number"
                  value={takeProfitPct}
                  onChange={(e) => setTakeProfitPct(e.target.value)}
                  placeholder="4"
                  min="0"
                  step="0.1"
                  className="w-full bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-sm text-white placeholder:text-white/30 focus:outline-none focus:border-emerald-500/50 focus:ring-1 focus:ring-emerald-500/30 transition-all"
                />
              </div>
            </div>
          </SectionCard>
        </RevealItem>

        <RevealItem>
          <SectionCard title="Preview" subtitle="Plain-English reading of your strategy">
            <div className="space-y-4 text-sm">
              <div className="flex items-start gap-3">
                <div className="p-1.5 rounded-md bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 mt-0.5 shrink-0">
                  <Eye size={13} />
                </div>
                <div className="flex-1 space-y-1">
                  <p className="text-[10px] font-semibold text-emerald-400 uppercase tracking-widest">
                    Buy when
                  </p>
                  <p className="text-white/80 leading-relaxed font-medium wrap-break-word">
                    {formatGroup(entry)}
                  </p>
                </div>
              </div>

              <div className="flex items-start gap-3">
                <div className="p-1.5 rounded-md bg-red-500/10 border border-red-500/20 text-red-400 mt-0.5 shrink-0">
                  <Eye size={13} />
                </div>
                <div className="flex-1 space-y-1">
                  <p className="text-[10px] font-semibold text-red-400 uppercase tracking-widest">
                    Sell when
                  </p>
                  <p className="text-white/80 leading-relaxed font-medium wrap-break-word">
                    {formatGroup(exit)}
                  </p>
                </div>
              </div>

              <div className="flex items-start gap-3">
                <div className="p-1.5 rounded-md bg-white/5 border border-white/10 text-white/50 mt-0.5 shrink-0">
                  <Sparkles size={13} />
                </div>
                <div className="flex-1">
                  <p className="text-[10px] font-semibold text-white/40 uppercase tracking-widest mb-1">
                    Protective exits
                  </p>
                  <p className="text-white/80 leading-relaxed font-medium">
                    Stop loss at {stopLossPct || 0}% · Take profit at{" "}
                    {takeProfitPct || 0}%
                  </p>
                </div>
              </div>
            </div>
          </SectionCard>
        </RevealItem>

        {error && (
          <motion.div
            initial={{ opacity: 0, y: -6 }}
            animate={{ opacity: 1, y: 0 }}
            className="p-4 bg-red-500/10 border border-red-500/20 rounded-xl text-red-400 text-sm font-medium flex items-center gap-2"
          >
            <AlertTriangle size={15} /> {error}
          </motion.div>
        )}

        <RevealItem className="flex items-center gap-3 pt-2">
          <span className={`inline-flex items-center gap-1 px-2.5 py-1 rounded-md text-[10px] font-bold uppercase tracking-wider border ${riskColor}`}>
            <ShieldCheck size={10} />
            {risk}
          </span>

          <div className="ml-auto flex items-center gap-3">
            <Link
              href="/strategies"
              className="px-5 py-3 rounded-xl text-xs font-bold uppercase tracking-wider text-white/50 border border-white/10 hover:bg-white/5 transition-all"
            >
              Cancel
            </Link>
            <PixelHover
              variant="emerald"
              className="group rounded-xl border border-white/10 bg-white/5 hover:bg-white/10 hover:border-emerald-500/40 transition-all active:scale-[0.98]"
            >
              <button
                onClick={handleSave}
                disabled={saving}
                className="inline-flex items-center gap-2 px-6 py-3 rounded-xl text-xs font-bold uppercase tracking-wider text-white bg-transparent transition-colors disabled:opacity-50"
              >
                {saving ? "Saving..." : "Save Strategy"}
              </button>
            </PixelHover>
          </div>
        </RevealItem>
      </RevealStagger>
    </PageEnter>
  );
}
