"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useState, type ReactNode } from "react";
import {
  Bell,
  BookOpen,
  Home,
  LogOut,
  Menu,
  MessageCircle,
  Settings,
  Shield,
  Sparkles,
  TrendingUp,
  Users,
  Inbox,
} from "lucide-react";
import { useAuth } from "@/lib/auth";
import { UserAvatar } from "@/components/user-avatar";
import { Button } from "@/components/ui/button";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { ThemeModeSwitch } from "@/components/theme-mode-switch";
import { BrandLogo } from "@/components/brand-logo";
import { useUnreadNotifications } from "@/lib/use-unread-notifications";
import { cn } from "@/lib/utils";

const links = [
  { href: "/home", label: "Home", icon: Home, tour: "nav-home" },
  { href: "/skills", label: "Skills", icon: BookOpen, tour: "nav-skills" },
  { href: "/progress", label: "Progress", icon: TrendingUp, tour: "nav-progress" },
  { href: "/pods", label: "Pods", icon: Users, tour: "nav-pods" },
  {
    href: "/community",
    label: "Community",
    icon: MessageCircle,
    tour: "nav-community",
  },
  {
    href: "/notifications",
    label: "Notifications",
    icon: Bell,
    tour: "nav-notifications",
  },
  { href: "/requests", label: "My requests", icon: Inbox, tour: "nav-requests" },
  { href: "/upgrade", label: "Premium", icon: Sparkles, tour: "nav-premium" },
];

function isChatSurface(pathname: string) {
  if (/^\/community\/[^/]+\/?$/.test(pathname)) return true;
  if (/^\/pods\/[^/]+\/?$/.test(pathname)) return true;
  return false;
}

function SidebarBackground() {
  return (
    <>
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 bg-[url('/images/sidebar.jpg')] bg-cover bg-bottom bg-no-repeat opacity-18 dark:opacity-12"
      />
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 bg-gradient-to-t from-sidebar via-sidebar/98 to-sidebar"
      />
    </>
  );
}

function UnreadBadge({ count }: { count: number }) {
  if (count <= 0) return null;
  const label = count > 9 ? "9+" : String(count);
  return (
    <span className="ml-auto inline-flex min-w-5 items-center justify-center rounded-full bg-primary px-1.5 py-0.5 text-[10px] font-semibold tabular-nums text-primary-foreground">
      {label}
    </span>
  );
}

function NavLinks({ onNavigate }: { onNavigate?: () => void }) {
  const pathname = usePathname();
  const { isAdmin } = useAuth();
  const { count: unread } = useUnreadNotifications();
  const nav = isAdmin
    ? [...links, { href: "/admin", label: "Admin console", icon: Shield, tour: "nav-admin" }]
    : links;

  return (
    <nav className="flex flex-col gap-0.5">
      {nav.map(({ href, label, icon: Icon, tour }) => {
        const active =
          pathname === href || (href !== "/home" && pathname.startsWith(href));
        const isAdminLink = href === "/admin";
        const isPremium = href === "/upgrade";
        const isNotifications = href === "/notifications";
        return (
          <Link
            key={href}
            href={href}
            data-tour={tour}
            onClick={onNavigate}
            className={cn(
              "inline-flex items-center gap-2.5 rounded-lg px-3 py-2 text-sm font-medium transition-colors",
              isPremium && "mt-2",
              active
                ? isAdminLink
                  ? "bg-amber-500/20 text-foreground"
                  : "bg-primary/15 text-foreground"
                : isAdminLink
                  ? "text-amber-800 hover:bg-amber-500/10 dark:text-amber-200"
                  : "text-sidebar-foreground/85 hover:bg-sidebar-accent hover:text-sidebar-foreground"
            )}
          >
            <span className="relative shrink-0">
              <Icon className="size-4" />
              {isNotifications && unread > 0 ? (
                <span
                  aria-hidden
                  className="absolute -right-0.5 -top-0.5 size-2 rounded-full bg-primary ring-2 ring-sidebar"
                />
              ) : null}
            </span>
            <span className="min-w-0 flex-1 truncate">{label}</span>
            {isNotifications ? <UnreadBadge count={unread} /> : null}
          </Link>
        );
      })}
    </nav>
  );
}

function SidebarBody({ onNavigate }: { onNavigate?: () => void }) {
  const { username, logout, avatarUrl } = useAuth();
  const router = useRouter();

  return (
    <div className="flex h-full min-h-0 flex-col gap-4">
      <Link
        href="/home"
        onClick={onNavigate}
        className="flex shrink-0 px-1 text-lg"
      >
        <BrandLogo size={32} wordmarkClassName="text-lg" />
      </Link>

      <div className="min-h-0 flex-1 overflow-y-auto overscroll-contain pr-1">
        <NavLinks onNavigate={onNavigate} />
      </div>

      <div className="shrink-0 space-y-3 border-t border-sidebar-border pt-4">
        <button
          type="button"
          onClick={() => {
            onNavigate?.();
            router.push(`/u/${username}`);
          }}
          className="flex w-full min-w-0 items-center gap-2 rounded-lg px-2 py-1.5 text-left hover:bg-sidebar-accent"
        >
          <UserAvatar
            username={username}
            src={avatarUrl}
            className="size-8"
            fallbackClassName="text-xs"
          />
          <span className="truncate text-sm">@{username}</span>
        </button>
        <ThemeModeSwitch />
        <div className="grid gap-0.5">
          <Button
            variant="ghost"
            className="justify-start gap-2"
            onClick={() => {
              onNavigate?.();
              router.push("/settings");
            }}
          >
            <Settings className="size-4" />
            Settings
          </Button>
          <Button
            variant="ghost"
            className="justify-start gap-2 text-destructive hover:text-destructive"
            onClick={async () => {
              await logout();
              router.replace("/login");
            }}
          >
            <LogOut className="size-4" />
            Log out
          </Button>
        </div>
      </div>
    </div>
  );
}

export function AppShell({ children }: { children: ReactNode }) {
  const [open, setOpen] = useState(false);
  const pathname = usePathname();
  const chatSurface = isChatSurface(pathname);
  const { count: unread } = useUnreadNotifications();

  return (
    <div
      className={cn(
        "relative flex bg-transparent",
        chatSurface ? "h-dvh overflow-hidden" : "min-h-dvh"
      )}
    >
      <aside className="relative sticky top-0 z-20 hidden h-dvh w-60 shrink-0 flex-col overflow-hidden border-r border-sidebar-border bg-sidebar px-3 py-5 text-sidebar-foreground md:flex">
        <SidebarBackground />
        <div className="relative z-10 flex h-full min-h-0 flex-col">
          <SidebarBody />
        </div>
      </aside>

      <div className="relative z-10 flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden">
        <div className="sticky top-0 z-40 flex h-12 shrink-0 items-center gap-3 border-b border-border bg-background/95 px-3 md:hidden">
          <Button
            variant="ghost"
            size="icon"
            aria-label="Open menu"
            onClick={() => setOpen(true)}
          >
            <Menu className="size-5" />
          </Button>
          <Link href="/home" className="text-lg">
            <BrandLogo size={28} wordmarkClassName="text-lg" />
          </Link>
          <Link
            href="/notifications"
            className="relative ml-auto inline-flex size-9 items-center justify-center rounded-md hover:bg-muted"
            aria-label={
              unread > 0
                ? `Notifications, ${unread} unread`
                : "Notifications"
            }
          >
            <Bell className="size-5" />
            {unread > 0 ? (
              <span className="absolute right-1 top-1 flex size-4 items-center justify-center rounded-full bg-primary text-[9px] font-bold text-primary-foreground">
                {unread > 9 ? "9+" : unread}
              </span>
            ) : null}
          </Link>
        </div>

        <main
          className={cn(
            "min-h-0 flex-1",
            chatSurface
              ? "flex flex-col overflow-hidden p-0"
              : "mx-auto w-full max-w-6xl overflow-y-auto px-4 py-5 sm:px-6 sm:py-6 lg:px-8 lg:py-7"
          )}
        >
          {children}
        </main>
      </div>

      <Sheet open={open} onOpenChange={setOpen}>
        <SheetContent
          side="left"
          className="relative flex w-[min(18rem,85vw)] flex-col overflow-hidden bg-sidebar p-4 text-sidebar-foreground"
        >
          <SidebarBackground />
          <div className="relative z-10 flex h-full min-h-0 flex-col">
            <SheetHeader className="sr-only">
              <SheetTitle>Navigation</SheetTitle>
            </SheetHeader>
            <SidebarBody onNavigate={() => setOpen(false)} />
          </div>
        </SheetContent>
      </Sheet>
    </div>
  );
}
