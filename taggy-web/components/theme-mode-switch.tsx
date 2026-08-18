"use client";

import { Moon, Sun } from "lucide-react";
import { useTheme } from "next-themes";
import { useEffect, useState } from "react";
import { withThemeTransition } from "@/lib/theme-transition";
import { cn } from "@/lib/utils";

/** Light / dark mode switch for the app sidebar. One click flips the theme. */
export function ThemeModeSwitch({ className }: { className?: string }) {
  const { resolvedTheme, setTheme } = useTheme();
  const [mounted, setMounted] = useState(false);
  const [dark, setDark] = useState(false);

  useEffect(() => {
    setMounted(true);
    setDark(resolvedTheme === "dark");
  }, [resolvedTheme]);

  const isDark = mounted && dark;

  return (
    <button
      type="button"
      disabled={!mounted}
      aria-pressed={isDark}
      aria-label={isDark ? "Switch to light mode" : "Switch to dark mode"}
      className={cn(
        "flex w-full cursor-pointer items-center justify-between gap-3 rounded-lg border border-border/60 bg-background/50 px-2.5 py-2 text-left transition-[background-color,border-color] duration-300",
        "hover:bg-muted/60",
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
      <div className="flex items-center gap-2 text-sm text-foreground">
        <span className="relative inline-flex size-4 items-center justify-center">
          <Sun
            className={cn(
              "size-4 text-primary transition-all duration-300",
              isDark ? "scale-0 rotate-90 opacity-0" : "scale-100 rotate-0 opacity-100"
            )}
          />
          <Moon
            className={cn(
              "absolute size-4 text-primary transition-all duration-300",
              isDark ? "scale-100 rotate-0 opacity-100" : "scale-0 -rotate-90 opacity-0"
            )}
          />
        </span>
        <span className="font-medium">{isDark ? "Dark mode" : "Light mode"}</span>
      </div>
      <span
        aria-hidden
        className={cn(
          "relative inline-flex h-5 w-9 shrink-0 items-center rounded-full border transition-colors duration-300",
          isDark
            ? "border-primary bg-primary"
            : "border-foreground/30 bg-foreground/15"
        )}
      >
        <span
          className={cn(
            "block size-4 rounded-full bg-background shadow-md ring-1 ring-foreground/15 transition-transform duration-300 ease-out",
            isDark ? "translate-x-[1.15rem] bg-primary-foreground" : "translate-x-0.5"
          )}
        />
      </span>
    </button>
  );
}
