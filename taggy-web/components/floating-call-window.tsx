"use client";

import {
  useCallback,
  useEffect,
  useRef,
  useState,
  type PointerEvent as ReactPointerEvent,
} from "react";
import Link from "next/link";
import { GripHorizontal, Mic, MicOff, PhoneOff } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

const PANEL_W = 300;
const PANEL_H = 120;
const MARGIN = 12;

type FloatingCallWindowProps = {
  title: string;
  hostUsername?: string;
  micOn: boolean;
  busy: boolean;
  onToggleMic: () => void;
  onLeave: () => void;
  onEnd?: () => void;
  returnHref?: string | null;
};

function clamp(n: number, min: number, max: number) {
  return Math.min(max, Math.max(min, n));
}

function defaultPosition() {
  if (typeof window === "undefined") return { x: MARGIN, y: MARGIN };
  return {
    x: Math.max(MARGIN, window.innerWidth - PANEL_W - MARGIN - 8),
    y: Math.max(MARGIN, window.innerHeight - PANEL_H - MARGIN - 8),
  };
}

export function FloatingCallWindow({
  title,
  hostUsername,
  micOn,
  busy,
  onToggleMic,
  onLeave,
  onEnd,
  returnHref,
}: FloatingCallWindowProps) {
  const [pos, setPos] = useState(defaultPosition);
  const [dragging, setDragging] = useState(false);
  const dragRef = useRef<{
    startX: number;
    startY: number;
    originX: number;
    originY: number;
  } | null>(null);
  const panelRef = useRef<HTMLDivElement | null>(null);

  const keepInViewport = useCallback(() => {
    setPos((p) => {
      const w = panelRef.current?.offsetWidth ?? PANEL_W;
      const h = panelRef.current?.offsetHeight ?? PANEL_H;
      return {
        x: clamp(p.x, MARGIN, window.innerWidth - w - MARGIN),
        y: clamp(p.y, MARGIN, window.innerHeight - h - MARGIN),
      };
    });
  }, []);

  useEffect(() => {
    keepInViewport();
    window.addEventListener("resize", keepInViewport);
    return () => window.removeEventListener("resize", keepInViewport);
  }, [keepInViewport]);

  function onPointerDown(e: ReactPointerEvent<HTMLDivElement>) {
    if (e.button !== 0) return;
    e.currentTarget.setPointerCapture(e.pointerId);
    dragRef.current = {
      startX: e.clientX,
      startY: e.clientY,
      originX: pos.x,
      originY: pos.y,
    };
    setDragging(true);
  }

  function onPointerMove(e: ReactPointerEvent<HTMLDivElement>) {
    if (!dragRef.current) return;
    const dx = e.clientX - dragRef.current.startX;
    const dy = e.clientY - dragRef.current.startY;
    const w = panelRef.current?.offsetWidth ?? PANEL_W;
    const h = panelRef.current?.offsetHeight ?? PANEL_H;
    setPos({
      x: clamp(
        dragRef.current.originX + dx,
        MARGIN,
        window.innerWidth - w - MARGIN
      ),
      y: clamp(
        dragRef.current.originY + dy,
        MARGIN,
        window.innerHeight - h - MARGIN
      ),
    });
  }

  function onPointerUp(e: ReactPointerEvent<HTMLDivElement>) {
    if (!dragRef.current) return;
    try {
      e.currentTarget.releasePointerCapture(e.pointerId);
    } catch {
      /* already released */
    }
    dragRef.current = null;
    setDragging(false);
  }

  return (
    <div
      ref={panelRef}
      role="dialog"
      aria-label={`Call: ${title}`}
      className={cn(
        "fixed z-[60] flex w-[min(18.5rem,calc(100vw-1.5rem))] flex-col overflow-hidden rounded-2xl border border-border/80 bg-card/95 shadow-2xl backdrop-blur-md",
        dragging && "cursor-grabbing select-none"
      )}
      style={{ left: pos.x, top: pos.y }}
    >
      <div
        className={cn(
          "flex cursor-grab items-center gap-2 border-b border-border/70 bg-muted/40 px-3 py-2 active:cursor-grabbing",
          dragging && "cursor-grabbing"
        )}
        onPointerDown={onPointerDown}
        onPointerMove={onPointerMove}
        onPointerUp={onPointerUp}
        onPointerCancel={onPointerUp}
      >
        <GripHorizontal className="size-4 shrink-0 text-foreground/50" />
        <div className="min-w-0 flex-1">
          <p className="truncate text-sm font-medium">{title}</p>
          <p className="truncate text-[11px] text-foreground/60">
            Live call
            {hostUsername ? ` · @${hostUsername}` : ""}
          </p>
        </div>
        <span className="inline-flex items-center gap-1 rounded-full bg-primary/15 px-2 py-0.5 text-[10px] font-medium text-primary">
          <span className="size-1.5 animate-pulse rounded-full bg-primary" />
          Live
        </span>
      </div>

      <div className="flex flex-wrap items-center gap-2 px-3 py-2.5">
        <Button
          type="button"
          size="sm"
          variant="outline"
          disabled={busy}
          onClick={onToggleMic}
          className="gap-1.5"
        >
          {micOn ? <Mic className="size-3.5" /> : <MicOff className="size-3.5" />}
          {micOn ? "Mute" : "Unmute"}
        </Button>
        <Button
          type="button"
          size="sm"
          variant="destructive"
          disabled={busy}
          onClick={onLeave}
          className="gap-1.5"
        >
          <PhoneOff className="size-3.5" />
          Leave
        </Button>
        {onEnd ? (
          <Button
            type="button"
            size="sm"
            variant="ghost"
            disabled={busy}
            onClick={onEnd}
          >
            End
          </Button>
        ) : null}
        {returnHref ? (
          <Link
            href={returnHref}
            className="ml-auto text-xs font-medium text-primary hover:underline"
          >
            Open room
          </Link>
        ) : null}
      </div>
    </div>
  );
}
