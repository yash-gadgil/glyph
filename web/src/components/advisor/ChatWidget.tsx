'use client';

import { useEffect, useRef, useState } from "react";
import { usePathname } from "next/navigation";
import { motion, AnimatePresence } from "motion/react";
import { Send, X, Square, MessageSquare, Trash2 } from "lucide-react";
import PixelHover from "@/components/ui/PixelHover";
import { TextShimmer } from "@/components/ui/TextShimmer";
import KenazIcon from "@/components/advisor/KenazIcon";
import {
  getChatSession,
  pollChatSession,
  streamChat,
  clearChatSession,
  type ChatTurn,
  type ChatProvider,
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
  const [provider, setProvider] = useState<ChatProvider>("gemini");

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
    function onClear() {
      abortRef.current?.abort();
      accumRef.current = "";
      setMessages([]);
      setStreaming(false);
      setStreamingText("");
      setError("");
      setOpen(false);
    }
    window.addEventListener("kenaz:clear", onClear);
    return () => window.removeEventListener("kenaz:clear", onClear);
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

    streamChat(message, provider, {
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

  function clearChat() {
    abortRef.current?.abort();
    accumRef.current = "";
    setMessages([]);
    setStreaming(false);
    setStreamingText("");
    setError("");
    clearChatSession();
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
            className="fixed bottom-24 right-6 z-50 flex w-[min(92vw,380px)] h-[min(70vh,560px)] flex-col rounded-2xl border border-white/10 bg-neutral-950/55 backdrop-blur-2xl shadow-2xl shadow-black/40 overflow-hidden font-mono"
          >
            <div className="flex items-center justify-between gap-2 px-4 py-3 border-b border-white/10 bg-white/[0.03]">
              <div className="flex items-center gap-2">
                <span style={{ color: ACCENT }}><KenazIcon size={16} /></span>
                <span className="text-sm font-bold text-white tracking-tight">Kenaz</span>
              </div>
              <div className="flex items-center gap-1.5">
                <div className="relative">
                  <select
                    value={provider}
                    onChange={(e) => setProvider(e.target.value as ChatProvider)}
                    className="appearance-none rounded-lg border border-white/10 bg-white/5 py-1 pl-2 pr-6 text-[10px] font-semibold uppercase tracking-wider text-white/70 hover:text-white focus:outline-none focus:border-[#5600a2]/60 transition-colors cursor-pointer"
                    title="Model"
                  >
                    <option value="gemini" className="bg-neutral-900">Gemini</option>
                    <option value="inference" className="bg-neutral-900">Inference</option>
                  </select>
                  <span className="pointer-events-none absolute right-1.5 top-1/2 -translate-y-1/2 text-white/40 text-[8px]">▼</span>
                </div>
                <button
                  onClick={clearChat}
                  className="p-1.5 rounded-lg text-white/40 hover:text-white hover:bg-white/10 transition-colors"
                  title="Clear chat"
                >
                  <Trash2 size={15} />
                </button>
                <button
                  onClick={() => setOpen(false)}
                  className="p-1.5 rounded-lg text-white/40 hover:text-white hover:bg-white/10 transition-colors"
                >
                  <X size={16} />
                </button>
              </div>
            </div>

            <div ref={scrollRef} className="flex-1 overflow-y-auto px-4 py-4 space-y-3">
              {empty && (
                <div className="h-full flex flex-col items-center justify-center text-center gap-2 text-neutral-500 text-xs px-4">
                  <MessageSquare size={22} className="text-neutral-600" />
                  <p>Ask about your portfolio, a stock, the market today, or have me generate a strategy.</p>
                </div>
              )}

              {messages.map((m, i) => (
                <Bubble key={i} role={m.role} content={m.content} />
              ))}

              {streaming && <Bubble role="assistant" content={streamingText} pending />}

              {error && <div className="text-xs text-red-400 px-1">{error}</div>}
            </div>

            <div className="border-t border-white/10 p-3 bg-white/[0.03]">
              <div className="flex items-center gap-2">
                <input
                  value={input}
                  onChange={(e) => setInput(e.target.value)}
                  onKeyDown={onKeyDown}
                  disabled={streaming}
                  placeholder={streaming ? "Answering…" : "Ask Kenaz…"}
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
            aria-label="Open Kenaz"
            className="flex h-14 w-14 items-center justify-center rounded-full text-white transition-colors hover:text-[#c9a6f0]"
          >
            {open ? <X size={22} /> : <KenazIcon size={24} />}
          </button>
        </PixelHover>
      </motion.div>
    </>
  );
}

function stripMarkdown(s: string): string {
  return s
    .replace(/`([^`]+)`/g, "$1")
    .replace(/`+/g, "")
    .replace(/\*\*([^*]+)\*\*/g, "$1")
    .replace(/__([^_]+)__/g, "$1")
    .replace(/\*([^*]+)\*/g, "$1")
    .replace(/^#{1,6}\s+/gm, "")
    .replace(/^\s*[-*+]\s+/gm, "• ")
    .replace(/^\s*>\s?/gm, "");
}

function Bubble({ role, content, pending }: { role: "user" | "assistant"; content: string; pending?: boolean }) {
  const isUser = role === "user";
  const text = isUser ? content : stripMarkdown(content);
  return (
    <div className={`flex ${isUser ? "justify-end" : "justify-start"}`}>
      <div
        className={`max-w-[85%] rounded-2xl px-3.5 py-2.5 text-sm leading-relaxed whitespace-pre-wrap wrap-break-word ${
          isUser
            ? "bg-white/10 text-white border border-white/10"
            : "bg-[#5600a2]/10 text-neutral-200 border border-[#5600a2]/20"
        }`}
      >
        {text ? text : pending ? <TextShimmer as="span" className="text-sm">Thinking…</TextShimmer> : null}
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
