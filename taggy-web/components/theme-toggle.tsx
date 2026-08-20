"use client";

import { Moon, Sun } from "lucide-react";
import { useTheme } from "next-themes";
import { useEffect, useState } from "react";
import { withThemeTransition } from "@/lib/theme-transition";
import { cn } from "@/lib/utils";

/** One click flips light ↔ dark. */
export function ThemeToggle({ className }: { className?: string }) {
  const { resolvedTheme, setTheme } = useTheme();
  const [mounted, setMounted] = useState(false);
  const [dark, setDark] = useState(false);

  useEffect(() => {
    setMounted(true);
    setDark(resolvedTheme === "dark");
  }, [resolvedTheme]);

  return (
    <button
      type="button"
      disabled={!mounted}
      aria-label={dark ? "Switch to light mode" : "Switch to dark mode"}
      className={cn(
        "inline-flex size-7 cursor-pointer items-center justify-center rounded-full border border-foreground/20 bg-background/70 text-muted-foreground outline-none transition-[color,background-color,border-color,box-shadow] duration-200",
        "hover:border-foreground/40 hover:bg-muted hover:text-foreground hover:shadow-sm",
        "focus-visible:ring-3 focus-visible:ring-ring/50",
        "disabled:cursor-not-allowed disabled:opacity-50",
        className
      )}
      onClick={() => {
        const nextDark = !dark;
        setDark(nextDark);
        withThemeTransition(() => setTheme(nextDark ? "dark" : "light"));
      }}
    >
      <span className="relative inline-flex size-4 items-center justify-center">
        <Sun className="size-4 scale-100 rotate-0 transition-all duration-300 dark:scale-0 dark:-rotate-90" />
        <Moon className="absolute size-4 scale-0 rotate-90 transition-all duration-300 dark:scale-100 dark:rotate-0" />
      </span>
    </button>
  );
}
