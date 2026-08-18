import type { ReactNode } from "react";
import { cn } from "@/lib/utils";

/** Last-N-day bar chart (SVG). */
export function MiniBars({
  values,
  className,
  height = 56,
}: {
  values: number[];
  className?: string;
  height?: number;
}) {
  const max = Math.max(...values, 1);
  const gap = 3;
  const barW = values.length > 0 ? (100 - gap * (values.length - 1)) / values.length : 0;

  return (
    <svg
      viewBox={`0 0 100 ${height}`}
      className={cn("h-14 w-full overflow-visible", className)}
      preserveAspectRatio="none"
      aria-hidden
    >
      {values.map((v, i) => {
        const h = (v / max) * (height - 4);
        const x = i * (barW + gap);
        const y = height - h;
        return (
          <rect
            key={i}
            x={x}
            y={y}
            width={barW}
            height={Math.max(h, v > 0 ? 2 : 1)}
            rx={1.5}
            className={v > 0 ? "fill-primary" : "fill-muted"}
            opacity={v > 0 ? 0.85 : 0.5}
          />
        );
      })}
    </svg>
  );
}

/** Soft area sparkline. */
export function Sparkline({
  values,
  className,
  height = 48,
}: {
  values: number[];
  className?: string;
  height?: number;
}) {
  if (values.length < 2) {
    return <div className={cn("h-12 w-full rounded-md bg-muted/50", className)} />;
  }
  const max = Math.max(...values, 1);
  const min = Math.min(...values, 0);
  const range = Math.max(max - min, 1);
  const step = 100 / (values.length - 1);
  const points = values
    .map((v, i) => {
      const x = i * step;
      const y = height - ((v - min) / range) * (height - 6) - 3;
      return `${x},${y}`;
    })
    .join(" ");
  const area = `0,${height} ${points} 100,${height}`;

  return (
    <svg
      viewBox={`0 0 100 ${height}`}
      className={cn("h-12 w-full", className)}
      preserveAspectRatio="none"
      aria-hidden
    >
      <polygon points={area} className="fill-primary/20" />
      <polyline
        points={points}
        fill="none"
        className="stroke-primary"
        strokeWidth="2"
        strokeLinejoin="round"
        strokeLinecap="round"
        vectorEffect="non-scaling-stroke"
      />
    </svg>
  );
}

/** Circular progress ring (0–1). */
export function ProgressRing({
  value,
  size = 72,
  stroke = 7,
  className,
  children,
}: {
  value: number;
  size?: number;
  stroke?: number;
  className?: string;
  children?: ReactNode;
}) {
  const r = (size - stroke) / 2;
  const c = 2 * Math.PI * r;
  const pct = Math.max(0, Math.min(1, value));
  const offset = c * (1 - pct);

  return (
    <div
      className={cn("relative inline-flex items-center justify-center", className)}
      style={{ width: size, height: size }}
    >
      <svg width={size} height={size} className="-rotate-90" aria-hidden>
        <circle
          cx={size / 2}
          cy={size / 2}
          r={r}
          fill="none"
          className="stroke-muted"
          strokeWidth={stroke}
        />
        <circle
          cx={size / 2}
          cy={size / 2}
          r={r}
          fill="none"
          className="stroke-primary"
          strokeWidth={stroke}
          strokeLinecap="round"
          strokeDasharray={c}
          strokeDashoffset={offset}
        />
      </svg>
      <div className="absolute inset-0 flex items-center justify-center text-center">
        {children}
      </div>
    </div>
  );
}

/** Horizontal stacked segments (week / month / rest). */
export function StackedBar({
  segments,
  className,
}: {
  segments: Array<{ value: number; className: string; label: string }>;
  className?: string;
}) {
  const total = Math.max(
    segments.reduce((s, seg) => s + seg.value, 0),
    1
  );

  return (
    <div className={cn("space-y-2", className)}>
      <div className="flex h-2.5 overflow-hidden rounded-full bg-muted">
        {segments.map((seg) =>
          seg.value <= 0 ? null : (
            <div
              key={seg.label}
              title={`${seg.label}: ${seg.value}`}
              className={cn("h-full", seg.className)}
              style={{ width: `${(seg.value / total) * 100}%` }}
            />
          )
        )}
      </div>
      <div className="flex flex-wrap gap-x-3 gap-y-1 text-[11px] text-muted-foreground">
        {segments.map((seg) => (
          <span key={seg.label} className="inline-flex items-center gap-1.5">
            <span className={cn("size-1.5 rounded-full", seg.className)} />
            {seg.label}
          </span>
        ))}
      </div>
    </div>
  );
}

/** 7-day activity dots (presence). */
export function ActivityDots({
  active,
  className,
}: {
  active: boolean[];
  className?: string;
}) {
  return (
    <div className={cn("flex items-center gap-1.5", className)}>
      {active.map((on, i) => (
        <span
          key={i}
          className={cn(
            "size-2 rounded-full",
            on ? "bg-primary" : "bg-muted"
          )}
        />
      ))}
    </div>
  );
}
