
"use client";
import * as React from "react";
import { cn } from "@/lib/utils";
import { useMotionTemplate, useMotionValue, motion } from "motion/react";
import { Eye, EyeOff } from "lucide-react";

export interface InputProps
  extends React.InputHTMLAttributes<HTMLInputElement> { }

const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, type, ...props }, ref) => {
    const radius = 100;
    const [visible, setVisible] = React.useState(false);
    const [showPassword, setShowPassword] = React.useState(false);

    const isPassword = type === "password";
    const inputType = isPassword ? (showPassword ? "text" : "password") : type;

    const mouseX = useMotionValue(0);
    const mouseY = useMotionValue(0);

    function handleMouseMove({ currentTarget, clientX, clientY }: any) {
      const { left, top } = currentTarget.getBoundingClientRect();

      mouseX.set(clientX - left);
      mouseY.set(clientY - top);
    }
    return (
      <motion.div
        style={{
          background: useMotionTemplate`
        radial-gradient(
          ${visible ? radius + "px" : "0px"} circle at ${mouseX}px ${mouseY}px,
          #E2434B,
          transparent 80%
        )
      `,
        }}
        onMouseMove={handleMouseMove}
        onMouseEnter={() => setVisible(true)}
        onMouseLeave={() => setVisible(false)}
        className="group/input rounded-lg p-0.5 transition duration-300"
      >
        <div className="relative">
          <input
            type={inputType}
            className={cn(
              `pointer-events-auto shadow-input dark:placeholder-text-neutral-600 flex h-10 w-full
              rounded-md border-none bg-gray-50 px-3 py-2 text-sm text-black transition duration-400
              group-hover/input:shadow-none file:border-0 file:bg-transparent file:text-sm file:font-medium
              placeholder:text-neutral-400 focus-visible:ring-2 focus-visible:ring-neutral-400
              focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50 dark:bg-zinc-800
              dark:text-white dark:shadow-[0px_0px_1px_1px_#404040] dark:focus-visible:ring-neutral-600`,
              isPassword && "pr-10",
              className,
            )}
            ref={ref}
            {...props}
          />
          {isPassword && (
            <button
              type="button"
              onClick={() => setShowPassword((s) => !s)}
              onMouseDown={(e) => e.preventDefault()}
              aria-label={showPassword ? "Hide password" : "Show password"}
              aria-pressed={showPassword}
              className="pointer-events-auto absolute inset-y-0 right-0 flex cursor-pointer items-center pr-3 text-neutral-400 transition-colors hover:text-neutral-600 dark:hover:text-neutral-200"
            >
              {showPassword ? <EyeOff size={16} /> : <Eye size={16} />}
            </button>
          )}
        </div>
      </motion.div>
    );
  },
);
Input.displayName = "Input";

export { Input };
