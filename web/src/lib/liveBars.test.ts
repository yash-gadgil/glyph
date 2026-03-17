import { describe, it, expect } from "vitest";
import { mergeLiveBar } from "./liveBars";
import type { RealBar } from "./marketData";

function bar(time: string, o: number, h: number, l: number, c: number, v = 0): RealBar {
  return { time, open: o, high: h, low: l, close: c, volume: v };
}

function tick(time: string, price: number, v = 0): RealBar {
  return bar(time, price, price, price, price, v);
}

describe("mergeLiveBar", () => {
  it("merges a tick into the current minute bucket (1D)", () => {
    const bars = [bar("2026-01-02T15:00:00Z", 100, 101, 99, 100.5, 1000)];
    const next = mergeLiveBar(bars, tick("2026-01-02T15:00:30Z", 102, 5), "1D");

    expect(next).toHaveLength(1);
    expect(next[0].open).toBe(100);
    expect(next[0].high).toBe(102);
    expect(next[0].low).toBe(99);
    expect(next[0].close).toBe(102);
    expect(next[0].volume).toBe(1005);
  });

  it("appends a new bar when the tick opens a newer minute", () => {
    const bars = [bar("2026-01-02T15:00:00Z", 100, 101, 99, 100.5, 1000)];
    const next = mergeLiveBar(bars, tick("2026-01-02T15:01:10Z", 103, 7), "1D");

    expect(next).toHaveLength(2);
    expect(next[1].time).toBe("2026-01-02T15:01:00Z");
    expect(next[1].close).toBe(103);
    expect(next[1].volume).toBe(7);
  });

  it("lets a real bar replace volume in the same bucket", () => {
    const bars = [bar("2026-01-02T15:00:00Z", 100, 101, 99, 100.5, 1000)];
    const real = bar("2026-01-02T15:00:00Z", 100, 104, 98, 101, 2500);
    const next = mergeLiveBar(bars, real, "1D");

    expect(next).toHaveLength(1);
    expect(next[0].high).toBe(104);
    expect(next[0].low).toBe(98);
    expect(next[0].volume).toBe(2500);
  });

  it("ignores an out-of-order (older) bucket", () => {
    const bars = [bar("2026-01-02T15:05:00Z", 100, 101, 99, 100.5, 1000)];
    const next = mergeLiveBar(bars, tick("2026-01-02T15:02:00Z", 200), "1D");
    expect(next).toBe(bars);
  });

  it("buckets by hour for 1W/1M", () => {
    const bars = [bar("2026-01-02T15:00:00Z", 100, 101, 99, 100.5, 1000)];
    const same = mergeLiveBar(bars, tick("2026-01-02T15:45:00Z", 102), "1W");
    expect(same).toHaveLength(1);
    expect(same[0].close).toBe(102);
    const nextHour = mergeLiveBar(bars, tick("2026-01-02T16:05:00Z", 103), "1W");
    expect(nextHour).toHaveLength(2);
    expect(nextHour[1].time).toBe("2026-01-02T16:00:00Z");
  });

  it("buckets by UTC date for daily timeframes", () => {
    const bars = [bar("2026-01-02T00:00:00Z", 100, 101, 99, 100.5, 1000)];
    const same = mergeLiveBar(bars, tick("2026-01-02T19:30:00Z", 105), "3M");
    expect(same).toHaveLength(1);
    expect(same[0].close).toBe(105);
    const nextDay = mergeLiveBar(bars, tick("2026-01-03T14:00:00Z", 106), "3M");
    expect(nextDay).toHaveLength(2);
    expect(nextDay[1].time).toBe("2026-01-03T00:00:00Z");
  });

  it("seeds from an empty series", () => {
    const next = mergeLiveBar([], tick("2026-01-02T15:00:30Z", 100, 3), "1D");
    expect(next).toHaveLength(1);
    expect(next[0].time).toBe("2026-01-02T15:00:00Z");
    expect(next[0].close).toBe(100);
  });
});
