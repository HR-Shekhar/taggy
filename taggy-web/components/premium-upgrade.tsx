"use client";

import Link from "next/link";
import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { Sparkles } from "lucide-react";

/** Soft note when free users hit the 1-skill join limit. */
export function PremiumUpgradePrompt({
  message,
}: {
  message?: string | null;
  onUpgraded?: () => void;
}) {
  if (!message) return null;

  return (
    <div className="flex flex-col gap-3 rounded-xl border border-primary/25 bg-primary/5 px-4 py-4 sm:flex-row sm:items-center sm:justify-between">
      <div className="flex gap-3">
        <span className="inline-flex size-9 shrink-0 items-center justify-center rounded-lg bg-primary/15 text-primary">
          <Sparkles className="size-4" />
        </span>
        <div>
          <p className="font-medium text-foreground">Free plan limit</p>
          <p className="mt-0.5 text-sm text-muted-foreground">
            Free accounts can follow one skill at a time. Upgrade for unlimited
            skills and pods.
          </p>
        </div>
      </div>
      <Link
        href="/upgrade"
        className={cn(buttonVariants({ size: "sm" }), "shrink-0")}
      >
        View Premium
      </Link>
    </div>
  );
}

/** @deprecated Checkout removed; kept for any stale imports. */
export async function startPremiumCheckout(): Promise<{
  ok: boolean;
  subscription?: string;
  error?: string;
}> {
  return {
    ok: false,
    error: "Premium checkout is coming soon.",
  };
}

export function PremiumCheckoutError({ message }: { message: string | null }) {
  if (!message) return null;
  return null;
}
