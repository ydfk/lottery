import type { ReactNode } from "react";
import { motion, useReducedMotion, type HTMLMotionProps } from "framer-motion";
import { cn } from "@/lib/utils";

type ShimmerButtonProps = Omit<HTMLMotionProps<"button">, "children"> & {
  children?: ReactNode;
};

export function ShimmerButton({ className, children, disabled, ...props }: ShimmerButtonProps) {
  const reduceMotion = useReducedMotion();

  return (
    <motion.button
      className={cn(
        "relative inline-flex items-center justify-center overflow-hidden rounded-md border border-border bg-foreground px-4 py-2 text-sm font-medium text-background shadow-sm transition-colors",
        "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
        "disabled:cursor-not-allowed disabled:opacity-60",
        className
      )}
      disabled={disabled}
      {...props}
      initial={false}
      animate={reduceMotion || disabled ? {} : { backgroundPosition: ["0% 50%", "100% 50%"] }}
      transition={
        reduceMotion || disabled ? undefined : { duration: 2, ease: "linear", repeat: Infinity }
      }
      style={
        reduceMotion || disabled
          ? undefined
          : {
              backgroundImage:
                "linear-gradient(120deg, rgba(255,255,255,0.15) 0%, rgba(255,255,255,0.35) 20%, rgba(255,255,255,0.15) 40%)",
              backgroundSize: "200% 100%",
            }
      }
    >
      <span className="relative z-10">{children}</span>
    </motion.button>
  );
}
