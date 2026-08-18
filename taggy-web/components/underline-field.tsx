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
    <div className="space-y-1.5">
      <label
        htmlFor={id}
        className="block text-[10px] font-medium uppercase tracking-[0.2em] text-muted-foreground"
      >
        {label}
      </label>
      <input
        id={id}
        className={cn(
          "w-full border-0 border-b border-foreground/25 bg-transparent pb-1.5 text-sm text-foreground outline-none transition-colors",
          "placeholder:text-muted-foreground/60 focus:border-primary dark:border-foreground/30 dark:placeholder:text-muted-foreground/50",
          className
        )}
        {...props}
      />
    </div>
  );
}
