"use client";

/** Briefly enables CSS theme color transitions while next-themes swaps classes. */
export function withThemeTransition(apply: () => void) {
  if (typeof document === "undefined") {
    apply();
    return;
  }

  const root = document.documentElement;
  root.classList.add("theme-transition");
  apply();

  window.setTimeout(() => {
    root.classList.remove("theme-transition");
  }, 420);
}
