"use client";

import { usePathname } from "next/navigation";

type BgVariant = "app" | "landing" | "auth";

function variantForPath(pathname: string): BgVariant {
  if (pathname === "/") return "landing";
  if (
    pathname === "/login" ||
    pathname === "/register" ||
    pathname === "/verify" ||
    pathname.startsWith("/auth/")
  ) {
    return "auth";
  }
  return "app";
}

/** Calm backdrop: solid for app, soft imagery for landing/auth. */
export function PageBackground() {
  const pathname = usePathname();
  const variant = variantForPath(pathname);

  if (variant === "app") {
    return (
      <div
        className="pointer-events-none fixed inset-0 z-0 overflow-hidden bg-background"
        aria-hidden
      >
        <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_top,oklch(0.9_0.03_154/0.35),transparent_55%)] dark:bg-[radial-gradient(ellipse_at_top,oklch(0.35_0.04_154/0.25),transparent_55%)]" />
      </div>
    );
  }

  const src =
    variant === "landing" ? "/images/landing.jpg" : "/images/authpage.jpg";

  return (
    <div
      className="pointer-events-none fixed inset-0 z-0 overflow-hidden"
      aria-hidden
    >
      <div
        className={
          variant === "landing"
            ? "absolute inset-0 bg-cover bg-center bg-no-repeat opacity-[0.78] dark:opacity-[0.62]"
            : "absolute inset-0 bg-cover bg-center bg-no-repeat opacity-[0.18] dark:opacity-[0.14]"
        }
        style={{ backgroundImage: `url(${src})` }}
      />
      <div
        className={
          variant === "landing"
            ? "absolute inset-0 bg-background/22 dark:bg-background/40"
            : "absolute inset-0 bg-background/72 dark:bg-background/78"
        }
      />
      <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_top,oklch(0.63_0.05_154/0.12),transparent_55%)]" />
    </div>
  );
}

/** @deprecated use PageBackground */
export const AmbientBackground = PageBackground;
