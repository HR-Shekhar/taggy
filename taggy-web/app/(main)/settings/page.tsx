"use client";

import Link from "next/link";
import { useAuth } from "@/lib/auth";
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

export default function SettingsPage() {
  const { username } = useAuth();

  return (
    <div className="space-y-6">
      <PageHeader
        title="Settings"
        description="Account shortcuts for your Taggy profile."
      />
      <Card className="rounded-xl ring-1 ring-foreground/10">
        <CardHeader>
          <CardTitle className="font-serif text-lg">Account</CardTitle>
          <CardDescription>
            Signed in as <span className="font-medium">@{username}</span>
          </CardDescription>
        </CardHeader>
        <CardContent className="flex flex-wrap gap-2">
          <Link
            href={`/u/${username}`}
            className={cn(buttonVariants())}
          >
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
            href="/dev"
            className={cn(buttonVariants({ variant: "outline" }))}
          >
            API tester
          </Link>
        </CardContent>
      </Card>
    </div>
  );
}
