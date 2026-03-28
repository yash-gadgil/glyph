import { useCallback, useEffect, useRef, useState } from "react";
import { API_BASE_URL } from "@/lib/api";
import { isMockMode } from "@/lib/mock";

export type AdvisorStatus = "idle" | "streaming" | "done" | "error";

const MOCK_ANALYSIS = `Your book is sitting on a large cash position with a single equity holding in AAPL, which is carrying a healthy unrealized gain.

The main thing to watch is concentration: nearly all of your invested capital is in one name, so a move in AAPL drives your whole return. Holding that much cash also means a chunk of your balance is not working for you.

If you want to stay long but cut single-name risk, consider trimming AAPL into strength and spreading into a couple of uncorrelated tickers. A book like this tends to suit trend-following entries with a fixed stop, or a simple mean-reversion rule on the names you already follow.

This is a simulated account for educational use, not financial advice.`;

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

export function useAnalyzePortfolio() {
  const [text, setText] = useState("");
  const [status, setStatus] = useState<AdvisorStatus>("idle");
  const abortRef = useRef<AbortController | null>(null);

  useEffect(() => () => abortRef.current?.abort(), []);

  const start = useCallback(async () => {
    abortRef.current?.abort();
    setText("");
    setStatus("streaming");

    if (isMockMode()) {
      for (const word of MOCK_ANALYSIS.split(" ")) {
        await new Promise((r) => setTimeout(r, 18));
        setText((prev) => prev + word + " ");
      }
      setStatus("done");
      return;
    }

    const controller = new AbortController();
    abortRef.current = controller;

    try {
      const res = await fetch(`${API_BASE_URL}/advisor/analyze`, {
        credentials: "include",
        headers: { Accept: "text/event-stream" },
        signal: controller.signal,
      });

      if (!res.ok || !res.body) {
        setStatus("error");
        return;
      }

      const reader = res.body.getReader();
      const decoder = new TextDecoder();
      let buffer = "";

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const blocks = buffer.split("\n\n");
        buffer = blocks.pop() ?? "";

        for (const block of blocks) {
          if (!block.trim()) continue;
          const { name, data } = parseEvent(block);
          if (name === "error") {
            setStatus("error");
            await reader.cancel();
            return;
          }
          if (name === "done") {
            setStatus("done");
            await reader.cancel();
            return;
          }
          if (data) {
            setText((prev) => prev + data.replace(/\\n/g, "\n"));
          }
        }
      }

      setStatus("done");
    } catch (err) {
      if ((err as Error).name !== "AbortError") {
        setStatus("error");
      }
    }
  }, []);

  return { text, status, start };
}
