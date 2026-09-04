"use client";

import Link from "next/link";
import { Check, Sparkles } from "lucide-react";
import { PageHeader } from "@/components/app-ui";
import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";

const freeFeatures = [
  "Multiple skills",
  "Roadmaps & milestones",
  "Pods & community chat",
  "Study streaks & progress",
];

const premiumFeatures = [
  "Everything in Free",
  "Priority skill requests",
  "One-time unlock (coming soon)",
];

export default function UpgradePage() {
  return (
    <div className="mx-auto max-w-4xl space-y-8">
      <PageHeader
        title="Taggy Premium"
        description="Support Taggy and get extras as they ship."
        backHref="/home"
      />

      <div className="grid gap-4 md:grid-cols-2">
        <div className="rounded-xl border border-border bg-card p-6">
          <p className="text-sm font-medium text-foreground/75">Free</p>
          <h2 className="mt-1 font-serif text-2xl">Starter</h2>
          <p className="mt-2 text-sm text-foreground/75">
            Join any skills you want and learn at your own pace.
          </p>
          <ul className="mt-6 space-y-2.5">
            {freeFeatures.map((f) => (
              <li key={f} className="flex items-start gap-2 text-sm">
                <Check className="mt-0.5 size-4 shrink-0 text-primary" />
                {f}
              </li>
            ))}
          </ul>
          <Link
            href="/skills"
            className={cn(
              buttonVariants({ variant: "outline" }),
              "mt-6 w-full"
            )}
          >
            Continue free
          </Link>
        </div>

        <div className="relative rounded-xl border border-primary/30 bg-primary/5 p-6">
          <span className="absolute right-4 top-4 inline-flex items-center gap-1 rounded-full bg-primary/15 px-2.5 py-0.5 text-xs font-medium text-primary">
            <Sparkles className="size-3" />
            Soon
          </span>
          <p className="text-sm font-medium text-primary">Premium</p>
          <h2 className="mt-1 font-serif text-2xl">Unlimited</h2>
          <p className="mt-2 text-sm text-foreground/75">
            Billing is finishing up. Premium extras will unlock with a
            one-time purchase.
          </p>
          <ul className="mt-6 space-y-2.5">
            {premiumFeatures.map((f) => (
              <li key={f} className="flex items-start gap-2 text-sm">
                <Check className="mt-0.5 size-4 shrink-0 text-primary" />
                {f}
              </li>
            ))}
          </ul>
          <button
            type="button"
            disabled
            className={cn(buttonVariants(), "mt-6 w-full opacity-60")}
          >
            Checkout coming soon
          </button>
        </div>
      </div>
    </div>
  );
}
