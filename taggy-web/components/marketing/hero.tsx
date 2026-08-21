import Image from "next/image";
import Link from "next/link";
import { ArrowRight } from "lucide-react";
import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { Reveal } from "@/components/marketing/reveal";

export function Hero() {
  return (
    <section className="relative min-h-[min(100dvh,52rem)] overflow-hidden border-b border-border">
      <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_top_right,var(--color-primary)_0%,transparent_50%)] opacity-10" />
      <div className="relative mx-auto flex min-h-[min(100dvh,52rem)] max-w-7xl flex-col justify-center px-6 pb-16 pt-28 sm:px-8 sm:pb-20 lg:flex-row lg:items-center lg:justify-between lg:gap-10 lg:px-10 lg:pb-24 lg:pt-32">
        <Reveal className="max-w-2xl space-y-8 lg:w-[42%]">
          <div className="inline-flex items-center rounded-full border border-primary/30 bg-primary/10 px-3 py-1 text-sm font-medium text-primary">
            <span className="mr-2 flex h-2 w-2 rounded-full bg-primary" />
            Now in public beta
          </div>
          <h1 className="font-serif text-5xl leading-[1.1] tracking-tight sm:text-6xl lg:text-7xl">
            Learn with structure.
            <span className="mt-2 block text-primary">
              Stay accountable together.
            </span>
          </h1>
          <p className="max-w-lg text-lg text-foreground/80 sm:text-xl">
            Roadmaps, small pods, and community chat — so you finish what you
            start, not alone.
          </p>
          <div className="flex flex-wrap gap-4 pt-4">
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
          className="mt-12 w-full lg:mt-0 lg:w-[58%]"
        >
          <div className="relative mx-auto w-full max-w-3xl">
            <div className="absolute -inset-3 rounded-2xl bg-gradient-to-tr from-primary/25 to-transparent blur-2xl" />
            <div className="relative overflow-hidden rounded-2xl border border-border/50 bg-card/40 shadow-2xl ring-1 ring-white/10 backdrop-blur-sm">
              <Image
                src="/images/landing-card.png"
                alt="Taggy community workspace with channels, chat, and audio rooms"
                width={1600}
                height={1000}
                priority
                className="h-auto w-full object-cover object-top"
                sizes="(max-width: 1024px) 100vw, 58vw"
              />
            </div>
          </div>
        </Reveal>
      </div>
    </section>
  );
}
