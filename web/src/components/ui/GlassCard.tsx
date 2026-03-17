import GlassSurface from "@/components/primitives/GlassSurface";
import React from "react";

export function GlassCard({ children, title, icon: Icon, flexCol = false }: { children: React.ReactNode, title: string, icon?: any, flexCol?: boolean }) {
  return (
    <div className="w-full h-full relative pointer-events-auto">
      <GlassSurface
        displace={10}
        distortionScale={-40}
        redOffset={2}
        greenOffset={5}
        blueOffset={8}
        brightness={15}
        opacity={0.03}
        mixBlendMode="overlay"
        flexDirection="col"
        alignItems="stretch"
        order="start"
        className="h-full w-full"
      >
        <div className={`w-full flex ${flexCol ? 'flex-col gap-y-4' : 'flex-col'} p-3 sm:p-5 h-full`}>
          <div className="flex items-center gap-x-3 mb-4 border-b border-white/10 pb-4">
            <h2 className="text-lg sm:text-xl font-medium text-white/90">{title}</h2>
          </div>
          <div className={`flex-1 flex ${flexCol ? 'flex-col gap-y-5' : 'flex-col'} justify-start`}>
            {children}
          </div>
        </div>
      </GlassSurface>
    </div>
  );
}

export function SettingRow({ label, description, rightElement }: { label: string, description: string, rightElement: React.ReactNode }) {
  return (
    <div className="flex items-center justify-between w-full py-2">
      <div className="flex flex-col gap-y-1 max-w-[65%]">
        <span className="text-sm font-medium text-white/80">{label}</span>
        <span className="text-xs text-white/40 leading-snug">{description}</span>
      </div>
      <div className="shrink-0">{rightElement}</div>
    </div>
  );
}
