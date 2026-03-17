'use client';
import { useState, useEffect, useId } from "react";
import GlassSurface from "../primitives/GlassSurface";
import { motion } from "motion/react";
import { FieldOption } from "./GenericForm";

interface GlassToggleProps {
  options: [FieldOption, FieldOption];
  value?: any;
  onChange?: (value: any) => void;
}

export default function GlassToggle({
  options,
  value,
  onChange
}: GlassToggleProps) {

  const [localValue, setLocalValue] = useState<any>(options[0].value);
  const currentValue = value !== undefined ? value : localValue;
  const toggleId = useId();

  useEffect(() => {
    if (onChange && value === undefined) {
      onChange(options[0].value);
    }
  }, [onChange, value, options]);

  const handleSelect = (idx: 0 | 1) => {
    const val = options[idx].value;
    if (onChange) {
      onChange(val);
    } else {
      setLocalValue(val);
    }
  };

  const toggleState = () => {
    const currentIdx = currentValue === options[0].value ? 0 : 1;
    const nextIdx = currentIdx === 0 ? 1 : 0;
    handleSelect(nextIdx);
  };

  return (
    <GlassSurface
      displace={25}
      distortionScale={-100}
      redOffset={10}
      greenOffset={15}
      blueOffset={20}
      brightness={60}
      opacity={0.85}
      mixBlendMode="overlay"
      borderRadius={50}
      className="w-fit h-fit pointer-events-auto shadow-inner"
    >
      <div
        className="relative flex items-center p-1 gap-1 z-10 w-full h-full cursor-pointer"
        onClick={toggleState}
      >
        {options.map((option, idx) => {
          const isSelected = currentValue === option.value;
          return (
            <div
              key={String(option.value) + idx}
              className="relative px-5 py-2 text-sm font-bold uppercase tracking-widest pointer-events-none z-20 group"
            >
              {isSelected && (
                <motion.div
                  layoutId={`active-glass-pill-${toggleId}`}
                  className="absolute inset-0 bg-white/20 border border-white/30 rounded-full -z-10 shadow-[0_0_15px_rgba(255,255,255,0.1)] backdrop-blur-md"
                  transition={{
                    type: "spring",
                    stiffness: 500,
                    damping: 30,
                    mass: 0.8
                  }}
                />
              )}
              <span className={`relative z-10 flex w-full justify-center text-center transition-colors duration-300 ${isSelected ? "text-white drop-shadow-md" : "text-white/40 group-hover:text-white/70"}`}>
                {option.label}
              </span>
            </div>
          )
        })}
      </div>
    </GlassSurface>
  );
}