"use client";

import { useLayoutEffect, useRef } from "react";
import { useTheme } from "next-themes";

/** Forces dark mode while the landing page is mounted; restores prior theme on leave. */
export function ForceDark({ children }: { children: React.ReactNode }) {
  const { setTheme } = useTheme();
  const previousRef = useRef<string | null>(null);

  useLayoutEffect(() => {
    previousRef.current = window.localStorage.getItem("theme") ?? "system";
    setTheme("dark");
    document.documentElement.classList.add("dark");
    return () => {
      setTheme(previousRef.current ?? "system");
    };
  }, [setTheme]);

  return <>{children}</>;
}
