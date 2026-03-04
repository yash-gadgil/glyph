'use client';
import { useState } from "react";
import GlassSurface from "./GlassSurface";
import { motion, AnimatePresence } from "motion/react";

interface ExpandableGlassProps {
  closed: (expand: () => void) => React.ReactNode;
  opened: (close: () => void) => React.ReactNode;
}

export default function ExpandableGlass({
  closed,
  opened,
}: ExpandableGlassProps) {

  const [open, setOpen] = useState<boolean>(false);

  const expand = () => setOpen(true);
  const close = () => setOpen(false);

  return (
    <GlassSurface
      displace={10}
      distortionScale={-40}
      redOffset={2}
      greenOffset={5}
      blueOffset={8}
      brightness={60}
      opacity={0.8}
      mixBlendMode="overlay"
      className="w-fit overflow-hidden h-fit pointer-events-auto"
    >
      <motion.div
        layout
        transition={{ type: "spring", stiffness: 400, damping: 30 }}
        className="flex items-center gap-x-1 py-1 px-2"
      >
        <AnimatePresence mode="sync">
          {open ? (
            <motion.div
              key="expanded"
              className="flex items-center overflow-hidden"
              initial={{ opacity: 0, width: 0 }}
              animate={{ opacity: 1, width: "auto" }}
              exit={{ opacity: 0, width: 0 }}
              transition={{ duration: 0.15 }}
            >
              {opened(close)}
            </motion.div>
          ) : (
            <motion.div
              key="collapsed"
              className="flex items-center overflow-hidden"
              initial={{ opacity: 0, width: 0 }}
              animate={{ opacity: 1, width: "auto" }}
              exit={{ opacity: 0, width: 0 }}
              transition={{ duration: 0.15 }}
            >
              {closed(expand)}
            </motion.div>
          )}
        </AnimatePresence>
      </motion.div>
    </GlassSurface>
  );
}