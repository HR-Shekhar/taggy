import Link from "next/link";
import { buttonVariants } from "@/components/ui/button";
import { ThemeToggle } from "@/components/theme-toggle";
import { cn } from "@/lib/utils";

export function SiteHeader() {
  return (
    <div className="pointer-events-none fixed inset-x-0 top-0 z-50 flex items-center justify-between p-4 sm:p-5">
      <Link
        href="/"
        className="pointer-events-auto flex items-center gap-2 font-serif text-xl font-medium"
      >
        <span className="inline-flex size-8 items-center justify-center rounded-md bg-primary text-sm font-bold text-primary-foreground">
          T
        </span>
        Taggy
      </Link>
      <div className="pointer-events-auto flex items-center gap-1 rounded-full border border-border/70 bg-background/80 p-1 shadow-sm backdrop-blur-md">
        <ThemeToggle />
        <Link
          href="/login"
          className={cn(buttonVariants({ variant: "ghost" }), "rounded-full")}
        >
          Log in
        </Link>
        <Link href="/register" className={cn(buttonVariants(), "rounded-full")}>
          Get started
        </Link>
      </div>
    </div>
  );
}
