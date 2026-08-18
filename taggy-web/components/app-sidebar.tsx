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
import { cn } from "@/lib/utils";

const links = [
  { href: "/home", label: "Home", icon: Home },
  { href: "/skills", label: "Skills", icon: BookOpen },
  { href: "/progress", label: "Progress", icon: TrendingUp },
  { href: "/pods", label: "Pods", icon: Users },
  { href: "/community", label: "Community", icon: MessageCircle },
  { href: "/notifications", label: "Notifications", icon: Bell },
  { href: "/requests", label: "My requests", icon: Inbox },
  { href: "/upgrade", label: "Premium", icon: Sparkles },
];

function isChatSurface(pathname: string) {
  if (/^\/community\/[^/]+\/?$/.test(pathname)) return true;
  if (/^\/pods\/[^/]+\/?$/.test(pathname)) return true;
  return false;
}

function NavLinks({ onNavigate }: { onNavigate?: () => void }) {
  const pathname = usePathname();
  const { isAdmin } = useAuth();
  const nav = isAdmin
    ? [...links, { href: "/admin", label: "Admin console", icon: Shield }]
    : links;

  return (
    <nav className="flex flex-col gap-1">
      {nav.map(({ href, label, icon: Icon }) => {
        const active =
          pathname === href || (href !== "/home" && pathname.startsWith(href));
        const isAdminLink = href === "/admin";
        const isPremium = href === "/upgrade";
        return (
          <Link
            key={href}
            href={href}
            onClick={onNavigate}
            className={cn(
              "inline-flex items-center gap-2.5 rounded-lg px-3 py-2 text-sm transition-colors",
              isPremium && "mt-2 border border-primary/20 bg-primary/5",
              active
                ? isAdminLink
                  ? "bg-amber-500/20 text-foreground"
                  : "bg-secondary text-foreground"
                : isAdminLink
                  ? "text-amber-800 hover:bg-amber-500/10 dark:text-amber-200"
                  : "text-muted-foreground hover:bg-muted hover:text-foreground"
            )}
          >
            <Icon className="size-4 shrink-0" />
            {label}
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
        className="flex shrink-0 items-center gap-2.5 px-1 font-serif text-lg font-medium"
      >
        <span className="inline-flex size-8 items-center justify-center rounded-md bg-primary text-xs font-bold text-primary-foreground">
          T
        </span>
        Taggy
      </Link>

      <div className="min-h-0 flex-1 overflow-y-auto overscroll-contain pr-1">
        <NavLinks onNavigate={onNavigate} />
      </div>

      <div className="shrink-0 space-y-3 border-t border-border/70 pt-4">
        <button
          type="button"
          onClick={() => {
            onNavigate?.();
            router.push(`/u/${username}`);
          }}
          className="flex w-full min-w-0 items-center gap-2 rounded-lg px-1 py-1 text-left hover:bg-muted"
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
        <div className="grid gap-1">
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

  return (
    <div
      className={cn(
        "relative flex bg-transparent",
        chatSurface ? "h-dvh overflow-hidden" : "min-h-dvh"
      )}
    >
      <aside className="relative sticky top-0 z-20 hidden h-dvh w-60 shrink-0 flex-col overflow-hidden border-r border-border/70 bg-sidebar/80 px-3 py-5 text-sidebar-foreground backdrop-blur-md md:flex">
        <div className="relative z-10 flex h-full min-h-0 flex-col">
          <SidebarBody />
        </div>
      </aside>

      <div className="relative z-10 flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden">
        <div className="sticky top-0 z-40 flex h-12 shrink-0 items-center gap-3 border-b border-border/70 bg-background/80 px-3 backdrop-blur-md md:hidden">
          <Button
            variant="ghost"
            size="icon"
            aria-label="Open menu"
            onClick={() => setOpen(true)}
          >
            <Menu className="size-5" />
          </Button>
          <Link href="/home" className="font-serif text-lg font-medium">
            Taggy
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
          className="relative flex w-[min(18rem,85vw)] flex-col overflow-hidden bg-sidebar/80 p-4 backdrop-blur-md"
        >
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
