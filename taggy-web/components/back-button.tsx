"use client";

import { useRouter, useSearchParams } from "next/navigation";
import { Suspense } from "react";
import { ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

type BackButtonProps = {
  /** Used when there is no `?from=` and browser history is empty. */
  fallbackHref?: string;
  label?: string;
  className?: string;
  size?: "default" | "sm" | "xs";
  variant?: "outline" | "ghost";
};

function safeFrom(raw: string | null): string | null {
  if (!raw) return null;
  // Only allow same-origin relative paths
  if (!raw.startsWith("/") || raw.startsWith("//")) return null;
  return raw;
}

function BackButtonInner({
  fallbackHref = "/home",
  label = "Back",
  className,
  size = "sm",
  variant = "outline",
}: BackButtonProps) {
  const router = useRouter();
  const params = useSearchParams();
  const from = safeFrom(params.get("from"));

  function goBack() {
    if (from) {
      router.push(from);
      return;
    }
    if (typeof window !== "undefined" && window.history.length > 1) {
      router.back();
      return;
    }
    router.push(fallbackHref);
  }

  return (
    <Button
      type="button"
      variant={variant}
      size={size}
      onClick={goBack}
      className={cn("gap-1.5", className)}
    >
      <ArrowLeft className="size-4" />
      {label}
    </Button>
  );
}

/** Back control: prefers `?from=`, then history, then fallback. */
export function BackButton(props: BackButtonProps) {
  return (
    <Suspense
      fallback={
        <Button
          type="button"
          variant={props.variant ?? "outline"}
          size={props.size ?? "sm"}
          className={cn("gap-1.5", props.className)}
          disabled
        >
          <ArrowLeft className="size-4" />
          {props.label ?? "Back"}
        </Button>
      }
    >
      <BackButtonInner {...props} />
    </Suspense>
  );
}
