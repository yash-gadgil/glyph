'use client';
import React from 'react';

export interface GlassSurfaceProps {
  children?: React.ReactNode;
  width?: number | string;
  height?: number | string;
  borderRadius?: number;
  borderWidth?: number;
  brightness?: number;
  opacity?: number;
  blur?: number;
  displace?: number;
  backgroundOpacity?: number;
  saturation?: number;
  distortionScale?: number;
  redOffset?: number;
  greenOffset?: number;
  blueOffset?: number;
  xChannel?: 'R' | 'G' | 'B';
  yChannel?: 'R' | 'G' | 'B';
  mixBlendMode?: string;
  order?: 'center' | 'between' | 'start' | 'end';
  flexDirection?: 'row' | 'col';
  alignItems?: 'center' | 'start' | 'end' | 'stretch';
  className?: string;
  innerClassName?: string;
  style?: React.CSSProperties;
}

const ORDER_CLASS: Record<NonNullable<GlassSurfaceProps['order']>, string> = {
  center: 'justify-center',
  between: 'justify-between',
  start: 'justify-start',
  end: 'justify-end',
};

const ALIGN_CLASS: Record<NonNullable<GlassSurfaceProps['alignItems']>, string> = {
  center: 'items-center',
  start: 'items-start',
  end: 'items-end',
  stretch: 'items-stretch',
};

const GlassSurface: React.FC<GlassSurfaceProps> = ({
  children,
  width,
  height,
  borderRadius = 20,
  className = '',
  innerClassName = '',
  style,
  order = 'center',
  flexDirection = 'row',
  alignItems = 'center',
}) => {
  const styleOverride: React.CSSProperties = { borderRadius: `${borderRadius}px`, ...style };
  if (width !== undefined) styleOverride.width = typeof width === 'number' ? `${width}px` : width;
  if (height !== undefined) styleOverride.height = typeof height === 'number' ? `${height}px` : height;

  const flexDir = flexDirection === 'col' ? 'flex-col' : 'flex-row';

  return (
    <div
      className={`relative overflow-hidden border border-white/10 bg-black/40 backdrop-blur-xl backdrop-saturate-150 shadow-xl shadow-black/30 [box-shadow:inset_0_1px_0_0_rgba(255,255,255,0.08),inset_0_-1px_0_0_rgba(255,255,255,0.04),0_8px_24px_0_rgba(0,0,0,0.25)] ${className}`}
      style={styleOverride}
    >
      <div
        className={`w-full h-full p-2 flex ${flexDir} ${ALIGN_CLASS[alignItems]} ${ORDER_CLASS[order]} rounded-[inherit] relative z-10 ${innerClassName}`}
      >
        {children}
      </div>
    </div>
  );
};

export default GlassSurface;
