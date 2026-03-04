'use client';

import { useEffect, useState } from "react";
import PixelEther from "../primitives/PixelEther";

const DARK_COLORS: [number, number, number][] = [
  [0.33, 0.1, 0.66],
  [1.0, 0.0, 0.2],
];

export default function PixelEtherBackground() {
  const [isMobile, setIsMobile] = useState(false);
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setIsMobile(/iPhone|iPad|iPod|Android/i.test(navigator.userAgent));
    setMounted(true);
  }, []);

  if (!mounted) {
    return <div className="fixed inset-0 w-full h-full z-0 bg-black dark:bg-black" />;
  }

  return (
    <div className="fixed inset-0 w-full h-full z-0  pointer-events-none  " style={{ backgroundColor: '#000000' }}>
      <PixelEther
        colors={DARK_COLORS}
        mouseForce={isMobile ? 10 : 20}
        cursorSize={isMobile ? 50 : 100}
        isViscous={false}
        viscous={30}
        iterationsViscous={isMobile ? 8 : 32}
        iterationsPoisson={isMobile ? 8 : 32}
        resolution={isMobile ? 0.25 : 0.5}
        isBounce={false}
        autoDemo={true}
        autoSpeed={0.5}
        autoIntensity={isMobile ? 1.5 : 2.2}
        takeoverDuration={0.25}
        autoResumeDelay={3000}
        autoRampDuration={0.6}
      />
    </div>
  );
}
