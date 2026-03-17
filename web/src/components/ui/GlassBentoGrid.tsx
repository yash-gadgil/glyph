import { cn } from "@/lib/utils";
import { ReactNode } from "react";
import GlassSurface from "../primitives/GlassSurface";



interface GlassBentoGridProps {
  children: ReactNode;
  className?: string;
}

export function GlassBentoGrid({
  children,
  className
}: GlassBentoGridProps) {

  return (
    <div className={cn(
      "mx-auto grid max-w-7xl w-full grid-cols-1 gap-4 md:auto-rows-[18rem] md:grid-cols-3", className)}>

      {children}

    </div>
  );
}


interface BentoGridItemProps {
  className?: string;
  header?: ReactNode;
  title?: string | ReactNode;
  description?: string | ReactNode;
  icon?: ReactNode;
}

export function BentoGridItem({
  className,
  header,
  title,
  description,
  icon
}: BentoGridItemProps) {


  return (

    <GlassSurface
      displace={15}
      distortionScale={-150}
      redOffset={5}
      greenOffset={15}
      blueOffset={25}
      brightness={60}
      opacity={0.8}
      mixBlendMode="overlay"
      order="between"
      flexDirection="col"
      alignItems="stretch"
      className={cn("group/bento shadow-input row-span-1 flex-col space-y-4 rounded-xl border border-neutral-200 p-4 transition duration-200 hover:shadow-xl dark:border-white/20 dark:shadow-none", className)}>

      {header}
      <div className="transition duration-200 group-hover/bento:translate-x-2">
        {icon}
        <div className="mt-2 mb-2">
          {title}
          <div className=" text-xs ">
            {description}
          </div>
        </div>
      </div>

    </GlassSurface>

  );
}