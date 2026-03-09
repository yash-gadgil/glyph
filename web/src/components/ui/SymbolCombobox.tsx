"use client";

import { useEffect, useLayoutEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { searchSymbols } from "@/services/watchlists/queries";
import GlassSurface from "../primitives/GlassSurface";
import { motion, AnimatePresence } from "motion/react";

interface SymbolComboboxProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  inputClassName?: string;
}

function useDebounce<T>(value: T, delay: number): T {
  const [debounced, setDebounced] = useState(value);
  useEffect(() => {
    const t = setTimeout(() => setDebounced(value), delay);
    return () => clearTimeout(t);
  }, [value, delay]);
  return debounced;
}

export default function SymbolCombobox({
  value,
  onChange,
  placeholder = "AAPL",
  inputClassName,
}: SymbolComboboxProps) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [open, setOpen] = useState(false);
  const [mounted, setMounted] = useState(false);
  const [rect, setRect] = useState<DOMRect | null>(null);

  const query = useDebounce(value.trim(), 250);
  const { data, isFetching } = searchSymbols(query);
  const results: { name: string; company_name?: string }[] = data?.symbols ?? [];

  useEffect(() => {
    setMounted(true);
    return () => setMounted(false);
  }, []);

  useLayoutEffect(() => {
    if (!open) return;
    const update = () => {
      if (inputRef.current) setRect(inputRef.current.getBoundingClientRect());
    };
    update();
    window.addEventListener("scroll", update, true);
    window.addEventListener("resize", update);
    return () => {
      window.removeEventListener("scroll", update, true);
      window.removeEventListener("resize", update);
    };
  }, [open, value]);

  function select(name: string) {
    onChange(name);
    setOpen(false);
  }

  const showDropdown = open && mounted && rect && query.length >= 1;

  return (
    <>
      <input
        ref={inputRef}
        value={value}
        placeholder={placeholder}
        className={inputClassName}
        onChange={(e) => {
          onChange(e.target.value);
          setOpen(true);
        }}
        onFocus={() => value.trim().length >= 1 && setOpen(true)}
        onBlur={() => setTimeout(() => setOpen(false), 120)}
      />

      {mounted &&
        createPortal(
          <AnimatePresence>
            {showDropdown && rect && (
              <motion.div
                style={{
                  position: "fixed",
                  top: rect.bottom + 6,
                  left: rect.left,
                  width: rect.width,
                  zIndex: 60,
                }}
                className="origin-top isolate overflow-hidden rounded-[18px] pointer-events-auto"
                variants={{
                  open: { scale: 1, y: 0, opacity: 1, transition: { type: "spring", stiffness: 500, damping: 22, mass: 0.6 } },
                  closed: { scale: 0.85, y: -8, opacity: 0, transition: { type: "tween", duration: 0.12, ease: "easeIn" } },
                }}
                initial="closed"
                animate="open"
                exit="closed"
                layout
                transition={{ layout: { type: "spring", stiffness: 400, damping: 30 } }}
              >
                <GlassSurface
                  flexDirection="col"
                  alignItems="stretch"
                  order="start"
                  borderRadius={18}
                  className="w-full"
                  innerClassName="p-1"
                >
                  <motion.ul layout className="w-full max-h-60 overflow-y-auto">
                    <AnimatePresence initial={false} mode="popLayout">
                      {isFetching && results.length === 0 && (
                        <motion.li
                          key="searching"
                          layout
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
                          layout
                          initial={{ opacity: 0 }}
                          animate={{ opacity: 1 }}
                          exit={{ opacity: 0 }}
                          transition={{ duration: 0.2 }}
                          className="px-3 py-2 text-xs text-white/50"
                        >
                          No results
                        </motion.li>
                      )}
                      {!isFetching &&
                        results.map(({ name, company_name }) => (
                          <motion.li
                            key={name}
                            layout
                            initial={{ opacity: 0 }}
                            animate={{ opacity: 1 }}
                            exit={{ opacity: 0 }}
                            transition={{ opacity: { duration: 0.2 } }}
                          >
                            <button
                              type="button"
                              onMouseDown={(e) => {
                                e.preventDefault();
                                select(name);
                              }}
                              className="flex w-full items-center justify-between gap-2 rounded-lg px-3 py-1.5 text-left hover:bg-white/10 transition-colors"
                            >
                              <div className="flex min-w-0 flex-col">
                                <span className="text-sm font-semibold uppercase text-white">{name}</span>
                                {company_name && (
                                  <span className="truncate text-[11px] leading-tight text-white/50">
                                    {company_name}
                                  </span>
                                )}
                              </div>
                            </button>
                          </motion.li>
                        ))}
                    </AnimatePresence>
                  </motion.ul>
                </GlassSurface>
              </motion.div>
            )}
          </AnimatePresence>,
          document.body
        )}
    </>
  );
}
