import Image from "next/image";
import { cn } from "@/lib/utils";

type BrandLogoProps = {
  className?: string;
  /** Icon size in rem-ish pixels; default 32 */
  size?: number;
  showWordmark?: boolean;
  wordmarkClassName?: string;
};

/** Taggy mark: hero icon + optional wordmark. */
export function BrandLogo({
  className,
  size = 32,
  showWordmark = true,
  wordmarkClassName,
}: BrandLogoProps) {
  return (
    <span className={cn("inline-flex items-center gap-2", className)}>
      <Image
        src="/images/hero-icon.jpg"
        alt=""
        width={size}
        height={size}
        className="shrink-0 rounded-md object-cover ring-1 ring-border/40"
        aria-hidden
      />
      {showWordmark ? (
        <span className={cn("font-serif font-medium", wordmarkClassName)}>
          Taggy
        </span>
      ) : null}
    </span>
  );
}
