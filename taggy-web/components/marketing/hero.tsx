import Image from "next/image";
import Link from "next/link";
import { ArrowRight } from "lucide-react";
import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { Reveal } from "@/components/marketing/reveal";

export function Hero() {
  return (
    <section className="relative min-h-[min(100dvh,52rem)] overflow-hidden border-b border-border">
      <div className="relative mx-auto flex min-h-[min(100dvh,52rem)] max-w-5xl flex-col items-center justify-center gap-10 px-6 pb-16 pt-28 sm:px-8 sm:pb-20 lg:flex-row lg:items-center lg:justify-between lg:gap-12 lg:px-10 lg:pb-24 lg:pt-32">
        <Reveal className="order-2 w-full max-w-2xl space-y-6 lg:order-1 lg:flex-1">
          <p className="flex items-center gap-3 font-serif text-3xl tracking-tight text-foreground sm:text-4xl">
            <Image
              src="/images/hero-icon.jpg"
              alt=""
              width={48}
              height={48}
              className="size-10 rounded-lg object-cover ring-1 ring-border/40 sm:size-12"
              aria-hidden
            />
            Taggy
          </p>
          <h1 className="font-serif text-4xl leading-[1.1] tracking-tight sm:text-5xl lg:text-6xl">
            Learn with structure.
            <span className="mt-1 block text-primary">
              Stay accountable together.
            </span>
          </h1>
          <p className="max-w-lg text-base text-muted-foreground sm:text-lg">
            Roadmaps, small pods, and community chat — so you finish what you
            start, not alone.
          </p>
          <div className="flex flex-wrap gap-3">
            <Link
              href="/register"
              className={cn(buttonVariants({ size: "lg" }), "gap-2")}
            >
              Start learning free
              <ArrowRight className="size-4" />
            </Link>
            <Link
              href="/login"
              className={cn(buttonVariants({ size: "lg", variant: "outline" }))}
            >
              I have an account
            </Link>
          </div>
        </Reveal>

        <Reveal
          delay={80}
          className="order-1 w-full max-w-[16rem] shrink-0 sm:max-w-[18rem] lg:order-2 lg:max-w-[22rem]"
        >
          <Image
            src="/images/hero-icon.jpg"
            alt="Taggy"
            width={704}
            height={704}
            priority
            className="h-auto w-full rounded-2xl object-cover shadow-lg ring-1 ring-border/40"
          />
        </Reveal>
      </div>
    </section>
  );
}
