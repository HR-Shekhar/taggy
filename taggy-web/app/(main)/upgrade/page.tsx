"use client";

import Link from "next/link";
import { PageHeader } from "@/components/app-ui";
import { buttonVariants } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { cn } from "@/lib/utils";

export default function UpgradePage() {
  return (
    <div className="mx-auto max-w-lg space-y-6">
      <PageHeader
        title="Taggy Premium"
        description="Unlimited skills and more — launching soon."
      />

      <Card className="rounded-xl ring-1 ring-foreground/10">
        <CardHeader>
          <CardTitle className="font-serif text-2xl">Coming soon</CardTitle>
          <CardDescription>
            We&apos;re finishing Premium billing. For now, free accounts can
            follow one active skill. Premium will unlock unlimited skill
            enrollments.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          <ul className="list-inside list-disc space-y-1 text-sm text-muted-foreground">
            <li>Unlimited active skills</li>
            <li>One-time unlock (Razorpay)</li>
            <li>Same roadmaps, pods, and communities</li>
          </ul>
          <Link href="/skills" className={cn(buttonVariants())}>
            Back to skills
          </Link>
        </CardContent>
      </Card>
    </div>
  );
}
