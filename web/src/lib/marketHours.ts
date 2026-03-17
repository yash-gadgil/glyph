
const ET_TZ = "America/New_York";
const OPEN_MINUTES = 9 * 60 + 30;
const CLOSE_MINUTES = 16 * 60;

const WEEKDAY_INDEX: Record<string, number> = {
  Sun: 0,
  Mon: 1,
  Tue: 2,
  Wed: 3,
  Thu: 4,
  Fri: 5,
  Sat: 6,
};

function easternParts(date: Date): { weekday: number; minutes: number } {
  const parts = new Intl.DateTimeFormat("en-US", {
    timeZone: ET_TZ,
    weekday: "short",
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
  }).formatToParts(date);

  const get = (type: string) => parts.find((p) => p.type === type)?.value ?? "";
  const weekday = WEEKDAY_INDEX[get("weekday")] ?? 0;
  const hour = parseInt(get("hour"), 10) % 24;
  const minute = parseInt(get("minute"), 10);
  return { weekday, minutes: hour * 60 + minute };
}

export function isUSMarketOpen(date: Date = new Date()): boolean {
  const { weekday, minutes } = easternParts(date);
  if (weekday === 0 || weekday === 6) return false;
  return minutes >= OPEN_MINUTES && minutes < CLOSE_MINUTES;
}

function tzOffsetMinutes(date: Date, tz: string): number {
  const parts = new Intl.DateTimeFormat("en-US", {
    timeZone: tz,
    hour12: false,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  }).formatToParts(date);
  const get = (t: string) => parseInt(parts.find((p) => p.type === t)?.value ?? "0", 10);
  const asUTC = Date.UTC(get("year"), get("month") - 1, get("day"), get("hour") % 24, get("minute"), get("second"));
  return Math.round((asUTC - date.getTime()) / 60000);
}

function easternWallToInstant(y: number, mo: number, d: number, hh: number, mm: number): Date {
  const guess = Date.UTC(y, mo - 1, d, hh, mm);
  const offset = tzOffsetMinutes(new Date(guess), ET_TZ);
  return new Date(guess - offset * 60000);
}

export function nextMarketOpen(from: Date = new Date()): Date {
  const { weekday, minutes } = easternParts(from);
  const dateParts = new Intl.DateTimeFormat("en-US", {
    timeZone: ET_TZ,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).formatToParts(from);
  const get = (t: string) => parseInt(dateParts.find((p) => p.type === t)?.value ?? "0", 10);

  let add = 0;
  const openLaterToday = weekday >= 1 && weekday <= 5 && minutes < OPEN_MINUTES;
  if (!openLaterToday) {
    add = 1;
    let wd = (weekday + 1) % 7;
    while (wd === 0 || wd === 6) {
      add++;
      wd = (wd + 1) % 7;
    }
  }

  const target = new Date(Date.UTC(get("year"), get("month") - 1, get("day")));
  target.setUTCDate(target.getUTCDate() + add);
  return easternWallToInstant(
    target.getUTCFullYear(),
    target.getUTCMonth() + 1,
    target.getUTCDate(),
    9,
    30
  );
}

export function formatCountdown(ms: number): string {
  if (ms <= 0) return "now";
  const totalMin = Math.floor(ms / 60000);
  const h = Math.floor(totalMin / 60);
  const m = totalMin % 60;
  if (h >= 24) return `${Math.floor(h / 24)}d ${h % 24}h`;
  if (h > 0) return `${h}h ${m}m`;
  if (m > 0) return `${m}m`;
  return "<1m";
}

export function localOpenLabel(from: Date = new Date()): string {
  return nextMarketOpen(from).toLocaleTimeString([], { hour: "numeric", minute: "2-digit" });
}
