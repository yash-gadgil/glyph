'use client';

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useEffect, useRef, useState } from "react";
import GlassSurface from "../primitives/GlassSurface";
import MarketStatusBadge from "./MarketStatusBadge";
import { Settings, Menu, X } from "lucide-react";
import { motion, AnimatePresence } from "motion/react";

const NAV_LINKS = [
  { href: "/portfolio", label: "Portfolio" },
  { href: "/explore", label: "Explore" },
  { href: "/watchlist", label: "Watchlists" },
  { href: "/strategies", label: "Strategies" },
  { href: "/orders", label: "Orders" },
];

const menuListVariants = {
  open: { transition: { staggerChildren: 0.045, delayChildren: 0.05 } },
  closed: { transition: { staggerChildren: 0.02, staggerDirection: -1 } },
};
const menuItemVariants = {
  open: { opacity: 1, x: 0, transition: { type: "spring", stiffness: 500, damping: 30 } },
  closed: { opacity: 0, x: 12 },
} as const;

export default function Navbar() {
  const pathname = usePathname();
  const [menuOpen, setMenuOpen] = useState(false);
  const mobileRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!menuOpen) return;
    function onDown(e: MouseEvent) {
      if (mobileRef.current && !mobileRef.current.contains(e.target as Node)) {
        setMenuOpen(false);
      }
    }
    document.addEventListener("mousedown", onDown);
    return () => document.removeEventListener("mousedown", onDown);
  }, [menuOpen]);

  return (
    <motion.nav
      initial={{ y: -16, opacity: 0 }}
      animate={{ y: 0, opacity: 1 }}
      transition={{ duration: 0.5, ease: [0.22, 1, 0.36, 1] }}
      className="fixed top-3 h-20 w-full px-4 flex flex-row justify-between items-start lg:items-center gap-x-4 z-10"
    >
      <GlassSurface
        displace={12}
        distortionScale={-45}
        redOffset={2}
        greenOffset={5}
        blueOffset={8}
        brightness={60}
        opacity={0.8}
        mixBlendMode="overlay"
        borderRadius={50}
        order="between"
        className="w-fit lg:w-full h-fit lg:h-full"
      >
        <Link href="/dashboard" className="font-bold text-2xl group">
          <motion.span
            className="inline-block"
            whileHover={{ scale: 1.05, letterSpacing: "0.05em" }}
            transition={{ duration: 0.25, ease: [0.22, 1, 0.36, 1] }}
          >
            GLYPH
          </motion.span>
        </Link>
        <div className="hidden lg:flex w-fit justify-center items-center gap-x-10 uppercase relative">
          {NAV_LINKS.map((link) => {
            const active = pathname?.startsWith(link.href);
            return (
              <Link
                key={link.href}
                href={link.href}
                className="relative px-1 py-1.5 transition-colors group"
              >
                <span className={active ? "text-white" : "text-white/70 group-hover:text-white"}>
                  {link.label}
                </span>
                {active && (
                  <motion.div
                    layoutId="navbar-active-underline"
                    className="absolute -bottom-0.5 left-0 right-0 h-0.5 bg-white rounded-full"
                    transition={{ type: "spring", stiffness: 380, damping: 30 }}
                  />
                )}
              </Link>
            );
          })}
        </div>
        <div className="hidden lg:flex items-center gap-6">
          <MarketStatusBadge variant="compact" />
          <Link
            href="/account"
            className="hover:underline uppercase tracking-widest font-semibold flex items-center gap-2"
          >
            Account
          </Link>
          <Link
            href="/settings"
            className="hover:text-white/70 transition-colors p-1 rounded-full bg-white/5 border border-white/10 hover:bg-white/10"
          >
            <motion.span
              whileHover={{ rotate: 60 }}
              whileTap={{ rotate: 180 }}
              transition={{ duration: 0.4, ease: [0.22, 1, 0.36, 1] }}
              className="inline-flex align-middle"
            >
              <Settings size={24} />
            </motion.span>
          </Link>
        </div>
      </GlassSurface>

      <div ref={mobileRef} className="lg:hidden relative pointer-events-auto">
        <GlassSurface borderRadius={50} className="w-fit h-fit pointer-events-auto">
          <button
            onClick={() => setMenuOpen((o) => !o)}
            className="flex items-center p-1 hover:cursor-pointer"
            aria-label={menuOpen ? "Close navigation menu" : "Open navigation menu"}
            aria-expanded={menuOpen}
          >
            {menuOpen ? <X size={20} /> : <Menu size={20} />}
          </button>
        </GlassSurface>

        <AnimatePresence>
          {menuOpen && (
            <motion.div
              layout
              className="absolute top-full mt-2 right-0 z-50 origin-top-right isolate overflow-hidden rounded-[20px] pointer-events-auto"
              variants={{
                open: { scale: 1, y: 0, transition: { type: "spring", stiffness: 500, damping: 22, mass: 0.6 } },
                closed: { scale: 0.85, y: -8, transition: { type: "tween", duration: 0.12, ease: "easeIn" } },
              }}
              initial="closed"
              animate="open"
              exit="closed"
            >
              <GlassSurface borderRadius={20} className="w-44">
                <motion.div
                  variants={menuListVariants}
                  className="flex flex-col gap-3 uppercase text-sm tracking-wide w-full"
                >
                  {NAV_LINKS.map((link) => (
                    <motion.div key={link.href} variants={menuItemVariants}>
                      <Link
                        href={link.href}
                        onClick={() => setMenuOpen(false)}
                        className="block hover:opacity-70 transition-opacity whitespace-nowrap px-3"
                      >
                        {link.label}
                      </Link>
                    </motion.div>
                  ))}
                  <motion.div variants={menuItemVariants}>
                    <Link
                      href="/account"
                      onClick={() => setMenuOpen(false)}
                      className="block hover:opacity-70 transition-opacity whitespace-nowrap px-3"
                    >
                      Account
                    </Link>
                  </motion.div>
                  <motion.div variants={menuItemVariants}>
                    <Link
                      href="/settings"
                      onClick={() => setMenuOpen(false)}
                      className="hover:opacity-70 transition-opacity whitespace-nowrap flex items-center gap-2 px-3"
                    >
                      <Settings size={14} /> Settings
                    </Link>
                  </motion.div>
                </motion.div>
              </GlassSurface>
            </motion.div>
          )}
        </AnimatePresence>
      </div>
    </motion.nav>
  );
}
