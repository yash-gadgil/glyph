'use client';
import { useRouter } from "next/navigation";
import { ArrowLeft } from "lucide-react";
import { motion } from "motion/react";

export default function BackButton() {
  const router = useRouter();
  return (
    <motion.button
      onClick={() => router.back()}
      whileTap={{ scale: 0.9 }}
      transition={{ duration: 0.2, ease: [0.22, 1, 0.36, 1] }}
      className="absolute top-5 left-5 pointer-events-auto text-white/80 transition-colors cursor-pointer"
      aria-label="Go back"
    >
      <ArrowLeft size={18} />
    </motion.button>
  );
}
