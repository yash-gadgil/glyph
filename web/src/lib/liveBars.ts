
import type { RealBar, Timeframe } from "./marketData";

function bucketKey(time: string, tf: Timeframe): string {
  if (tf === "1D") return time.slice(0, 16);
  if (tf === "1W" || tf === "1M") return time.slice(0, 13);
  return time.slice(0, 10);
}

function bucketTime(time: string, tf: Timeframe): string {
  if (tf === "1D") return time.slice(0, 16) + ":00Z";
  if (tf === "1W" || tf === "1M") return time.slice(0, 13) + ":00:00Z";
  return time.slice(0, 10) + "T00:00:00Z";
}

function isTick(b: RealBar): boolean {
  return b.open === b.high && b.high === b.low && b.low === b.close;
}

export function mergeLiveBar(
  bars: RealBar[],
  update: RealBar,
  tf: Timeframe
): RealBar[] {
  if (bars.length === 0) {
    return [{ ...update, time: bucketTime(update.time, tf) }];
  }

  const last = bars[bars.length - 1];
  const lastKey = bucketKey(last.time, tf);
  const updateKey = bucketKey(update.time, tf);

  if (updateKey === lastKey) {
    const merged: RealBar = {
      ...last,
      high: Math.max(last.high, update.high),
      low: Math.min(last.low, update.low),
      close: update.close,
      volume: isTick(update) ? last.volume + update.volume : update.volume,
    };
    return [...bars.slice(0, -1), merged];
  }

  if (updateKey > lastKey) {
    return [...bars, { ...update, time: bucketTime(update.time, tf) }];
  }

  return bars;
}
