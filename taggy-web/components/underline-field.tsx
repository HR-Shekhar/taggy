"use client";

import { cn } from "@/lib/utils";
import type { ComponentProps } from "react";

type UnderlineFieldProps = ComponentProps<"input"> & {
  label: string;
};

export function UnderlineField({
  id,
  label,
  className,
  ...props
}: UnderlineFieldProps) {
  return (
    <div className="space-y-1">
      <label
        htmlFor={id}
        className="block text-[11px] font-medium uppercase tracking-[0.16em] text-foreground/60"
      >
        {label}
      </label>
      <input
        id={id}
        className={cn(
          "w-full border-0 border-b border-foreground/25 bg-transparent py-2 text-[15px] text-foreground outline-none transition-colors",
          "placeholder:text-foreground/50 focus:border-primary dark:border-foreground/30 dark:placeholder:text-foreground/40",
          className
        )}
        {...props}
      />
    </div>
  );
}
