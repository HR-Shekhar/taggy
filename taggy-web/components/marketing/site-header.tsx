import Link from "next/link";
import { buttonVariants } from "@/components/ui/button";
import { ThemeToggle } from "@/components/theme-toggle";
import { BrandLogo } from "@/components/brand-logo";
import { cn } from "@/lib/utils";

export function SiteHeader() {
  return (
    <div className="pointer-events-none fixed inset-x-0 top-0 z-50 flex items-center justify-between p-4 sm:p-5">
      <Link href="/" className="pointer-events-auto text-xl">
        <BrandLogo size={32} wordmarkClassName="text-xl" />
      </Link>
      <div className="pointer-events-auto flex items-center gap-1 rounded-full border border-border/60 bg-background/55 p-1.5 shadow-sm backdrop-blur-md">
        <ThemeToggle className="shrink-0" />
        <Link
          href="/login"
          className={cn(
            buttonVariants({ variant: "ghost", size: "sm" }),
            "rounded-full"
          )}
        >
          Log in
        </Link>
        <Link
          href="/register"
          className={cn(buttonVariants({ size: "sm" }), "rounded-full")}
        >
          Get started
        </Link>
      </div>
    </div>
  );
}
