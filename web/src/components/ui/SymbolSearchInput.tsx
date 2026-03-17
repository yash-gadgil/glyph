"use client";

import { useState, useRef, useEffect } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { Plus, Minus } from "lucide-react";
import { searchSymbols } from "@/services/watchlists/queries";
import { modifyWatchlist } from "@/services/watchlists/mutations";
import GlassSurface from "../primitives/GlassSurface";
import { motion, AnimatePresence } from "motion/react";

interface SymbolSearchInputProps {
  watchlistId: string;
  currentSymbols: string[];
}

function useDebounce(value: string, delay: number) {
  const [debounced, setDebounced] = useState(value);
  useEffect(() => {
    const t = setTimeout(() => setDebounced(value), delay);
    return () => clearTimeout(t);
  }, [value, delay]);
  return debounced;
}

export default function SymbolSearchInput({
  watchlistId,
  currentSymbols
}: SymbolSearchInputProps) {

  const [input, setInput] = useState("");
  const [open, setOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);
  const query = useDebounce(input, 300);

  const { data, isFetching } = searchSymbols(query);
  const modify = modifyWatchlist();
  const qc = useQueryClient();

  const symbolSet = new Set(currentSymbols);

  function invalidateWatchlist() {
    qc.invalidateQueries({ queryKey: ["watchlists"] });
    qc.invalidateQueries({ queryKey: ["watchlist", watchlistId] });
  }

  function patchCachedSymbols(mutate: (symbols: string[]) => string[]) {
    qc.setQueryData(
      ["watchlist", watchlistId],
      (old: { symbols?: string[] } | undefined) =>
        old ? { ...old, symbols: mutate(old.symbols ?? []) } : old
    );
  }

  function add(symbol: string) {
    if (!watchlistId) return;
    patchCachedSymbols((symbols) =>
      symbols.includes(symbol) ? symbols : [...symbols, symbol]
    );
    modify.mutate(
      { watchlistId, action: "subscribe", symbols: [symbol] },
      { onSettled: invalidateWatchlist }
    );
  }

  function remove(symbol: string) {
    if (!watchlistId) return;
    patchCachedSymbols((symbols) => symbols.filter((s) => s !== symbol));
    modify.mutate(
      { watchlistId, action: 'unsubscribe', symbols: [symbol] },
      { onSettled: invalidateWatchlist }
    );
  }

  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClick);
    return () => document.removeEventListener("mousedown", handleClick);
  }, []);

  const results: { name: string; company_name?: string }[] = data?.symbols ?? [];

  return (
    <div ref={containerRef} className="relative pointer-events-auto text-base font-mono">

      <GlassSurface
        displace={15}
        distortionScale={-150}
        redOffset={5}
        greenOffset={15}
        blueOffset={25}
        brightness={60}
        opacity={0.8}
        mixBlendMode="overlay"
        className="w-fit h-fit py-1 px-2 pointer-events-auto"
      >

        <input
          className="focus:outline-none w-72"
          placeholder="Search symbols or companies…"
          value={input}
          onChange={(e) => {
            setInput(e.target.value);
            setOpen(true);
          }}
          onFocus={() => input.length >= 1 && setOpen(true)}
        />
      </GlassSurface>

      <AnimatePresence mode="sync">
        {open && query.length >= 1 && (
          <motion.div
            layout
            className="absolute top-full mt-1 left-0 z-50 pointer-events-auto origin-top-left isolate overflow-hidden rounded-[20px]"
            variants={{
              open: { scale: 1, y: 0, transition: { type: "spring", stiffness: 500, damping: 22, mass: 0.6 } },
              closed: { scale: 0.85, y: -8, transition: { type: "tween", duration: 0.12, ease: "easeIn" } },
            }}
            initial="closed"
            animate="open"
            exit="closed"
            transition={{ layout: { type: "spring", stiffness: 400, damping: 30 } }}
          >
            <GlassSurface
              displace={15}
              distortionScale={-150}
              redOffset={5}
              greenOffset={15}
              blueOffset={25}
              brightness={60}
              opacity={0.8}
              mixBlendMode="overlay"
              className="w-96"
            >
              <ul className="w-full max-h-60 overflow-y-auto">
                <AnimatePresence initial={false} mode="popLayout">
                  {isFetching && results.length === 0 && (
                    <motion.li
                      key="searching"
                      initial={{ opacity: 0 }}
                      animate={{ opacity: 1 }}
                      exit={{ opacity: 0 }}
                      transition={{ duration: 0.2 }}
                      className="px-3 py-2 text-xs text-white/50"
                    >
                      Searching…
                    </motion.li>
                  )}
                  {!isFetching && results.length === 0 && (
                    <motion.li
                      key="no-results"
                      initial={{ opacity: 0 }}
                      animate={{ opacity: 1 }}
                      exit={{ opacity: 0 }}
                      transition={{ duration: 0.2 }}
                      className="px-3 py-2 text-xs text-white/50"
                    >
                      No results
                    </motion.li>
                  )}
                  {!isFetching && results.map(({ name, company_name }) => {
                    const inList = symbolSet.has(name);
                    return (
                      <motion.li
                        key={name}
                        initial={{ opacity: 0 }}
                        animate={{ opacity: 1 }}
                        exit={{ opacity: 0 }}
                        transition={{ opacity: { duration: 0.2 } }}
                        className="flex items-center justify-between px-3 py-1.5 text-sm gap-2"
                      >
                        <div className="flex flex-col min-w-0">
                          <span className="font-semibold uppercase">{name}</span>
                          {company_name && (
                            <span className="text-[11px] text-white/50 truncate leading-tight">{company_name}</span>
                          )}
                        </div>
                        <button
                          onClick={() => (inList ? remove(name) : add(name))}
                          className="ml-2 p-0.5 rounded hover:bg-white/20 transition-colors shrink-0"
                          title={inList ? "Remove from watchlist" : "Add to watchlist"}
                        >
                          {inList ? <Minus size={14} /> : <Plus size={14} />}
                        </button>
                      </motion.li>
                    );
                  })}
                </AnimatePresence>
              </ul>
            </GlassSurface>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}
