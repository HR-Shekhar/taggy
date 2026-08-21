import Image from "next/image";
import Link from "next/link";
import { ArrowRight } from "lucide-react";
import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { Reveal } from "@/components/marketing/reveal";

export function Hero() {
  return (
    <section className="relative flex min-h-dvh flex-col overflow-hidden border-b border-border">
      <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_top_right,var(--color-primary)_0%,transparent_50%)] opacity-10" />
      <div className="relative mx-auto flex w-full max-w-7xl flex-1 flex-col px-6 pb-16 pt-16 sm:px-8 sm:pb-20 lg:flex-row lg:justify-between lg:gap-10 lg:px-10 lg:pb-24 lg:pt-20">
        <Reveal className="max-w-xl space-y-5 lg:w-[42%] lg:shrink-0">
          <div className="inline-flex items-center rounded-full border border-primary/30 bg-primary/10 px-3 py-1 text-sm font-medium text-primary">
            <span className="mr-2 flex h-2 w-2 rounded-full bg-primary" />
            Now in public beta
          </div>
          <h1 className="font-serif text-4xl leading-[1.12] tracking-tight sm:text-5xl lg:text-[3.25rem]">
            Learn with structure.
            <span className="mt-1.5 block text-primary">
              Stay accountable together.
            </span>
          </h1>
          <p className="max-w-md text-base text-foreground/80 sm:text-lg">
            Roadmaps, small pods, and community chat — so you finish what you
            start, not alone.
          </p>
          <div className="flex flex-wrap gap-3 pt-1">
            <Link
              href="/register"
              className={cn(
                buttonVariants({ size: "lg" }),
                "gap-2 rounded-full px-8"
              )}
            >
              Start learning free
              <ArrowRight className="size-4" />
            </Link>
            <Link
              href="/login"
              className={cn(
                buttonVariants({ size: "lg", variant: "outline" }),
                "rounded-full px-8"
              )}
            >
              I have an account
            </Link>
          </div>
        </Reveal>

        <Reveal
          delay={200}
          className="mt-8 flex w-full items-center lg:mt-0 lg:min-h-0 lg:w-[52%] lg:self-stretch"
        >
          <div className="relative mx-auto w-full max-w-xl">
            <div className="absolute -inset-3 rounded-2xl bg-gradient-to-br from-primary/15 via-transparent to-black/35 blur-2xl" />
            <div
              className="relative overflow-hidden rounded-lg bg-[#1c1c1e]
                shadow-[0_0_0_1px_rgba(255,255,255,0.08),0_1px_2px_rgba(0,0,0,0.35),0_10px_24px_rgba(0,0,0,0.4),0_24px_48px_-12px_rgba(0,0,0,0.5)]"
            >
              {/* Compact macOS title bar */}
              <div className="relative flex h-6 items-center bg-gradient-to-b from-[#3a3a3c] to-[#2c2c2e] px-2.5">
                <div className="flex items-center gap-1.5">
                  <span className="size-2 rounded-full bg-[#ff5f57]" />
                  <span className="size-2 rounded-full bg-[#febc2e]" />
                  <span className="size-2 rounded-full bg-[#28c840]" />
                </div>
                <span className="pointer-events-none absolute inset-x-0 text-center text-[9px] font-medium tracking-wide text-white/45">
                  taggy/community
                </span>
              </div>
              <div className="border-t border-black/40">
                <Image
                  src="/images/landing-card.png"
                  alt="Taggy community workspace with channels, chat, and audio rooms"
                  width={1600}
                  height={1000}
                  priority
                  className="h-auto w-full object-cover object-top"
                  sizes="(max-width: 1024px) 100vw, 36rem"
                />
              </div>
            </div>
          </div>
        </Reveal>
      </div>
    </section>
  );
}
