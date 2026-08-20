import Link from "next/link";
import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { Reveal } from "@/components/marketing/reveal";

export function CtaSection() {
  return (
    <section className="relative py-16">
      <Reveal className="mx-auto max-w-3xl px-4 text-center sm:px-6">
        <h2 className="font-serif text-3xl tracking-tight sm:text-4xl">
          Ready to grow with your people?
        </h2>
        <p className="mt-4 text-muted-foreground">
          Create your account, verify email, pick a skill, and find a pod today.
        </p>
        <div className="mt-8 flex flex-wrap justify-center gap-3">
          <Link href="/register" className={cn(buttonVariants({ size: "lg" }))}>
            Create free account
          </Link>
        </div>
      </Reveal>
    </section>
  );
}

export function SiteFooter() {
  return (
    <footer className="px-4 pb-8 sm:px-6">
      <div className="mx-auto max-w-6xl overflow-visible rounded-xl border border-border/60 bg-card/45 shadow-sm backdrop-blur-md dark:bg-card/40">
        <div className="relative px-4 pb-2 pt-6 text-center sm:px-6 sm:pt-8">
          <p className="overflow-visible bg-gradient-to-b from-foreground to-transparent bg-clip-text font-serif text-[clamp(5.5rem,20vw,12rem)] leading-[1.15] tracking-tight text-transparent">
            Taggy
          </p>
        </div>
        <div className="flex flex-col gap-3 px-6 pb-5 pt-1 text-sm text-muted-foreground sm:flex-row sm:items-center sm:justify-between sm:px-10">
          <p>© {new Date().getFullYear()} Taggy</p>
          <div className="flex gap-4">
            <Link href="/login" className="hover:text-foreground">
              Log in
            </Link>
            <Link href="/register" className="hover:text-foreground">
              Register
            </Link>
          </div>
        </div>
      </div>
    </footer>
  );
}
