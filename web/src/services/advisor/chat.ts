import { API_BASE_URL } from "@/lib/api";
import { isMockMode } from "@/lib/mock";

export type ChatRole = "user" | "assistant";
export type ChatTurn = { role: ChatRole; content: string };
export type ChatSession = { turns: ChatTurn[]; in_flight: boolean; partial_text: string };

export type StreamEvent = "error" | "busy" | "network";

type StreamHandlers = {
  onToken: (token: string) => void;
  onDone: () => void;
  onError: (kind: StreamEvent) => void;
  signal?: AbortSignal;
};

const EMPTY_SESSION: ChatSession = { turns: [], in_flight: false, partial_text: "" };

const MOCK_REPLY =
  "Your book is concentrated in a single name with a large cash balance. If you want to cut single-name risk you could trim into strength and spread into a couple of uncorrelated tickers. This is a simulated account for educational use, not financial advice.";

function parseEvent(block: string): { name: string; data: string } {
  let name = "message";
  let data = "";
  for (const line of block.split("\n")) {
    if (line.startsWith("event:")) {
      name = line.slice(6).trim();
    } else if (line.startsWith("data:")) {
      let d = line.slice(5);
      if (d.startsWith(" ")) d = d.slice(1);
      data += d;
    }
  }
  return { name, data };
}

export async function getChatSession(): Promise<ChatSession> {
  if (isMockMode()) {
    return EMPTY_SESSION;
  }
  try {
    const res = await fetch(`${API_BASE_URL}/advisor/chat/session`, { credentials: "include" });
    if (!res.ok) {
      return EMPTY_SESSION;
    }
    return res.json();
  } catch {
    return EMPTY_SESSION;
  }
}

export async function pollChatSession(signal?: AbortSignal): Promise<ChatSession> {
  const deadline = Date.now() + 120000;
  while (Date.now() < deadline) {
    if (signal?.aborted) {
      return EMPTY_SESSION;
    }
    await new Promise((r) => setTimeout(r, 1500));
    const session = await getChatSession();
    if (!session.in_flight) {
      return session;
    }
  }
  return EMPTY_SESSION;
}

export async function streamChat(message: string, handlers: StreamHandlers): Promise<void> {
  if (isMockMode()) {
    for (const word of MOCK_REPLY.split(" ")) {
      if (handlers.signal?.aborted) break;
      await new Promise((r) => setTimeout(r, 25));
      handlers.onToken(word + " ");
    }
    handlers.onDone();
    return;
  }

  let res: Response;
  try {
    res = await fetch(`${API_BASE_URL}/advisor/chat`, {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json", Accept: "text/event-stream" },
      body: JSON.stringify({ message }),
      signal: handlers.signal,
    });
  } catch (err) {
    if ((err as Error).name !== "AbortError") {
      handlers.onError("network");
    }
    return;
  }

  if (!res.ok || !res.body) {
    handlers.onError("network");
    return;
  }

  const reader = res.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";

  try {
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;

      buffer += decoder.decode(value, { stream: true });
      const blocks = buffer.split("\n\n");
      buffer = blocks.pop() ?? "";

      for (const block of blocks) {
        if (!block.trim()) continue;
        const { name, data } = parseEvent(block);
        if (name === "busy") {
          handlers.onError("busy");
          await reader.cancel();
          return;
        }
        if (name === "error") {
          handlers.onError("error");
          await reader.cancel();
          return;
        }
        if (name === "done") {
          handlers.onDone();
          await reader.cancel();
          return;
        }
        if (data) {
          handlers.onToken(data.replace(/\\n/g, "\n"));
        }
      }
    }
    handlers.onDone();
  } catch (err) {
    if ((err as Error).name !== "AbortError") {
      handlers.onError("network");
    }
  }
}
