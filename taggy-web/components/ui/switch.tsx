"use client";

import { Switch as SwitchPrimitive } from "@base-ui/react/switch";
import { cn } from "@/lib/utils";

function Switch({ className, ...props }: SwitchPrimitive.Root.Props) {
  return (
    <SwitchPrimitive.Root
      data-slot="switch"
      className={cn(
        "peer group/switch relative inline-flex h-5 w-9 shrink-0 cursor-pointer items-center rounded-full border transition-colors duration-300 outline-none",
        "border-foreground/25 bg-muted",
        "focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50",
        "data-checked:border-primary data-checked:bg-primary",
        "data-unchecked:border-foreground/30 data-unchecked:bg-foreground/15",
        "hover:border-foreground/45 data-checked:hover:border-primary/80",
        "disabled:cursor-not-allowed disabled:opacity-50",
        className
      )}
      {...props}
    >
      <SwitchPrimitive.Thumb
        data-slot="switch-thumb"
        className={cn(
          "pointer-events-none block size-4 rounded-full bg-background shadow-md ring-1 ring-foreground/15 transition-transform duration-300 ease-out",
          "data-unchecked:translate-x-0.5 data-checked:translate-x-[1.15rem]",
          "group-data-checked/switch:bg-primary-foreground"
        )}
      />
    </SwitchPrimitive.Root>
  );
}

export { Switch };
