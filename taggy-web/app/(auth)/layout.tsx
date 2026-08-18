"use client";

import { usePathname } from "next/navigation";
import { AuthPair } from "@/components/auth-shell";

export default function AuthLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();

  // Keep login + signup mounted together so the photo can slide as a cover.
  if (pathname === "/login" || pathname === "/register") {
    return <AuthPair />;
  }

  return children;
}
