"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useAuth } from "@/lib/auth";

const links = [
  { href: "/home", label: "Home" },
  { href: "/skills", label: "Skills" },
  { href: "/progress", label: "Progress" },
  { href: "/pods", label: "Pods" },
  { href: "/search", label: "Search" },
  { href: "/notifications", label: "Notifications" },
];

export function TopNav() {
  const { username, logout, isAuthenticated } = useAuth();
  const pathname = usePathname();
  const router = useRouter();

  if (!isAuthenticated) return null;

  return (
    <header className="topnav">
      <Link href="/home" className="brand">
        Taggy
      </Link>
      {links.map((l) => (
        <Link
          key={l.href}
          href={l.href}
          className={pathname.startsWith(l.href) ? "active" : undefined}
        >
          {l.label}
        </Link>
      ))}
      <div className="spacer" />
      <Link href={`/u/${username}`} className="user">
        @{username}
      </Link>
      <Link href="/settings">Settings</Link>
      <button
        className="btn secondary"
        type="button"
        onClick={async () => {
          await logout();
          router.replace("/login");
        }}
      >
        Log out
      </button>
    </header>
  );
}
