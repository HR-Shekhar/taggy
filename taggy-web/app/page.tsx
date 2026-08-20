"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/lib/auth";
import { Loading } from "@/components/app-ui";
import { SiteHeader } from "@/components/marketing/site-header";
import { Hero } from "@/components/marketing/hero";
import { Features, HowItWorks } from "@/components/marketing/features";
import { CtaSection, SiteFooter } from "@/components/marketing/footer";

export default function LandingPage() {
  const { ready, isAuthenticated } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (ready && isAuthenticated) router.replace("/home");
  }, [ready, isAuthenticated, router]);

  if (!ready || isAuthenticated) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <Loading />
      </div>
    );
  }

  return (
    <div className="relative flex min-h-dvh flex-col overflow-x-clip bg-transparent">
      <SiteHeader />
      <main className="relative z-10 flex-1">
        <Hero />
        <HowItWorks />
        <Features />
        <CtaSection />
      </main>
      <div className="relative z-10">
        <SiteFooter />
      </div>
    </div>
  );
}
