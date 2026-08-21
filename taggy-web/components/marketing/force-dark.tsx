"use client";

/**
 * Visually forces dark appearance for the landing page only.
 * Does not call next-themes / localStorage, so app light/dark switching stays intact.
 */
export function ForceDark({ children }: { children: React.ReactNode }) {
  return (
    <div className="dark min-h-dvh bg-background text-foreground">{children}</div>
  );
}
