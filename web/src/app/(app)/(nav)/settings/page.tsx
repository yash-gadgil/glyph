'use client';
import Link from "next/link";
import { GlassCard, SettingRow } from "@/components/ui/GlassCard";
import { Palette, Activity, Database } from "lucide-react";
import { TextEffect } from "@/components/primitives/TextEffect";
import { PageEnter, RevealStagger, RevealItem } from "@/components/primitives/Reveal";
import { motion } from "motion/react";
import GlassButton from "@/components/ui/GlassButton";

export default function SettingsPage() {
  return (
    <PageEnter className="relative min-h-screen w-full flex flex-col items-center justify-start font-mono p-4 pt-32 pb-32 md:p-8 md:pt-32 xl:p-12 xl:pt-32 z-0">
      <RevealStagger className="w-full max-w-5xl flex flex-col" stagger={0.08}>
        <RevealItem className="mb-8 flex justify-between items-end">
          <TextEffect as="h1" preset="fade-in-blur" per="word" className="text-3xl sm:text-4xl font-semibold text-white/90">
            Platform Settings
          </TextEffect>
        </RevealItem>

        <RevealStagger className="w-full grid grid-cols-1 lg:grid-cols-2 gap-6 lg:gap-8" stagger={0.08} delay={0.1}>
          <RevealItem>
            <motion.div transition={{ duration: 0.25 }}>
              <GlassCard title="Display" icon={Palette} flexCol>
                <SettingRow
                  label="Base Currency"
                  description="The currency all balances and charts are displayed in."
                  rightElement={
                    <span className="text-xs border border-white/10 px-2 py-1 rounded text-white/40 bg-white/5">USD ($)</span>
                  }
                />
              </GlassCard>
            </motion.div>
          </RevealItem>

          <RevealItem>
            <motion.div transition={{ duration: 0.25 }}>
              <GlassCard title="Trading" icon={Activity} flexCol>
                <SettingRow
                  label="Venue"
                  description="Orders fill against live market prices on Glyph's own paper engine, no real money, no real exchange."
                  rightElement={
                    <span className="text-xs border border-emerald-500/20 px-2 py-1 rounded text-emerald-400 bg-emerald-500/10">Paper</span>
                  }
                />
                <SettingRow
                  label="Market Hours"
                  description="Orders are accepted any time; fills happen 9:30 AM - 4:00 PM ET, Monday to Friday."
                  rightElement={
                    <span className="text-xs border border-white/10 px-2 py-1 rounded text-white/40 bg-white/5">NYSE</span>
                  }
                />
              </GlassCard>
            </motion.div>
          </RevealItem>

          <RevealItem className="lg:col-span-2">
            <motion.div transition={{ duration: 0.25 }}>
              <GlassCard title="Account Data" icon={Database} flexCol>
                <SettingRow
                  label="Starting Balance"
                  description="Every paper account starts with the same fixed balance, so performance stays comparable."
                  rightElement={
                    <span className="text-xs border border-white/10 px-2 py-1 rounded text-white/40 bg-white/5">$100,000</span>
                  }
                />
                <SettingRow
                  label="Reset Account"
                  description="Wipe positions and history and restore the starting balance from the Account page."
                  rightElement={
                    <GlassButton
                      text="Open Account"
                      href="/account"
                      className="text-xs px-3 py-1.5 rounded-lg "
                    >
                    </GlassButton>
                  }
                />
              </GlassCard>
            </motion.div>
          </RevealItem>
        </RevealStagger>
      </RevealStagger>
    </PageEnter>
  );
}
