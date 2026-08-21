"use client";

import { FormEvent, ReactNode, useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { EmptyArtGeneric } from "@/components/empty-art";
import { CheckCircle2, Loader2, X } from "lucide-react";
import { cn } from "@/lib/utils";
import { toastError } from "@/lib/toast";

export function ErrorBox({
  message,
  title,
}: {
  message: string | null;
  title?: string;
}) {
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    if (!message) {
      setVisible(false);
      return;
    }
    const id = requestAnimationFrame(() => setVisible(true));
    return () => cancelAnimationFrame(id);
  }, [message]);

  if (!message) return null;

  return (
    <div
      role="alert"
      className={cn(
        "rounded-xl border border-destructive/25 bg-destructive/5 px-4 py-3 text-sm transition-all duration-300 ease-out",
        visible ? "translate-y-0 opacity-100" : "-translate-y-1 opacity-0"
      )}
    >
      {title ? (
        <>
          <p className="font-medium text-foreground">{title}</p>
          <p className="mt-0.5 text-foreground/80">{message}</p>
        </>
      ) : (
        <p className="font-medium text-foreground">{message}</p>
      )}
    </div>
  );
}

export function SuccessDialog({
  open,
  title = "Done",
  description,
  confirmLabel = "Got it",
  onClose,
  onConfirm,
}: {
  open: boolean;
  title?: string;
  description: string;
  confirmLabel?: string;
  onClose: () => void;
  onConfirm?: () => void;
}) {
  if (!open) return null;

  const confirm = onConfirm ?? onClose;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <button
        type="button"
        aria-label="Dismiss"
        className="absolute inset-0 bg-background/70 backdrop-blur-sm transition-opacity"
        onClick={onClose}
      />
      <Card
        role="dialog"
        aria-modal="true"
        className="relative z-10 w-full max-w-md shadow-lg duration-200 animate-in fade-in zoom-in-95"
      >
        <CardHeader className="space-y-3">
          <div className="flex items-start justify-between gap-3">
            <span className="inline-flex size-10 items-center justify-center rounded-full bg-primary/15 text-primary">
              <CheckCircle2 className="size-5" />
            </span>
            <Button
              type="button"
              variant="ghost"
              size="icon"
              className="size-8 shrink-0"
              onClick={onClose}
            >
              <X className="size-4" />
            </Button>
          </div>
          <div className="space-y-1">
            <CardTitle className="font-serif text-xl">{title}</CardTitle>
            <CardDescription className="text-sm leading-relaxed">
              {description}
            </CardDescription>
          </div>
        </CardHeader>
        <CardFooter className="justify-end border-0 bg-transparent">
          <Button type="button" onClick={confirm}>
            {confirmLabel}
          </Button>
        </CardFooter>
      </Card>
    </div>
  );
}

export function GenerationWaitNote({
  active,
  label = "AI is drafting in the background — this can take several minutes. You can leave this page; we'll notify you when it's ready.",
}: {
  active: boolean;
  label?: string;
}) {
  if (!active) return null;
  return (
    <div className="flex items-start gap-2 rounded-xl border border-border bg-muted/50 px-3 py-2.5 text-sm text-foreground/80 transition-all duration-300">
      <Loader2 className="mt-0.5 size-4 shrink-0 animate-spin text-primary" />
      <p>{label}</p>
    </div>
  );
}

export function Loading({ label = "Loading…" }: { label?: string }) {
  return (
    <div className="flex items-center gap-2 text-sm text-foreground/80">
      <Loader2 className="size-4 animate-spin" />
      <span>{label}</span>
    </div>
  );
}

export function PageSkeleton({
  variant = "dashboard",
}: {
  variant?: "dashboard" | "list" | "detail";
}) {
  if (variant === "list") {
    return (
      <div className="space-y-4 animate-pulse">
        <div className="h-8 w-48 rounded-lg bg-muted" />
        <div className="h-4 w-72 max-w-full rounded bg-muted/80" />
        <div className="space-y-2 pt-2">
          {Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="h-14 rounded-xl border border-border bg-card" />
          ))}
        </div>
      </div>
    );
  }
  if (variant === "detail") {
    return (
      <div className="space-y-6 animate-pulse">
        <div className="h-9 w-64 rounded-lg bg-muted" />
        <div className="h-4 w-96 max-w-full rounded bg-muted/80" />
        <div className="h-48 rounded-xl border border-border bg-card" />
        <div className="space-y-3">
          {Array.from({ length: 4 }).map((_, i) => (
            <div key={i} className="h-12 rounded-lg bg-muted/60" />
          ))}
        </div>
      </div>
    );
  }
  return (
    <div className="space-y-6 animate-pulse">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
        <div className="space-y-2">
          <div className="h-9 w-56 rounded-lg bg-muted" />
          <div className="h-4 w-72 max-w-full rounded bg-muted/80" />
        </div>
        <div className="h-9 w-28 rounded-lg bg-muted" />
      </div>
      <div className="grid gap-3 sm:grid-cols-3">
        {Array.from({ length: 3 }).map((_, i) => (
          <div key={i} className="h-20 rounded-xl border border-border bg-card" />
        ))}
      </div>
      <div className="h-56 rounded-xl border border-border bg-card" />
    </div>
  );
}

export function Empty({
  children,
  title,
  description,
  action,
  art,
  className,
}: {
  children?: ReactNode;
  title?: string;
  description?: string;
  action?: ReactNode;
  art?: ReactNode;
  className?: string;
}) {
  if (title || description || action || art) {
    return (
      <div
        className={cn(
          "flex flex-col items-center rounded-xl border border-dashed border-border bg-muted/30 px-6 py-10 text-center",
          className
        )}
      >
        {art ?? <EmptyArtGeneric />}
        {title ? (
          <p className="mt-4 font-serif text-lg text-foreground">{title}</p>
        ) : null}
        {description ? (
          <p className="mt-1 max-w-sm text-sm text-foreground/80">
            {description}
          </p>
        ) : null}
        {children ? (
          <div className="mt-2 text-sm text-foreground/80">{children}</div>
        ) : null}
        {action ? <div className="mt-5">{action}</div> : null}
      </div>
    );
  }

  return (
    <p
      className={cn(
        "rounded-xl border border-dashed border-border bg-muted/30 px-4 py-8 text-center text-sm text-foreground/80",
        className
      )}
    >
      {children}
    </p>
  );
}

export function useAsyncAction() {
  const [busy, setBusy] = useState(false);

  async function run(fn: () => Promise<void>) {
    setBusy(true);
    try {
      await fn();
    } catch (e) {
      toastError(e instanceof Error ? e.message : "Something went wrong");
    } finally {
      setBusy(false);
    }
  }

  return { busy, run };
}

export function FormCard({
  title,
  description,
  children,
  onSubmit,
  submitLabel,
  busy,
  extra,
}: {
  title?: string;
  description?: string;
  children: ReactNode;
  onSubmit: (e: FormEvent) => void | Promise<void>;
  submitLabel: string;
  busy?: boolean;
  extra?: ReactNode;
}) {
  return (
    <Card>
      <form
        onSubmit={async (e) => {
          e.preventDefault();
          await onSubmit(e);
        }}
      >
        {(title || description) && (
          <CardHeader>
            {title && <CardTitle className="font-serif">{title}</CardTitle>}
            {description && <CardDescription>{description}</CardDescription>}
          </CardHeader>
        )}
        <CardContent className="mx-auto w-full max-w-[14rem] space-y-4 pt-3">
          {children}
        </CardContent>
        <CardFooter className="mx-auto w-full max-w-[14rem] flex-col gap-4 border-0 bg-transparent">
          <Button type="submit" disabled={busy} className="w-full" size="lg">
            {busy && <Loader2 className="size-4 animate-spin" />}
            {busy ? "Please wait…" : submitLabel}
          </Button>
        </CardFooter>
      </form>
      {extra ? (
        <div className="mx-auto w-full max-w-[14rem] px-0 pb-(--card-spacing)">
          {extra}
        </div>
      ) : null}
    </Card>
  );
}

export function PageHeader({
  title,
  description,
  children,
}: {
  title: string;
  description?: string;
  children?: ReactNode;
}) {
  return (
    <div className="mb-6 flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
      <div className="min-w-0 space-y-1.5">
        <h1 className="font-serif text-3xl font-medium tracking-tight text-foreground sm:text-4xl">
          {title}
        </h1>
        {description && (
          <p className="max-w-xl text-sm leading-relaxed text-foreground/75 sm:text-base">
            {description}
          </p>
        )}
      </div>
      {children ? (
        <div className="flex shrink-0 flex-wrap items-center gap-2">{children}</div>
      ) : null}
    </div>
  );
}

export function Section({
  title,
  description,
  action,
  children,
  className,
}: {
  title?: string;
  description?: string;
  action?: ReactNode;
  children: ReactNode;
  className?: string;
}) {
  return (
    <section className={cn("space-y-3", className)}>
      {(title || action) && (
        <div className="flex items-end justify-between gap-3">
          <div className="min-w-0">
            {title ? (
              <h2 className="font-serif text-xl text-foreground">{title}</h2>
            ) : null}
            {description ? (
              <p className="mt-0.5 text-sm text-foreground/75">{description}</p>
            ) : null}
          </div>
          {action}
        </div>
      )}
      {children}
    </section>
  );
}
