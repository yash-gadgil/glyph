import { isMockMode, MockSocket } from "./mock";

export const WS_BASE_URL =
  process.env.NEXT_PUBLIC_WS_URL ?? "ws://localhost:8080";

export const socketForWatchlist = (watchlist_id: string) => {
  if (isMockMode()) {
    return new MockSocket(watchlist_id) as unknown as WebSocket;
  }
  return new ReconnectingSocket(
    () => new WebSocket(`${WS_BASE_URL}/watchlists/${watchlist_id}`),
  ) as unknown as WebSocket;
};

export const socketForSymbols = (symbols: string[]) => {
  if (isMockMode()) {
    return new MockSocket(symbols.join(",")) as unknown as WebSocket;
  }
  const qs = encodeURIComponent(symbols.join(","));
  return new ReconnectingSocket(
    () => new WebSocket(`${WS_BASE_URL}/watchlists/stream?symbols=${qs}`),
  ) as unknown as WebSocket;
};

type Listener = (event: any) => void;

export type ReconnectingSocketStatus = "connecting" | "open" | "closed";

export class ReconnectingSocket {
  onmessage: Listener | null = null;
  onopen: Listener | null = null;
  onclose: Listener | null = null;
  onerror: Listener | null = null;
  onstatus: ((status: ReconnectingSocketStatus) => void) | null = null;

  private socket: WebSocket | null = null;
  private attempts = 0;
  private closedByUser = false;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private listeners: Record<string, Listener[]> = {};

  constructor(
    private readonly makeSocket: () => WebSocket,
    private readonly maxBackoffMs = 15_000,
  ) {
    this.connect();
  }

  private connect() {
    this.onstatus?.("connecting");
    const socket = this.makeSocket();
    this.socket = socket;

    socket.onopen = (event) => {
      this.attempts = 0;
      this.onstatus?.("open");
      this.onopen?.(event);
      this.dispatch("open", event);
    };
    socket.onmessage = (event) => {
      this.onmessage?.(event);
      this.dispatch("message", event);
    };
    socket.onerror = (event) => {
      this.onerror?.(event);
      this.dispatch("error", event);
    };
    socket.onclose = (event) => {
      this.onclose?.(event);
      this.dispatch("close", event);
      if (this.closedByUser) {
        this.onstatus?.("closed");
        return;
      }
      this.scheduleReconnect();
    };
  }

  addEventListener(type: string, fn: Listener) {
    (this.listeners[type] ||= []).push(fn);
  }

  removeEventListener(type: string, fn: Listener) {
    this.listeners[type] = (this.listeners[type] || []).filter((l) => l !== fn);
  }

  private dispatch(type: string, event: any) {
    for (const fn of this.listeners[type] || []) fn(event);
  }

    private scheduleReconnect() {
    this.onstatus?.("connecting");
    const backoff = Math.min(1000 * 2 ** this.attempts, this.maxBackoffMs);
    this.attempts += 1;
    this.reconnectTimer = setTimeout(() => this.connect(), backoff);
  }

  get readyState(): number {
    return this.socket?.readyState ?? WebSocket.CONNECTING;
  }

  send(data: string) {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      this.socket.send(data);
    }
  }

  close() {
    this.closedByUser = true;
    if (this.reconnectTimer) clearTimeout(this.reconnectTimer);
    this.socket?.close();
  }
}
