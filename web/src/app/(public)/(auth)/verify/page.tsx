'use client';
import Link from "next/link";
import { Mail } from "lucide-react";
import { TextEffect } from "@/components/primitives/TextEffect";
import { PageEnter, RevealStagger, RevealItem } from "@/components/primitives/Reveal";

export default function Verify() {
  return (
    <PageEnter className="w-full h-full flex flex-col items-center justify-center text-3xl gap-y-6">
      <RevealStagger className="w-full flex flex-col items-center gap-y-6" stagger={0.08} delay={0.05}>
        <RevealItem className="flex h-14 w-14 items-center justify-center rounded-full border border-white/10 bg-white/5">
          <Mail size={24} className="text-white/70" />
        </RevealItem>

        <RevealItem className="text-center flex flex-col items-center gap-2 px-6">
          <TextEffect as="h2" preset="fade-in-blur" per="word" speedReveal={1.3} className="text-xl font-semibold">
            Check your email
          </TextEffect>
          <TextEffect as="p" preset="fade" per="word" delay={0.2} className="text-sm text-white/60 mt-1 max-w-md">
            We sent a verification link to your inbox. Open it to activate your account and start trading.
          </TextEffect>
        </RevealItem>

        <RevealItem className="text-center">
          <p className="text-sm text-white/40">
            Already verified?{" "}
            <Link href="/signin" className="text-white/80 underline underline-offset-4 hover:text-white">
              Sign in
            </Link>
          </p>
        </RevealItem>
      </RevealStagger>
    </PageEnter>
  );
}
