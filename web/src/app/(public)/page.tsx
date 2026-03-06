'use client';
import GlassButton from "@/components/ui/GlassButton";
import GlassSurface from "@/components/primitives/GlassSurface";
import Image from "next/image";
import Link from "next/link";
import { TextEffect } from "@/components/primitives/TextEffect";
import { SlidingNumber } from "@/components/primitives/SlidingNumber";
import { AnimatePresence, motion } from "motion/react";
import {
  ArrowUpRight,
  Activity,
  Bot,
  Briefcase,
  CandlestickChart,
  LineChart,
  Minus,
  Newspaper,
  Plus,
  Wand2,
} from "lucide-react";
import { useState } from "react";

const FEATURES = [
  {
    icon: <CandlestickChart size={22} />,
    title: "Live Charts",
    body: "Multi-timeframe candlestick charts with crosshair tooltips and technical overlays. From 1-minute scalps to 5-year zooms.",
    accent: "text-white",
  },
  {
    icon: <Activity size={22} />,
    title: "Watchlists",
    body: "Stream real-time prices across unlimited custom lists. Sparklines and percentage moves update tick-by-tick.",
    accent: "text-white",
  },
  {
    icon: <Wand2 size={22} />,
    title: "Strategy Builder",
    body: "Compose entry and exit rules over technical indicators. No code, just logical operators and parameters.",
    accent: "text-white",
  },
  {
    icon: <Bot size={22} />,
    title: "Automation",
    body: "Deploy your strategies as autonomous agents. Pause, resume, or unwind any time, full audit trail per trade.",
    accent: "text-white",
  },
  {
    icon: <Briefcase size={22} />,
    title: "Portfolio",
    body: "Track positions, realized & unrealized P&L, margin usage, and buying power in one consolidated dashboard.",
    accent: "text-white",
  },
  {
    icon: <Newspaper size={22} />,
    title: "Market Pulse",
    body: "Top movers and curated financial news, ranked by symbols you actually care about.",
    accent: "text-white",
  },
];

const FAQS = [
  {
    q: "Is this real money or paper trading?",
    a: "Glyph is a paper-trading platform. Every account starts with a simulated cash balance and trades against live market data, so you can explore the full workflow, orders, fills, portfolio, strategies, without risking real money.",
  },
  {
    q: "Do I need to know how to code to build a strategy?",
    a: "No. The Strategy Builder uses logical rule groups (AND/OR) over indicators and price conditions. If you can describe a setup in plain English, you can build it in Glyph.",
  },
  {
    q: "What markets are supported?",
    a: "US equities and ETFs at launch, with crypto and options on the roadmap. Watchlists support any symbol shown in the search index.",
  },
];

const STEPS = [
  {
    n: "01",
    title: "Build a watchlist",
    body: "Search any ticker. Add it to a curated list. Streaming quotes start instantly.",
  },
  {
    n: "02",
    title: "Read the chart",
    body: "Open a symbol for full-screen analytics, RSI, moving averages, 52-week range, fundamentals.",
  },
  {
    n: "03",
    title: "Place a trade",
    body: "Market, limit, stop, or stop-limit. Confirm in one click from any screen on the platform.",
  },
  {
    n: "04",
    title: "Automate it",
    body: "Codify a working setup into a strategy. Deploy it. Monitor P&L from the strategies dashboard.",
  },
];

export default function Home() {
  const ease = [0.22, 1, 0.36, 1] as const;

  return (
    <div className="relative z-10 pointer-events-none">
      <nav className="fixed z-20 top-0 flex justify-between items-center font-mono font-bold py-3 px-6 w-full h-20 backdrop-blur-md bg-black/10">
        <motion.div
          initial={{ opacity: 0, x: -12 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ duration: 0.5, ease }}
          className="flex items-center justify-center gap-x-3 text-2xl lg:text-3xl h-full text-black dark:text-white"
        >
          <motion.div
            initial={{ rotate: -45, scale: 0.6, opacity: 0 }}
            animate={{ rotate: 0, scale: 1, opacity: 1 }}
            transition={{ duration: 0.6, ease }}
          >
            <Image className="dark:invert h-fit w-fit" src="/jera.svg" height={20} width={20} alt="Glyph" />
          </motion.div>
          <TextEffect as="span" preset="fade-in-blur" per="char" speedReveal={1.4} delay={0.2}>
            GLYPH
          </TextEffect>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, x: 12 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ duration: 0.5, delay: 0.3, ease }}
          className="flex justify-center items-center gap-x-4 pointer-events-auto"
        >
          <Link href="#features" className="hidden md:block text-xs uppercase tracking-widest text-white/60 hover:text-white transition-colors">
            Features
          </Link>
          <Link href="#how" className="hidden md:block text-xs uppercase tracking-widest text-white/60 hover:text-white transition-colors">
            How it Works
          </Link>
          <Link href="#faq" className="hidden md:block text-xs uppercase tracking-widest text-white/60 hover:text-white transition-colors">
            FAQ
          </Link>
          <GlassButton text="Sign In" href="/signin" />
        </motion.div>
      </nav>

      <main className="relative w-full overflow-hidden">
        <section className="min-h-screen flex flex-col items-center justify-center px-6 pt-24 pb-16 relative">
          <TextEffect
            as="h1"
            preset="fade-in-blur"
            per="word"
            delay={0.5}
            speedReveal={1.1}
            className="text-center font-mono font-bold text-5xl sm:text-7xl lg:text-8xl tracking-tighter leading-[0.95] text-white max-w-5xl drop-shadow-[0_8px_30px_rgba(0,0,0,0.5)]"
          >
            Trade with precision.
          </TextEffect>

          <TextEffect
            as="p"
            preset="fade"
            per="word"
            delay={1.0}
            speedReveal={1.4}
            className="mt-6 text-base sm:text-lg text-white/70 max-w-2xl text-center leading-relaxed font-mono"
          >
            Real-time charts, programmable strategies, and a low-latency order book, everything you need to move at market speed, in one terminal.
          </TextEffect>

          <motion.div
            initial={{ opacity: 0, y: 12 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 1.5, ease }}
            className="mt-10 flex flex-col sm:flex-row items-center gap-4 pointer-events-auto"
          >
            <GlassButton text="Get Started" href="/signup" className="px-6 py-2 text-sm" />
            <Link
              href="/signin"
              className="text-xs uppercase tracking-widest text-white/60 hover:text-white transition-colors flex items-center gap-2 font-mono"
            >
              I have an account <ArrowUpRight size={12} />
            </Link>
          </motion.div>

        </section>

        <section className="relative py-20 px-6">
          <div className="max-w-6xl mx-auto grid grid-cols-2 md:grid-cols-4 gap-6">
            {[
              { label: "Symbols Tracked", value: 8400, suffix: "+", delay: 0 },
              { label: "Avg Latency", value: 12, suffix: "ms", delay: 0.1 },
              { label: "Active Strategies", value: 320, suffix: "", delay: 0.2 },
              { label: "Markets Covered", value: 24, suffix: "/7", delay: 0.3 },
            ].map((s) => (
              <motion.div
                key={s.label}
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true, margin: "-80px" }}
                transition={{ duration: 0.5, delay: s.delay, ease }}
                className="text-center font-mono pointer-events-auto"
              >
                <p className="text-4xl md:text-5xl font-bold text-white tracking-tighter flex items-center justify-center">
                  <SlidingNumber value={s.value} />
                  <span className="text-white/40 ml-0.5">{s.suffix}</span>
                </p>
                <p className="mt-2 text-[10px] uppercase tracking-[0.25em] text-white/40">{s.label}</p>
              </motion.div>
            ))}
          </div>
        </section>

        <section id="features" className="relative py-24 px-6">
          <div className="max-w-6xl mx-auto">
            <motion.div
              initial={{ opacity: 0, y: 18 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true, margin: "-80px" }}
              transition={{ duration: 0.5, ease }}
              className="mb-16 text-center"
            >
              <TextEffect
                as="h2"
                preset="fade-in-blur"
                per="word"
                className="font-mono text-4xl md:text-5xl font-bold text-white tracking-tighter max-w-3xl mx-auto"
              >
                A complete trading terminal.
              </TextEffect>
              <p className="mt-4 text-white/60 text-sm max-w-xl mx-auto font-mono">
                From research to execution to automation, Glyph fits the entire trade lifecycle on one screen.
              </p>
            </motion.div>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5 pointer-events-auto">
              {FEATURES.map((f, i) => (
                <motion.div
                  key={f.title}
                  initial={{ opacity: 0, y: 24 }}
                  whileInView={{ opacity: 1, y: 0 }}
                  viewport={{ once: true, margin: "-60px" }}
                  transition={{ duration: 0.5, delay: i * 0.06, ease }}
                >
                  <GlassSurface
                    displace={15}
                    distortionScale={-100}
                    opacity={0.7}
                    borderRadius={24}
                    className="h-full hover:border-white/20 transition-colors"
                  >
                    <div className="p-7 w-full h-full flex flex-col gap-4">
                      <div className={`w-fit p-3 rounded-xl bg-white/5 border border-white/10 ${f.accent}`}>
                        {f.icon}
                      </div>
                      <h3 className="font-mono font-bold text-white text-xl tracking-tight">{f.title}</h3>
                      <p className="text-sm text-white/60 leading-relaxed font-mono">{f.body}</p>
                    </div>
                  </GlassSurface>
                </motion.div>
              ))}
            </div>
          </div>
        </section>

        <section className="relative py-20 px-6">
          <div className="max-w-6xl mx-auto grid grid-cols-1 lg:grid-cols-5 gap-8 items-center">
            <motion.div
              initial={{ opacity: 0, x: -24 }}
              whileInView={{ opacity: 1, x: 0 }}
              viewport={{ once: true, margin: "-100px" }}
              transition={{ duration: 0.6, ease }}
              className="lg:col-span-2"
            >
              <h2 className="font-mono text-4xl md:text-5xl font-bold text-white tracking-tighter leading-tight">
                One terminal.<br />Every market.
              </h2>
              <p className="mt-5 text-white/60 leading-relaxed font-mono text-sm">
                Charts, order book, watchlists, news, and strategy automation, all in a single keyboard-driven workspace.
                Switch context without losing flow.
              </p>

            </motion.div>

            <motion.div
              initial={{ opacity: 0, y: 30, scale: 0.96 }}
              whileInView={{ opacity: 1, y: 0, scale: 1 }}
              viewport={{ once: true, margin: "-100px" }}
              transition={{ duration: 0.7, ease }}
              className="lg:col-span-3 pointer-events-auto"
            >
              <GlassSurface displace={20} distortionScale={-150} opacity={0.7} borderRadius={28} className="overflow-hidden">
                <div className="w-full p-6 font-mono text-xs space-y-4">
                  <div className="flex items-center justify-between border-b border-white/10 pb-3">
                    <div className="flex items-center gap-2">
                      <span className="w-2.5 h-2.5 rounded-full bg-rose-400/70" />
                      <span className="w-2.5 h-2.5 rounded-full bg-amber-300/70" />
                      <span className="w-2.5 h-2.5 rounded-full bg-emerald-400/70" />
                    </div>
                    <span className="text-[10px] uppercase tracking-widest text-white/40">terminal · NVDA</span>
                  </div>

                  <div className="grid grid-cols-3 gap-3 pt-1">
                    {[
                      { label: "Bid", value: "875.28", tone: "text-emerald-400" },
                      { label: "Ask", value: "875.32", tone: "text-rose-400" },
                      { label: "Last", value: "875.30", tone: "text-white" },
                    ].map((q) => (
                      <div key={q.label} className="bg-black/30 border border-white/10 rounded-lg p-3">
                        <p className="text-[9px] uppercase tracking-widest text-white/40 mb-1">{q.label}</p>
                        <p className={`text-base font-bold ${q.tone}`}>{q.value}</p>
                      </div>
                    ))}
                  </div>

                  <DemoMiniChart />

                  <div className="grid grid-cols-2 gap-3 text-[10px]">
                    <div className="rounded-lg border border-emerald-500/20 bg-emerald-500/10 p-2 flex justify-between items-center">
                      <span className="font-bold text-emerald-400 uppercase tracking-widest">BUY</span>
                      <span className="text-emerald-300/80">100 @ 875.30</span>
                    </div>
                    <div className="rounded-lg border border-white/10 bg-white/5 p-2 flex justify-between items-center">
                      <span className="font-bold text-white/60 uppercase tracking-widest">FILLED</span>
                      <span className="text-white/50">0.21s</span>
                    </div>
                  </div>
                </div>
              </GlassSurface>
            </motion.div>
          </div>
        </section>

        <section id="how" className="relative py-24 px-6">
          <div className="max-w-6xl mx-auto">
            <motion.div
              initial={{ opacity: 0, y: 18 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true, margin: "-80px" }}
              transition={{ duration: 0.5, ease }}
              className="mb-16 text-center"
            >
              <TextEffect
                as="h2"
                preset="fade-in-blur"
                per="word"
                className="font-mono text-4xl md:text-5xl font-bold text-white tracking-tighter"
              >
                How it works.
              </TextEffect>
            </motion.div>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 pointer-events-auto">
              {STEPS.map((step, i) => (
                <motion.div
                  key={step.n}
                  initial={{ opacity: 0, y: 22 }}
                  whileInView={{ opacity: 1, y: 0 }}
                  viewport={{ once: true, margin: "-60px" }}
                  transition={{ duration: 0.5, delay: i * 0.08, ease }}
                  className="relative font-mono group"
                >
                  <div className="text-6xl font-bold text-white/90 transition-colors mb-4 leading-none">
                    {step.n}
                  </div>
                  <h3 className="text-lg font-bold text-white tracking-tight mb-2">{step.title}</h3>
                  <p className="text-sm text-white/60 leading-relaxed">{step.body}</p>
                  {i < STEPS.length - 1 && (
                    <div className="hidden lg:block absolute top-6 right-0 w-full h-px bg-linear-to-r from-white/10 to-transparent translate-x-1/2 -z-10" />
                  )}
                </motion.div>
              ))}
            </div>
          </div>
        </section>

        <section id="faq" className="relative py-24 px-6">
          <div className="max-w-3xl mx-auto">
            <motion.div
              initial={{ opacity: 0, y: 18 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true, margin: "-80px" }}
              transition={{ duration: 0.5, ease }}
              className="mb-12 text-center"
            >
              <TextEffect
                as="h2"
                preset="fade-in-blur"
                per="word"
                className="font-mono text-4xl md:text-5xl font-bold text-white tracking-tighter"
              >
                FAQs
              </TextEffect>
            </motion.div>

            <div className="space-y-3 pointer-events-auto">
              {FAQS.map((f, i) => (
                <FaqItem key={f.q} q={f.q} a={f.a} delay={i * 0.05} ease={ease} />
              ))}
            </div>
          </div>
        </section>

        <footer className="relative border-t border-white/10 bg-black/30 backdrop-blur-md mt-12 pointer-events-auto">
          <div className="max-w-6xl mx-auto px-6 py-10 grid grid-cols-2 md:grid-cols-4 gap-8 font-mono text-sm">
            <div className="col-span-2 md:col-span-1">
              <div className="flex items-center gap-2 mb-3">
                <Image className="dark:invert" src="/jera.svg" height={16} width={16} alt="Glyph" />
                <span className="font-bold text-white">GLYPH</span>
              </div>
              <p className="text-white/40 text-xs leading-relaxed">
                A modern trading terminal, built for speed, programmable from day one.
              </p>
            </div>
            <FooterCol title="Product" links={[
              { label: "Watchlists", href: "/watchlist" },
              { label: "Charts", href: "#features" },
              { label: "Strategies", href: "#features" },
              { label: "Orders", href: "#features" },
            ]} />
            <FooterCol title="Resources" links={[
              { label: "How it works", href: "#how" },
              { label: "FAQ", href: "#faq" },
              { label: "Sign In", href: "/signin" },
              { label: "Sign Up", href: "/signup" },
            ]} />
          </div>
          <div className="border-t border-white/10 px-6 py-5 max-w-6xl mx-auto flex flex-col sm:flex-row justify-between items-center gap-3 font-mono text-[10px] uppercase tracking-widest text-white/30">
            <span>© {new Date().getFullYear()} Glyph Markets, All rights reserved.</span>
          </div>
        </footer>
      </main>
    </div>
  );
}

function FaqItem({
  q,
  a,
  delay,
  ease,
}: {
  q: string;
  a: string;
  delay: number;
  ease: readonly [number, number, number, number];
}) {
  const [open, setOpen] = useState(false);
  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true, margin: "-40px" }}
      transition={{ duration: 0.4, delay, ease }}
    >
      <GlassSurface
        displace={10}
        distortionScale={-60}
        opacity={0.5}
        borderRadius={16}
        className="overflow-hidden hover:border-white/20 transition-colors"
      >
        <button
          onClick={() => setOpen((o) => !o)}
          className="w-full flex items-center justify-between gap-4 px-6 py-5 text-left font-mono group"
        >
          <span className="text-sm md:text-base font-medium text-white/85 group-hover:text-white transition-colors">
            {q}
          </span>
          <motion.div
            animate={{ rotate: open ? 180 : 0 }}
            transition={{ duration: 0.25, ease }}
            className="text-white/90 shrink-0"
          >
            {open ? <Minus size={16} /> : <Plus size={16} />}
          </motion.div>


          <AnimatePresence initial={false}>
            {open && (
              <motion.div
                key="content"
                initial={{ height: 0, opacity: 0 }}
                animate={{ height: "auto", opacity: 1 }}
                exit={{ height: 0, opacity: 0 }}
                transition={{
                  height: { duration: 0.3, ease },
                  opacity: { duration: 0.2 },
                }}
                className="overflow-hidden"
              >
                <div className="p-3">
                  <p className="text-sm md:text-[15px] leading-7 text-white/60">
                    {a}
                  </p>
                </div>
              </motion.div>
            )}
          </AnimatePresence>
        </button>
      </GlassSurface>
    </motion.div>
  );
}

function FooterCol({ title, links }: { title: string; links: { label: string; href: string }[] }) {
  return (
    <div>
      <p className="text-[10px] uppercase tracking-[0.25em] text-white/40 mb-3 font-bold">{title}</p>
      <ul className="space-y-2">
        {links.map((l) => (
          <li key={l.label}>
            <Link href={l.href} className="text-xs text-white/60 hover:text-white transition-colors">
              {l.label}
            </Link>
          </li>
        ))}
      </ul>
    </div>
  );
}

function DemoMiniChart() {
  const points = [40, 38, 42, 45, 41, 48, 52, 50, 55, 58, 56, 62, 65, 63, 68, 72, 70, 75, 78, 82];
  const max = Math.max(...points);
  const min = Math.min(...points);
  const norm = (v: number) => 100 - ((v - min) / (max - min)) * 80;
  const path = points
    .map((p, i) => `${i === 0 ? "M" : "L"} ${(i / (points.length - 1)) * 100} ${norm(p)}`)
    .join(" ");

  return (
    <div className="rounded-lg border border-white/10 bg-black/30 p-3">
      <div className="flex items-center justify-between text-[10px] mb-2">
        <span className="text-white/50 uppercase tracking-widest">1D · 5min</span>
        <span className="text-emerald-400 flex items-center gap-1">
          <LineChart size={10} /> +2.41%
        </span>
      </div>
      <svg viewBox="0 0 100 100" preserveAspectRatio="none" className="w-full h-20 overflow-visible">
        <defs>
          <linearGradient id="demoFill" x1="0" x2="0" y1="0" y2="1">
            <stop offset="0%" stopColor="#34d399" stopOpacity="0.35" />
            <stop offset="100%" stopColor="#34d399" stopOpacity="0" />
          </linearGradient>
        </defs>
        <motion.path
          d={`${path} L 100 100 L 0 100 Z`}
          fill="url(#demoFill)"
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6, delay: 0.4 }}
        />
        <motion.path
          d={path}
          fill="none"
          stroke="#34d399"
          strokeWidth="1.2"
          strokeLinecap="round"
          strokeLinejoin="round"
          initial={{ pathLength: 0 }}
          whileInView={{ pathLength: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 1.4, ease: [0.22, 1, 0.36, 1] }}
        />
      </svg>
    </div>
  );
}

