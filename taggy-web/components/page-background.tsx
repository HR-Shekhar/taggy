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

/** Clean backdrop: solid app surface; light photo wash on landing/auth only. */
export function PageBackground() {
  const pathname = usePathname();
  const variant = variantForPath(pathname);

  if (variant === "app") {
    return (
      <div
        className="pointer-events-none fixed inset-0 z-0 overflow-hidden bg-background"
        aria-hidden
      >
        <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_top,var(--color-primary)_0%,transparent_60%)] opacity-[0.03] dark:opacity-[0.02]" />
      </div>
    );
  }

  const src =
    variant === "landing" ? "/images/landing.jpg" : "/images/authpage.jpg";

  return (
    <div
      className="pointer-events-none fixed inset-0 z-0 overflow-hidden bg-background"
      aria-hidden
    >
      <div
        className={
          variant === "landing"
            ? "absolute inset-0 bg-cover bg-center bg-no-repeat opacity-[0.28] [filter:brightness(1.45)_saturate(0.85)_sepia(0.12)_hue-rotate(35deg)] dark:opacity-[0.4] dark:[filter:brightness(0.95)_saturate(0.9)]"
            : "absolute inset-0 bg-cover bg-center bg-no-repeat opacity-[0.16] dark:opacity-[0.12]"
        }
        style={{ backgroundImage: `url(${src})` }}
      />
      <div
        className={
          variant === "landing"
            ? "absolute inset-0 bg-background/70 dark:bg-background/55"
            : "absolute inset-0 bg-background/75 dark:bg-background/80"
        }
      />
    </div>
  );
}

/** @deprecated use PageBackground */
export const AmbientBackground = PageBackground;
