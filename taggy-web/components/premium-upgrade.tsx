"use client";

import Link from "next/link";
import { useEffect } from "react";
import { buttonVariants } from "@/components/ui/button";
import { toastError } from "@/lib/toast";
import { cn } from "@/lib/utils";

/** Soft note when free users hit the 1-skill join limit. */
export function PremiumUpgradePrompt({
  message,
}: {
  message?: string | null;
  onUpgraded?: () => void;
}) {
  if (!message) return null;

  return (
    <div className="space-y-2 rounded-lg border border-border/80 bg-muted/40 px-3 py-3 text-sm">
      <p className="font-medium text-foreground">Free plan limit</p>
      <p className="text-muted-foreground">
        Free accounts can follow one skill at a time. Taggy Premium (unlimited
        skills) is coming soon.
      </p>
      <Link
        href="/upgrade"
        className={cn(buttonVariants({ size: "sm" }), "inline-flex")}
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
  useEffect(() => {
    if (message) toastError(message, "Couldn't start checkout");
  }, [message]);
  return null;
}
