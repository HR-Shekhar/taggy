"use client";

import Link from "next/link";
import { useAuth } from "@/lib/auth";
import { PageHeader, Section } from "@/components/app-ui";
import { ThemeModeSwitch } from "@/components/theme-mode-switch";
import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";

export default function SettingsPage() {
  const { username } = useAuth();

  return (
    <div className="mx-auto max-w-2xl space-y-8">
      <PageHeader
        title="Settings"
        description="Account, appearance, and shortcuts."
        backHref="/home"
      />

      <Section title="Account" description={`Signed in as @${username}`}>
        <div className="flex flex-wrap gap-2 rounded-xl border border-border bg-card p-4">
          <Link href={`/u/${username}`} className={cn(buttonVariants())}>
            Edit profile
          </Link>
          <Link
            href="/requests"
            className={cn(buttonVariants({ variant: "outline" }))}
          >
            My requests
          </Link>
          <Link
            href="/reports"
            className={cn(buttonVariants({ variant: "outline" }))}
          >
            My reports
          </Link>
          <Link
            href="/notifications"
            className={cn(buttonVariants({ variant: "outline" }))}
          >
            Notifications
          </Link>
        </div>
      </Section>

      <Section
        title="Appearance"
        description="Switch between light and dark mode."
      >
        <div className="rounded-xl border border-border bg-card p-4">
          <ThemeModeSwitch />
        </div>
      </Section>

      <Section title="Developer" description="Tools for API exploration.">
        <div className="rounded-xl border border-border bg-card p-4">
          <Link
            href="/dev"
            className={cn(buttonVariants({ variant: "outline" }))}
          >
            API tester
          </Link>
        </div>
      </Section>
    </div>
  );
}
