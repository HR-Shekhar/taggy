"use client";

import { useEffect } from "react";
import { usePathname, useRouter } from "next/navigation";
import { AppShell } from "@/components/app-sidebar";
import { Loading } from "@/components/app-ui";
import { OnboardingProvider } from "@/components/onboarding-tour";
import { AudioRoomProvider } from "@/components/audio-room-provider";
import { ChatConnectionProvider } from "@/components/chat-connection-provider";
import { useAuth } from "@/lib/auth";

export default function MainLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { ready, isAuthenticated } = useAuth();
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    if (ready && !isAuthenticated) {
      router.replace(`/login?next=${encodeURIComponent(pathname)}`);
    }
  }, [ready, isAuthenticated, router, pathname]);

  if (!ready) {
    return (
      <div className="flex min-h-dvh items-center justify-center">
        <Loading />
      </div>
    );
  }
  if (!isAuthenticated) return null;

  return (
    <AudioRoomProvider>
      <ChatConnectionProvider>
        <OnboardingProvider>
          <AppShell>{children}</AppShell>
        </OnboardingProvider>
      </ChatConnectionProvider>
    </AudioRoomProvider>
  );
}
