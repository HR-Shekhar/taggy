import { cn } from "@/lib/utils";

type ArtProps = { className?: string };

export function EmptyArtSkills({ className }: ArtProps) {
  return (
    <svg
      viewBox="0 0 120 88"
      fill="none"
      aria-hidden
      className={cn("mx-auto size-20 text-primary", className)}
    >
      <rect
        x="18"
        y="14"
        width="52"
        height="60"
        rx="8"
        className="fill-primary/15 stroke-primary/40"
        strokeWidth="1.5"
      />
      <path
        d="M30 30h28M30 42h22M30 54h16"
        className="stroke-primary/50"
        strokeWidth="2"
        strokeLinecap="round"
      />
      <circle cx="86" cy="36" r="18" className="fill-secondary stroke-primary/30" strokeWidth="1.5" />
      <path
        d="M80 36l4 4 8-10"
        className="stroke-primary"
        strokeWidth="2.5"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

export function EmptyArtPods({ className }: ArtProps) {
  return (
    <svg
      viewBox="0 0 120 88"
      fill="none"
      aria-hidden
      className={cn("mx-auto size-20 text-primary", className)}
    >
      <circle cx="44" cy="34" r="14" className="fill-primary/15 stroke-primary/40" strokeWidth="1.5" />
      <circle cx="72" cy="34" r="14" className="fill-secondary stroke-primary/30" strokeWidth="1.5" />
      <circle cx="58" cy="52" r="14" className="fill-accent/60 stroke-primary/35" strokeWidth="1.5" />
      <path
        d="M30 72c6-10 18-14 28-14s22 4 28 14"
        className="stroke-muted-foreground/40"
        strokeWidth="2"
        strokeLinecap="round"
      />
    </svg>
  );
}

export function EmptyArtChat({ className }: ArtProps) {
  return (
    <svg
      viewBox="0 0 120 88"
      fill="none"
      aria-hidden
      className={cn("mx-auto size-20 text-primary", className)}
    >
      <rect
        x="16"
        y="18"
        width="58"
        height="40"
        rx="10"
        className="fill-primary/15 stroke-primary/40"
        strokeWidth="1.5"
      />
      <path d="M28 58l8-8h-8v8z" className="fill-primary/15 stroke-primary/40" strokeWidth="1.5" />
      <rect
        x="52"
        y="34"
        width="52"
        height="34"
        rx="10"
        className="fill-card stroke-border"
        strokeWidth="1.5"
      />
      <path
        d="M64 48h28M64 58h18"
        className="stroke-muted-foreground/50"
        strokeWidth="2"
        strokeLinecap="round"
      />
    </svg>
  );
}

export function EmptyArtNotifications({ className }: ArtProps) {
  return (
    <svg
      viewBox="0 0 120 88"
      fill="none"
      aria-hidden
      className={cn("mx-auto size-20 text-primary", className)}
    >
      <path
        d="M60 18c-12 0-22 9-22 20v10l-8 12h60l-8-12V38c0-11-10-20-22-20z"
        className="fill-primary/15 stroke-primary/40"
        strokeWidth="1.5"
        strokeLinejoin="round"
      />
      <path d="M52 72a8 8 0 0016 0" className="stroke-primary/50" strokeWidth="2" strokeLinecap="round" />
      <circle cx="84" cy="28" r="10" className="fill-primary text-primary-foreground" />
    </svg>
  );
}

export function EmptyArtGeneric({ className }: ArtProps) {
  return (
    <svg
      viewBox="0 0 120 88"
      fill="none"
      aria-hidden
      className={cn("mx-auto size-20 text-primary", className)}
    >
      <rect
        x="28"
        y="16"
        width="64"
        height="56"
        rx="12"
        className="fill-muted stroke-border"
        strokeWidth="1.5"
      />
      <circle cx="60" cy="40" r="12" className="fill-primary/20 stroke-primary/40" strokeWidth="1.5" />
      <path
        d="M44 62h32"
        className="stroke-muted-foreground/40"
        strokeWidth="2"
        strokeLinecap="round"
      />
    </svg>
  );
}
