"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import {
  Bell,
  BookOpen,
  Home,
  LogOut,
  Search,
  Settings,
  TrendingUp,
  Users,
} from "lucide-react";
import { useAuth } from "@/lib/auth";
import { BrandLogo } from "@/components/brand-logo";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { cn } from "@/lib/utils";
import { ThemeToggle } from "@/components/theme-toggle";
import { useUnreadNotifications } from "@/lib/use-unread-notifications";

const links = [
  { href: "/home", label: "Home", icon: Home },
  { href: "/skills", label: "Skills", icon: BookOpen },
  { href: "/progress", label: "Progress", icon: TrendingUp },
  { href: "/pods", label: "Pods", icon: Users },
  { href: "/search", label: "Search", icon: Search },
  { href: "/notifications", label: "Notifications", icon: Bell },
];

export function AppHeader() {
  const { username, logout, isAuthenticated } = useAuth();
  const pathname = usePathname();
  const router = useRouter();
  const { count: unread } = useUnreadNotifications();

  if (!isAuthenticated || !username) return null;

  const initials = username.slice(0, 2).toUpperCase();

  return (
    <header className="sticky top-0 z-40 border-b border-border/60 bg-background/90 backdrop-blur-md">
      <div className="mx-auto flex h-14 max-w-6xl items-center gap-4 px-4 sm:px-6">
        <Link href="/home" className="mr-2 text-lg">
          <BrandLogo
            size={28}
            wordmarkClassName="hidden text-lg sm:inline"
          />
        </Link>

        <nav className="hidden flex-1 items-center gap-1 md:flex">
          {links.map(({ href, label, icon: Icon }) => {
            const isNotifications = href === "/notifications";
            return (
              <Link
                key={href}
                href={href}
                className={cn(
                  "inline-flex items-center gap-1.5 rounded-md px-3 py-1.5 text-sm transition-colors",
                  pathname.startsWith(href)
                    ? "bg-secondary text-foreground"
                    : "text-muted-foreground hover:bg-muted hover:text-foreground"
                )}
              >
                <span className="relative">
                  <Icon className="size-4" />
                  {isNotifications && unread > 0 ? (
                    <span
                      aria-hidden
                      className="absolute -right-0.5 -top-0.5 size-2 rounded-full bg-primary ring-2 ring-background"
                    />
                  ) : null}
                </span>
                {label}
                {isNotifications && unread > 0 ? (
                  <span className="inline-flex min-w-5 items-center justify-center rounded-full bg-primary px-1.5 py-0.5 text-[10px] font-semibold text-primary-foreground">
                    {unread > 9 ? "9+" : unread}
                  </span>
                ) : null}
              </Link>
            );
          })}
        </nav>

        <div className="ml-auto flex items-center gap-2">
          <ThemeToggle />
          <DropdownMenu>
            <DropdownMenuTrigger
              className={cn(
                "inline-flex items-center gap-2 rounded-lg px-2 py-1.5 text-sm outline-none",
                "hover:bg-muted focus-visible:ring-3 focus-visible:ring-ring/50"
              )}
            >
              <Avatar className="size-7">
                <AvatarFallback className="text-xs">{initials}</AvatarFallback>
              </Avatar>
              <span className="hidden sm:inline">@{username}</span>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-48">
              <DropdownMenuItem
                onClick={() => router.push(`/u/${username}`)}
              >
                Profile
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => router.push("/settings")}>
                <Settings className="size-4" />
                Settings
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem
                variant="destructive"
                onClick={async () => {
                  await logout();
                  router.replace("/login");
                }}
              >
                <LogOut className="size-4" />
                Log out
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>
    </header>
  );
}
