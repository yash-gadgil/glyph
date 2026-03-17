'use client';

import { X } from "lucide-react";
import { useState } from "react";
import { motion, AnimatePresence } from "motion/react";

interface WatchlistButtonProps {
  name: string;
  id: string;
  selected?: boolean;
  changeWatchlist: (watchlistId: string) => void;
  onDelete?: (watchlistId: string) => void;
}

export default function WatchlistButton({
  name,
  id,
  changeWatchlist,
  onDelete,
  selected = false,
}: WatchlistButtonProps) {
  const [hovered, setHovered] = useState(false);

  return (
    <div
      className="flex items-center gap-1 pointer-events-auto relative"
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
    >
      <motion.button
        whileHover={{ scale: 1.05 }}
        whileTap={{ scale: 0.97 }}
        transition={{ duration: 0.2, ease: [0.22, 1, 0.36, 1] }}
        className={`relative text-xl hover:cursor-pointer px-1 py-1 ${selected ? "text-white" : "text-white/70 hover:text-white"}`}
        onClick={() => changeWatchlist(id)}
      >
        {name}
        {selected && (
          <motion.div
            layoutId="watchlist-tab-underline"
            className="absolute bottom-0 left-0 right-0 h-0.5 bg-white rounded-full"
            transition={{ type: "spring", stiffness: 400, damping: 30 }}
          />
        )}
      </motion.button>
      <AnimatePresence>
        {onDelete && hovered && (
          <motion.button
            initial={{ opacity: 0, scale: 0.7, x: -4 }}
            animate={{ opacity: 1, scale: 1, x: 0 }}
            exit={{ opacity: 0, scale: 0.7, x: -4 }}
            transition={{ duration: 0.15 }}
            onClick={() => onDelete(id)}
            className="p-0.5 rounded hover:bg-white/20"
            title="Delete watchlist"
          >
            <X size={14} />
          </motion.button>
        )}
      </AnimatePresence>
    </div>
  );
}
