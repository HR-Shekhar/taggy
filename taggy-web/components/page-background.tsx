"use client";

import { useEffect, useState } from "react";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";

type BgVariant = "app" | "landing" | "auth";

const configs: Record<
  BgVariant,
  {
    src: string;
    tileSize: string;
    opacity: string;
    overlay: string;
    extra?: string;
  }
> = {
  app: {
    src: "/images/global_place_holder.png",
    tileSize: "min(420px, 55vw)",
    opacity: "opacity-[0.62] dark:opacity-[0.48]",
    overlay: "bg-background/62 dark:bg-background/72",
  },
  landing: {
    src: "/images/landing2.png",
    tileSize: "min(520px, 70vw)",
    opacity: "opacity-[0.28] dark:opacity-[0.38]",
    overlay: "bg-background/45 dark:bg-background/55",
    extra:
      "bg-[radial-gradient(ellipse_at_top,oklch(0.8596_0.0291_119.9919/0.18),transparent_55%)] dark:bg-[radial-gradient(ellipse_at_top,oklch(0.3709_0.0248_153.9823/0.1),transparent_55%)]",
  },
  auth: {
    src: "/images/authpage.jpg",
    tileSize: "min(480px, 65vw)",
    opacity: "opacity-[0.22] dark:opacity-[0.16]",
    overlay: "bg-background/58 dark:bg-background/68",
    extra:
      "bg-[radial-gradient(ellipse_at_top,oklch(0.6333_0.0309_154.9039/0.12),transparent_50%)]",
  },
};

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

function scrollOffsetFromEvent(target: EventTarget | null) {
  if (
    !target ||
    target === document ||
    target === document.documentElement ||
    target === document.body
  ) {
    return window.scrollY;
  }
  if (target instanceof HTMLElement && target.scrollHeight > target.clientHeight) {
    return target.scrollTop;
  }
  return window.scrollY;
}

/** Tiled backdrop that moves with the user's scroll and never runs out. */
export function PageBackground() {
  const pathname = usePathname();
  const cfg = configs[variantForPath(pathname)];
  const [offset, setOffset] = useState(0);

  useEffect(() => {
    setOffset(window.scrollY);

    function onScroll(e: Event) {
      setOffset(scrollOffsetFromEvent(e.target));
    }

    window.addEventListener("scroll", onScroll, { capture: true, passive: true });
    return () => window.removeEventListener("scroll", onScroll, true);
  }, [pathname]);

  return (
    <div
      className="pointer-events-none fixed inset-0 z-0 overflow-hidden"
      aria-hidden
    >
      <div
        className={cn("absolute inset-0 bg-repeat", cfg.opacity)}
        style={{
          backgroundImage: `url(${cfg.src})`,
          backgroundSize: cfg.tileSize,
          backgroundPosition: `center ${-offset}px`,
        }}
      />
      <div className={cn("absolute inset-0", cfg.overlay)} />
      {cfg.extra ? <div className={cn("absolute inset-0", cfg.extra)} /> : null}
    </div>
  );
}

/** @deprecated use PageBackground */
export const AmbientBackground = PageBackground;
