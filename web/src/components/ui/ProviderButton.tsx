import { motion } from "motion/react"
import Image from "next/image";
import { ReactNode } from "react";


interface ProviderButtonProps {
  name: string;
  icon: string;
  color?: string;
  onClick: () => void;
}

export default function ProviderButton({
  color,
  icon,
  name,
  onClick
}: ProviderButtonProps) {




  return (
    <motion.button
      className={`
        bg-${color} w-3/4 h-12 py-2 pointer-events-auto
        hover:cursor-pointer flex justify-center items-center
        rounded-xl border-2 border-gray-600
        text-black text-xl gap-4 font-bold
      `}
      onClick={onClick}
    >
      <Image className="w-6 h-6" src={`/${icon}.svg`} height={15} width={15} alt="Logo" />
    </motion.button>
  );
}