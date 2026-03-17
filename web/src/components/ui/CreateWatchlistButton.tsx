"use client";

import { useRef, useState } from "react";
import { Plus, Check } from "lucide-react";
import { useQueryClient } from "@tanstack/react-query";
import { createWatchlist } from "@/services/watchlists/mutations";
import ExpandableGlass from "../primitives/ExpandableGlass";
import { motion } from "motion/react";

export default function CreateWatchlistButton() {

  const create = createWatchlist();
  const qc = useQueryClient();
  const [name, setName] = useState("");
  const inputRef = useRef<HTMLInputElement>(null);

  const submit = () => {
    const trimmed = name.trim();
    if (!trimmed) return;
    create.mutate(trimmed, {
      onSuccess: () => qc.invalidateQueries({ queryKey: ["watchlists"] }),
    });
    setName("");
  };

  return (
    <ExpandableGlass

      closed={(expand) =>
        <motion.button
          onClick={expand}
          whileHover={{ rotate: 90, scale: 1.15 }}
          whileTap={{ scale: 0.9 }}
          transition={{ duration: 0.25, ease: [0.22, 1, 0.36, 1] }}
          className="flex"
        >
          <Plus size={16} className="opacity-60 hover:opacity-100" />
        </motion.button>
      }

      opened={(close) =>
        <>
          <input
            ref={inputRef}
            value={name}
            onChange={(e) => setName(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter" || e.key === "Escape") {
                submit();
                close();
              }
            }}
            onBlur={() => { if (!name.trim()) (() => { })() }}
            placeholder="watchlist name"
            className="bg-transparent outline-none font-mono text-sm w-32 placeholder:opacity-40"
          />
          <button onClick={() => {
            submit();
            close();
          }} className="hover:cursor-pointer opacity-60 hover:opacity-100 transition-opacity">
            <Check size={16} />
          </button>
        </>
      }

    />
  );
}