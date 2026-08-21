import Link from "next/link";
import { ArrowRight } from "lucide-react";
import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { Reveal } from "@/components/marketing/reveal";

export function Hero() {
  return (
    <section className="relative min-h-[min(100dvh,52rem)] overflow-hidden border-b border-border">
      <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_top_right,var(--color-primary)_0%,transparent_50%)] opacity-20 dark:opacity-10" />
      <div className="relative mx-auto flex min-h-[min(100dvh,52rem)] max-w-7xl flex-col justify-center px-6 pb-16 pt-28 sm:px-8 sm:pb-20 lg:flex-row lg:items-center lg:justify-between lg:px-10 lg:pb-24 lg:pt-32">
        <Reveal className="max-w-2xl space-y-8 lg:w-1/2">
          <div className="inline-flex items-center rounded-full border border-primary/30 bg-primary/10 px-3 py-1 text-sm font-medium text-primary">
            <span className="flex h-2 w-2 rounded-full bg-primary mr-2"></span>
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
              className={cn(buttonVariants({ size: "lg" }), "gap-2 rounded-full px-8")}
            >
              Start learning free
              <ArrowRight className="size-4" />
            </Link>
            <Link
              href="/login"
              className={cn(buttonVariants({ size: "lg", variant: "outline" }), "rounded-full px-8")}
            >
              I have an account
            </Link>
          </div>
        </Reveal>
        
        <Reveal delay={200} className="mt-16 hidden lg:block lg:w-1/2 lg:mt-0">
          <div className="relative mx-auto w-full max-w-lg">
            <div className="absolute -inset-1 rounded-2xl bg-gradient-to-tr from-primary/20 to-secondary/20 blur-2xl" />
            <div className="relative rounded-2xl border border-border/50 bg-card/80 p-2 shadow-2xl backdrop-blur-sm">
              <div className="rounded-xl border border-border/50 bg-background overflow-hidden">
                <div className="flex items-center gap-1.5 border-b border-border/50 bg-muted/30 px-4 py-3">
                  <div className="h-3 w-3 rounded-full bg-destructive/80" />
                  <div className="h-3 w-3 rounded-full bg-amber-500/80" />
                  <div className="h-3 w-3 rounded-full bg-emerald-500/80" />
                </div>
                <div className="p-6 space-y-6">
                  <div className="flex items-center justify-between">
                    <div className="space-y-1.5">
                      <div className="h-5 w-32 rounded-md bg-foreground/10" />
                      <div className="h-4 w-48 rounded-md bg-muted-foreground/20" />
                    </div>
                    <div className="h-10 w-10 rounded-full bg-primary/20" />
                  </div>
                  <div className="space-y-3">
                    <div className="h-12 w-full rounded-lg border border-border/50 bg-muted/30" />
                    <div className="h-12 w-full rounded-lg border border-border/50 bg-muted/30" />
                    <div className="h-12 w-full rounded-lg border border-primary/30 bg-primary/10" />
                  </div>
                  <div className="flex gap-3">
                    <div className="h-24 w-1/2 rounded-xl border border-border/50 bg-muted/20" />
                    <div className="h-24 w-1/2 rounded-xl border border-border/50 bg-muted/20" />
                  </div>
                </div>
              </div>
            </div>
          </div>
        </Reveal>
      </div>
    </section>
  );
}
