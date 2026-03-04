'use client';
import Link from "next/link";
import GlassSurface from "../primitives/GlassSurface";
import { cn } from "@/lib/utils";
import { MouseEventHandler } from "react";
import { motion } from "motion/react"

interface GlassButtonProps {
  text?: string;
  href?: string;
  icon?: React.ReactNode;
  className?: string
  type?: string
  disabled?: boolean
  onClick?: MouseEventHandler<HTMLButtonElement>
}

export default function GlassButton({
  text,
  href,
  icon,
  className,
  type,
  disabled,
  onClick
}: GlassButtonProps) {


  return (
    <motion.div
      whileHover={{ filter: "brightness(2.0)" }}
      whileTap={{ scale: 0.95, filter: "brightness(0.9)" }}
      className="w-fit h-fit"
    >
      <GlassSurface
        displace={15}
        distortionScale={-150}
        redOffset={5}
        greenOffset={15}
        blueOffset={25}
        brightness={60}
        opacity={0.8}
        mixBlendMode="overlay"
        className={cn("w-fit h-fit py-1 px-2 pointer-events-auto hover:cursor-pointer", className)}
      >
        {
          href ?
            <Link href={href} className="flex items-center justify-center gap-x-1 pointer-events-auto hover:cursor-pointer">
              {icon} {text}
            </Link>
            :
            <button className="flex items-center justify-center gap-x-1 pointer-events-auto hover:cursor-pointer" onClick={onClick}
              disabled={disabled}
            >
              {icon}
              {text}
            </button>
        }
      </GlassSurface>
    </motion.div>
  );
}