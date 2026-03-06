'use client';

import { useRouter } from "next/navigation";
import { useEffect } from "react";
import { motion } from "motion/react";
import { TextEffect } from "@/components/primitives/TextEffect";

export default function Callback() {
  const router = useRouter();

  useEffect(() => {
    router.push("/dashboard");
  }, [router]);

  return (
    <div className="relative h-screen w-screen flex flex-col items-center justify-center text-xl font-mono gap-4">
      <motion.div
        animate={{ rotate: 360 }}
        transition={{ duration: 1.4, repeat: Infinity, ease: "linear" }}
        className="h-10 w-10 rounded-full border-2 border-white/10 border-t-emerald-400"
      />
      <TextEffect as="p" preset="fade-in-blur" per="char" className="text-white/80 tracking-wide">
        Signing in...
      </TextEffect>
    </div>
  );
}
