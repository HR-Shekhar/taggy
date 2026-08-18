"use client";

import Link from "next/link";
import { useEffect } from "react";
import { usePathname, useRouter } from "next/navigation";
import { Shield } from "lucide-react";
import { AppShell } from "@/components/app-sidebar";
import { Loading } from "@/components/app-ui";
import { useAuth } from "@/lib/auth";

/**
 * Admin-only route group. Access rules:
 * 1. Must be logged in
 * 2. Must have is_admin from auth/profile (backend also enforces /admin/* APIs)
 * Non-admins are redirected to /home; the Admin nav link is hidden for them.
 */
export default function AdminLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { ready, isAuthenticated, isAdmin, username } = useAuth();
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    if (!ready) return;
    if (!isAuthenticated) {
      router.replace(`/login?next=${encodeURIComponent(pathname)}`);
      return;
    }
    if (!isAdmin) {
      router.replace("/home");
    }
  }, [ready, isAuthenticated, isAdmin, router, pathname]);

  if (!ready) {
    return (
      <div className="flex min-h-dvh items-center justify-center">
        <Loading />
      </div>
    );
  }
  if (!isAuthenticated || !isAdmin) return null;

  return (
    <AppShell>
      <div className="mb-6 overflow-hidden rounded-xl border border-amber-500/30 bg-gradient-to-r from-amber-500/15 via-background to-background px-4 py-3 sm:px-5">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="flex items-center gap-3">
            <span className="inline-flex size-9 items-center justify-center rounded-lg bg-amber-500/20 text-amber-700 dark:text-amber-300">
              <Shield className="size-4" />
            </span>
            <div>
              <p className="font-serif text-base font-medium tracking-tight">
                Admin console
              </p>
              <p className="text-xs text-muted-foreground">
                Approvals only — signed in as @{username}
              </p>
            </div>
          </div>
          <Link
            href="/home"
            className="text-sm text-muted-foreground underline-offset-4 hover:underline"
          >
            Back to app
          </Link>
        </div>
      </div>
      {children}
    </AppShell>
  );
}
