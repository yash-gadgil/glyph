
export type RiskLevel = "low" | "medium" | "high";


export type IndicatorKind =
  | "price"
  | "sma"
  | "ema"
  | "rsi"
  | "macd_line"
  | "macd_signal"
  | "macd_histogram"
  | "bbands_upper"
  | "bbands_middle"
  | "bbands_lower"
  | "atr"
  | "volume"
  | "vwap"
  | "stoch_k"
  | "stoch_d";

export type ParamSpec = {
  key: string;
  label: string;
  default: number;
  min?: number;
  max?: number;
};

export type IndicatorSpec = {
  label: string;
  category: "price" | "momentum" | "trend" | "volatility" | "volume";
  description: string;
  params: ParamSpec[];
};

export const INDICATORS: Record<IndicatorKind, IndicatorSpec> = {
  price: {
    label: "Price",
    category: "price",
    description: "Latest close price",
    params: [],
  },
  sma: {
    label: "SMA",
    category: "trend",
    description: "Simple moving average",
    params: [{ key: "period", label: "Period", default: 20, min: 2, max: 500 }],
  },
  ema: {
    label: "EMA",
    category: "trend",
    description: "Exponential moving average",
    params: [{ key: "period", label: "Period", default: 20, min: 2, max: 500 }],
  },
  rsi: {
    label: "RSI",
    category: "momentum",
    description: "Relative Strength Index (0-100)",
    params: [{ key: "period", label: "Period", default: 14, min: 2, max: 100 }],
  },
  macd_line: {
    label: "MACD Line",
    category: "momentum",
    description: "EMA(fast) − EMA(slow)",
    params: [
      { key: "fast", label: "Fast", default: 12, min: 2, max: 100 },
      { key: "slow", label: "Slow", default: 26, min: 2, max: 200 },
    ],
  },
  macd_signal: {
    label: "MACD Signal",
    category: "momentum",
    description: "EMA of MACD line",
    params: [
      { key: "fast", label: "Fast", default: 12, min: 2, max: 100 },
      { key: "slow", label: "Slow", default: 26, min: 2, max: 200 },
      { key: "signal", label: "Signal", default: 9, min: 2, max: 100 },
    ],
  },
  macd_histogram: {
    label: "MACD Hist",
    category: "momentum",
    description: "MACD Line − MACD Signal",
    params: [
      { key: "fast", label: "Fast", default: 12, min: 2, max: 100 },
      { key: "slow", label: "Slow", default: 26, min: 2, max: 200 },
      { key: "signal", label: "Signal", default: 9, min: 2, max: 100 },
    ],
  },
  bbands_upper: {
    label: "BB Upper",
    category: "volatility",
    description: "Upper Bollinger Band",
    params: [
      { key: "period", label: "Period", default: 20, min: 2, max: 200 },
      { key: "stddev", label: "StdDev", default: 2, min: 1, max: 5 },
    ],
  },
  bbands_middle: {
    label: "BB Middle",
    category: "volatility",
    description: "Middle Bollinger Band (SMA)",
    params: [{ key: "period", label: "Period", default: 20, min: 2, max: 200 }],
  },
  bbands_lower: {
    label: "BB Lower",
    category: "volatility",
    description: "Lower Bollinger Band",
    params: [
      { key: "period", label: "Period", default: 20, min: 2, max: 200 },
      { key: "stddev", label: "StdDev", default: 2, min: 1, max: 5 },
    ],
  },
  atr: {
    label: "ATR",
    category: "volatility",
    description: "Average True Range",
    params: [{ key: "period", label: "Period", default: 14, min: 2, max: 100 }],
  },
  volume: {
    label: "Volume",
    category: "volume",
    description: "Latest bar volume",
    params: [],
  },
  vwap: {
    label: "VWAP",
    category: "volume",
    description: "Volume-weighted avg price (session)",
    params: [],
  },
  stoch_k: {
    label: "Stoch %K",
    category: "momentum",
    description: "Stochastic %K (0-100)",
    params: [{ key: "period", label: "Period", default: 14, min: 2, max: 100 }],
  },
  stoch_d: {
    label: "Stoch %D",
    category: "momentum",
    description: "Stochastic %D (0-100)",
    params: [{ key: "period", label: "Period", default: 14, min: 2, max: 100 }],
  },
};

export type Scale = "price" | "oscillator" | "macd" | "volume" | "atr";

const SCALE: Record<IndicatorKind, Scale> = {
  price: "price",
  sma: "price",
  ema: "price",
  bbands_upper: "price",
  bbands_middle: "price",
  bbands_lower: "price",
  vwap: "price",
  macd_line: "macd",
  macd_signal: "macd",
  macd_histogram: "macd",
  rsi: "oscillator",
  stoch_k: "oscillator",
  stoch_d: "oscillator",
  volume: "volume",
  atr: "atr",
};

export function indicatorScale(kind: IndicatorKind): Scale {
  return SCALE[kind];
}

export function compatibleIndicators(kind: IndicatorKind): IndicatorKind[] {
  const s = SCALE[kind];
  return (Object.keys(SCALE) as IndicatorKind[]).filter((k) => SCALE[k] === s);
}

export type ValueBounds = { min?: number; max?: number; default: number };

export function valueBoundsFor(kind: IndicatorKind): ValueBounds {
  switch (SCALE[kind]) {
    case "oscillator":
      return { min: 0, max: 100, default: 50 };
    case "macd":
      return { default: 0 };
    case "volume":
      return { min: 0, default: 1_000_000 };
    case "atr":
      return { min: 0, default: 1 };
    default:
      return { min: 0, default: 100 };
  }
}


export type IndicatorRef = {
  kind: IndicatorKind;
  params: Record<string, number>;
};

export type Operator = ">" | "<" | ">=" | "<=" | "crosses_above" | "crosses_below";

export const OPERATORS: { value: Operator; label: string; verbose: string }[] = [
  { value: ">", label: ">", verbose: "is greater than" },
  { value: "<", label: "<", verbose: "is less than" },
  { value: ">=", label: "≥", verbose: "is at least" },
  { value: "<=", label: "≤", verbose: "is at most" },
  { value: "crosses_above", label: "⤴ crosses above", verbose: "crosses above" },
  { value: "crosses_below", label: "⤵ crosses below", verbose: "crosses below" },
];

export type RuleRHS =
  | { kind: "value"; value: number }
  | { kind: "indicator"; indicator: IndicatorRef };

export type Rule = {
  id: string;
  lhs: IndicatorRef;
  op: Operator;
  rhs: RuleRHS;
};

export type Combinator = "AND" | "OR";

export type RuleGroup = {
  combinator: Combinator;
  rules: Rule[];
};

export type CustomStrategy = {
  id: string;
  name: string;
  description: string;
  risk: RiskLevel;
  tags: string[];
  entry: RuleGroup;
  exit: RuleGroup;
  stopLossPct: number;
  takeProfitPct: number;
  createdAt: string;
};


function uid() {
  return Math.random().toString(36).slice(2, 10);
}

export function defaultIndicatorRef(kind: IndicatorKind): IndicatorRef {
  const spec = INDICATORS[kind];
  return {
    kind,
    params: Object.fromEntries(spec.params.map((p) => [p.key, p.default])),
  };
}

export function coerceRuleForLhs(rule: Rule): Rule {
  if (rule.rhs.kind === "indicator") {
    if (indicatorScale(rule.rhs.indicator.kind) !== indicatorScale(rule.lhs.kind)) {
      return { ...rule, rhs: { kind: "indicator", indicator: defaultIndicatorRef(rule.lhs.kind) } };
    }
    return rule;
  }
  const bounds = valueBoundsFor(rule.lhs.kind);
  let value = rule.rhs.value;
  if (bounds.min !== undefined && value < bounds.min) value = bounds.min;
  if (bounds.max !== undefined && value > bounds.max) value = bounds.max;
  return { ...rule, rhs: { kind: "value", value } };
}

export function newRule(): Rule {
  return {
    id: uid(),
    lhs: defaultIndicatorRef("rsi"),
    op: "<",
    rhs: { kind: "value", value: 30 },
  };
}

export function newExitRule(): Rule {
  return {
    id: uid(),
    lhs: defaultIndicatorRef("rsi"),
    op: ">",
    rhs: { kind: "value", value: 70 },
  };
}


export function formatIndicator(ref: IndicatorRef): string {
  const spec = INDICATORS[ref.kind];
  if (spec.params.length === 0) return spec.label;
  const paramStr = spec.params
    .map((p) => String(ref.params[p.key] ?? p.default))
    .join(",");
  return `${spec.label}(${paramStr})`;
}

export function formatRule(rule: Rule): string {
  const opEntry = OPERATORS.find((o) => o.value === rule.op);
  const opLabel = opEntry ? opEntry.label : rule.op;
  const rhs =
    rule.rhs.kind === "value"
      ? String(rule.rhs.value)
      : formatIndicator(rule.rhs.indicator);
  return `${formatIndicator(rule.lhs)} ${opLabel} ${rhs}`;
}

export function formatGroup(group: RuleGroup): string {
  if (group.rules.length === 0) return "-";
  return group.rules.map(formatRule).join(`  ${group.combinator}  `);
}


import { api } from "./api";

type ApiStrategyRow = {
  id: string;
  name: string;
  config_json: string;
};

export async function loadCustomStrategies(): Promise<CustomStrategy[]> {
  const data = await api("strategies");
  const rows: ApiStrategyRow[] = data?.strategies ?? [];
  const list: CustomStrategy[] = [];
  for (const row of rows) {
    try {
      const cs = JSON.parse(row.config_json) as CustomStrategy;
      cs.id = row.id;
      cs.name = row.name;
      list.push(cs);
    } catch {
    }
  }
  return list;
}

export async function createCustomStrategy(cs: CustomStrategy): Promise<CustomStrategy> {
  const row = await api("strategies", {
    method: "POST",
    body: JSON.stringify({ name: cs.name, config_json: cs }),
  });
  return { ...cs, id: row?.id ?? cs.id };
}

export async function deleteCustomStrategyById(id: string): Promise<void> {
  await api(`strategies/${id}`, { method: "DELETE" });
}
