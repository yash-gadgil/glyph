'use client';

import { useEffect, useRef, useState } from "react";
import { usePathname } from "next/navigation";
import { motion, AnimatePresence } from "motion/react";
import { Sparkles, Send, X, Square, MessageSquare } from "lucide-react";
import PixelHover from "@/components/ui/PixelHover";
import {
  getChatSession,
  pollChatSession,
  streamChat,
  type ChatTurn,
} from "@/services/advisor/chat";

const ACCENT = "#5600a2";

const VISIBLE_ON = ["/dashboard", "/portfolio", "/explore", "/watchlist", "/strategies", "/orders"];

export default function ChatWidget() {
  const [open, setOpen] = useState(false);
  const [messages, setMessages] = useState<ChatTurn[]>([]);
  const [input, setInput] = useState("");
  const [streaming, setStreaming] = useState(false);
  const [streamingText, setStreamingText] = useState("");
  const [error, setError] = useState("");

  const accumRef = useRef("");
  const abortRef = useRef<AbortController | null>(null);
  const scrollRef = useRef<HTMLDivElement>(null);

  const pathname = usePathname();
  const visible = VISIBLE_ON.some((p) => pathname === p || pathname.startsWith(p + "/"));

  useEffect(() => {
    let active = true;
    getChatSession().then((session) => {
      if (!active) return;
      setMessages(session.turns ?? []);
      if (session.in_flight) {
        setStreaming(true);
        setStreamingText(session.partial_text ?? "");
        const controller = new AbortController();
        abortRef.current = controller;
        pollChatSession(controller.signal).then((final) => {
          if (!active) return;
          if (final.turns.length > 0) setMessages(final.turns);
          setStreaming(false);
          setStreamingText("");
        });
      }
    });
    return () => {
      active = false;
      abortRef.current?.abort();
    };
  }, []);

  useEffect(() => {
    scrollRef.current?.scrollTo({ top: scrollRef.current.scrollHeight, behavior: "smooth" });
  }, [messages, streamingText, open]);

  function send() {
    const message = input.trim();
    if (!message || streaming) return;
    setError("");
    setInput("");
    setMessages((prev) => [...prev, { role: "user", content: message }]);
    setStreaming(true);
    setStreamingText("");
    accumRef.current = "";

    const controller = new AbortController();
    abortRef.current = controller;

    streamChat(message, {
      signal: controller.signal,
      onToken: (token) => {
        accumRef.current += token;
        setStreamingText(accumRef.current);
      },
      onDone: () => {
        const text = accumRef.current.trim();
        if (text) setMessages((prev) => [...prev, { role: "assistant", content: text }]);
        setStreaming(false);
        setStreamingText("");
      },
      onError: (kind) => {
        setStreaming(false);
        setStreamingText("");
        setError(
          kind === "busy"
            ? "Still answering your previous message. One moment."
            : "Something went wrong. Try again."
        );
      },
    });
  }

  function stop() {
    abortRef.current?.abort();
    const text = accumRef.current.trim();
    if (text) setMessages((prev) => [...prev, { role: "assistant", content: text }]);
    setStreaming(false);
    setStreamingText("");
  }

  function onKeyDown(e: React.KeyboardEvent<HTMLInputElement>) {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      send();
    }
  }

  const empty = messages.length === 0 && !streaming;

  if (!visible) return null;

  return (
    <>
      <AnimatePresence>
        {open && (
          <motion.div
            initial={{ opacity: 0, y: 20, scale: 0.96 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, y: 20, scale: 0.96 }}
            transition={{ duration: 0.25, ease: [0.22, 1, 0.36, 1] }}
            className="fixed bottom-24 right-6 z-50 flex w-[min(92vw,380px)] h-[min(70vh,560px)] flex-col rounded-2xl border border-white/10 bg-neutral-950/95 backdrop-blur-md shadow-2xl overflow-hidden font-mono"
          >
            <div className="flex items-center justify-between gap-3 px-4 py-3 border-b border-white/10">
              <div className="flex items-center gap-2">
                <Sparkles size={16} style={{ color: ACCENT }} />
                <span className="text-sm font-bold text-white tracking-tight">Glyph Assistant</span>
              </div>
              <button
                onClick={() => setOpen(false)}
                className="p-1.5 rounded-lg text-white/40 hover:text-white hover:bg-white/10 transition-colors"
              >
                <X size={16} />
              </button>
            </div>

            <div ref={scrollRef} className="flex-1 overflow-y-auto px-4 py-4 space-y-3">
              {empty && (
                <div className="h-full flex flex-col items-center justify-center text-center gap-2 text-neutral-500 text-xs px-4">
                  <MessageSquare size={22} className="text-neutral-600" />
                  <p>Ask about your portfolio, a stock price, or have me generate a strategy for a ticker.</p>
                </div>
              )}

              {messages.map((m, i) => (
                <Bubble key={i} role={m.role} content={m.content} />
              ))}

              {streaming && (
                <Bubble role="assistant" content={streamingText} pending />
              )}

              {error && (
                <div className="text-xs text-red-400 px-1">{error}</div>
              )}
            </div>

            <div className="border-t border-white/10 p-3">
              <div className="flex items-center gap-2">
                <input
                  value={input}
                  onChange={(e) => setInput(e.target.value)}
                  onKeyDown={onKeyDown}
                  disabled={streaming}
                  placeholder={streaming ? "Answering…" : "Ask the assistant…"}
                  className="flex-1 bg-white/5 border border-white/10 rounded-xl px-3 py-2.5 text-sm text-white placeholder:text-white/30 focus:outline-none focus:border-[#5600a2]/60 focus:ring-1 focus:ring-[#5600a2]/40 transition-all disabled:opacity-60"
                />
                {streaming ? (
                  <button
                    onClick={stop}
                    className="p-2.5 rounded-xl border border-white/10 bg-white/5 text-white/60 hover:text-white hover:bg-white/10 transition-colors"
                    title="Stop"
                  >
                    <Square size={16} />
                  </button>
                ) : (
                  <button
                    onClick={send}
                    disabled={!input.trim()}
                    className="p-2.5 rounded-xl border border-white/10 bg-white/5 text-white hover:bg-white/10 hover:border-[#5600a2]/60 transition-colors disabled:opacity-40"
                    title="Send"
                  >
                    <Send size={16} />
                  </button>
                )}
              </div>
            </div>
          </motion.div>
        )}
      </AnimatePresence>

      <motion.div
        initial={{ opacity: 0, scale: 0.9 }}
        animate={{ opacity: 1, scale: 1 }}
        whileTap={{ scale: 0.94 }}
        className="fixed bottom-6 right-6 z-50"
      >
        <PixelHover
          gap={3}
          speed={40}
          colors="#a06cd5,#7d34c4,#5600a2"
          active={streaming}
          className="rounded-full border border-white/15 bg-white/5 backdrop-blur-md shadow-xl shadow-black/30"
        >
          <button
            onClick={() => setOpen((v) => !v)}
            aria-label="Open assistant"
            className="flex h-14 w-14 items-center justify-center rounded-full text-white transition-colors hover:text-[#c9a6f0]"
          >
            {open ? <X size={22} /> : <Sparkles size={22} />}
          </button>
        </PixelHover>
      </motion.div>
    </>
  );
}

function Bubble({ role, content, pending }: { role: "user" | "assistant"; content: string; pending?: boolean }) {
  const isUser = role === "user";
  return (
    <div className={`flex ${isUser ? "justify-end" : "justify-start"}`}>
      <div
        className={`max-w-[85%] rounded-2xl px-3.5 py-2.5 text-sm leading-relaxed whitespace-pre-wrap wrap-break-word ${
          isUser
            ? "bg-white/10 text-white border border-white/10"
            : "bg-[#5600a2]/10 text-neutral-200 border border-[#5600a2]/20"
        }`}
      >
        {content || (pending ? <span className="text-neutral-500">Thinking…</span> : null)}
        {pending && content && (
          <motion.span
            className="inline-block w-1.5 h-3.5 ml-0.5 -mb-0.5 align-middle"
            style={{ backgroundColor: ACCENT }}
            animate={{ opacity: [1, 0.2, 1] }}
            transition={{ duration: 1, repeat: Infinity }}
          />
        )}
      </div>
    </div>
  );
}
